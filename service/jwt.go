package service

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/zoenion/common/data"
	"github.com/zoenion/common/errors"
	authpb "github.com/zoenion/common/proto/auth"
	"path/filepath"
	"sync"
)

type JwtRevokedHandlerFunc func()

type jwtVerifier struct {
	sync.Mutex
	vars           *Vars
	storesMutex    sync.Mutex
	tokenVerifiers map[string]authpb.TokenVerifier
	syncedStores   map[string]*SyncedJwtStore
	CaCert         *x509.Certificate
}

func (j *jwtVerifier) Verify(ctx context.Context, t *authpb.JWT) (authpb.JWTState, error) {
	issuer := t.Claims.Iss

	verifier := j.getJwtVerifier(issuer)
	if verifier == nil {
		issCert, err := j.vars.Registry().Certificate(issuer)
		if err != nil {
			return 0, errors.Forbidden
		}
		verifier = authpb.NewTokenVerifier(issCert)
		j.saveJwtVerifier(t.Claims.Iss, verifier)
	}

	state, err := verifier.Verify(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("failed to verify to token: %s", errors.Internal)
	}
	if state != authpb.JWTState_VALID {
		return 0, errors.Forbidden
	}

	if t.Claims.Store != "" {
		jwtStore := j.getStore(t.Claims.Store)
		if jwtStore == nil {
			ci, err := j.vars.Registry().ConnectionInfo(t.Claims.Store, "gRPC")
			if err != nil {
				return 0, errors.Forbidden
			}
			dictStore, err := data.NewDictDB(filepath.Join(j.vars.dir, ""))
			if err != nil {
				return 0, errors.Internal
			}
			jwtStore = NewSyncJwtStore(ci.Address, ci.Certificate, dictStore)
			j.saveStore(t.Claims.Store, jwtStore)
		}

		state, err = jwtStore.State(t.Claims.Jti)
		if err != nil {
			return state, err
		}
	}

	ctx = context.WithValue(ctx, User, t.Claims.Sub)
	return authpb.JWTState_VALID, nil
}

func (j *jwtVerifier) saveJwtVerifier(name string, v authpb.TokenVerifier) {
	j.Lock()
	defer j.Unlock()
	j.tokenVerifiers[name] = v
}

func (j *jwtVerifier) getJwtVerifier(name string) authpb.TokenVerifier {
	j.Lock()
	defer j.Unlock()
	return j.tokenVerifiers[name]
}

func (j *jwtVerifier) getStore(name string) *SyncedJwtStore {
	j.Lock()
	defer j.Unlock()
	return j.syncedStores[name]
}

func (j *jwtVerifier) saveStore(name string, s *SyncedJwtStore) {
	j.Lock()
	defer j.Unlock()
	j.syncedStores[name] = s
}

func NewVerifier(v *Vars) (authpb.TokenVerifier, error) {
	verifier := &jwtVerifier{
		vars:           v,
		tokenVerifiers: map[string]authpb.TokenVerifier{},
		syncedStores:   map[string]*SyncedJwtStore{},
	}
	return verifier, nil
}
