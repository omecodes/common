package service

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/zoenion/common/errors"
	authpb "github.com/zoenion/common/proto/auth"
	servicepb "github.com/zoenion/common/proto/service"
	"log"
	"sync"
)

type jwtVerifier struct {
	sync.Mutex
	registry       *servicepb.SyncedRegistry
	tokenVerifiers map[string]authpb.TokenVerifier
	jwtStore       *authpb.SyncedJwtStoreClient
	CaCert         *x509.Certificate
}

func (j *jwtVerifier) Verify(ctx context.Context, t *authpb.JWT) (authpb.JWTState, error) {
	issuer := t.Claims.Iss

	state, err := j.jwtStore.State(t.Claims.Jti)
	if err != nil {
		return state, err
	}

	verifier := j.getJwtVerifier(issuer)
	if verifier == nil {
		info := j.registry.Get(t.Claims.Iss)
		if info == nil {
			return 0, errors.Forbidden
		}

		strCert, found := info.Meta["certificate"]
		if !found {
			return 0, errors.Forbidden
		}

		issCert, err := x509.ParseCertificate([]byte(strCert))
		if err != nil {
			log.Println("could not parse issuer certificate:", err)
			return 0, errors.Forbidden
		}

		verifier = authpb.NewTokenVerifier(issCert)
		j.saveJwtVerifier(t.Claims.Iss, verifier)
	}

	state, err = verifier.Verify(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("failed to verify to token: %s", errors.Internal)
	}

	if state != authpb.JWTState_VALID {
		return 0, errors.Forbidden
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

func NewVerifier(v *Vars) (authpb.TokenVerifier, error) {
	verifier := &jwtVerifier{
		registry:       v.Registry(),
		tokenVerifiers: map[string]authpb.TokenVerifier{},
	}
	return verifier, nil
}
