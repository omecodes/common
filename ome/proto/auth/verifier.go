package authpb

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type TokenVerifier interface {
	Verify(ctx context.Context, t *JWT) (JWTState, error)
}

type tokenVerifier struct {
	sync.Mutex
	key *ecdsa.PublicKey
}

func (v *tokenVerifier) verifyToken(t *JWT) (JWTState, error) {
	verified, err := EcdsaJwtSignatureVerify(v.key, t)
	if err != nil {
		return 0, err
	}

	if !verified {
		return JWTState_NOT_SIGNED, nil
	}

	if t.Claims.Exp != -1 && t.Claims.Exp <= time.Now().Unix() {
		return JWTState_EXPIRED, nil
	}

	if t.Claims.Nbf != -1 && t.Claims.Nbf > time.Now().Unix() {
		return JWTState_NOT_EFFECTIVE, nil
	}

	return JWTState_VALID, nil
}

func (v *tokenVerifier) Verify(ctx context.Context, t *JWT) (JWTState, error) {
	if t == nil {
		return JWTState_NOT_VALID, errors.New("forbidden")
	}

	state, err := v.verifyToken(t)
	if err != nil {
		return JWTState_NOT_VALID, errors.New("forbidden")
	}
	return state, nil
}

func NewTokenVerifier(key *ecdsa.PublicKey) *tokenVerifier {
	return &tokenVerifier{
		key: key,
	}
}

func ParseJWT(jwt string) (*JWT, error) {
	if jwt == "" {
		return nil, nil
	}

	jwt = strings.Replace(jwt, "Bearer ", "", 1)

	malformed := errors.New("malformed token")
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("missing parts")
	}

	var t JWT
	t.Header = new(JWTHeader)
	t.Claims = new(Claims)

	headerBytes, _ := base64.RawURLEncoding.DecodeString(parts[0])
	if headerBytes == nil {
		return nil, malformed
	}

	claimsBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	if claimsBytes == nil {
		return nil, malformed
	}

	signatureBytes, _ := base64.RawURLEncoding.DecodeString(parts[2])
	if signatureBytes == nil {
		return nil, malformed
	}

	err := json.Unmarshal(headerBytes, t.Header)
	if err != nil {
		return nil, malformed
	}

	err = json.Unmarshal(claimsBytes, t.Claims)
	if err != nil {
		return nil, malformed
	}

	err = json.Unmarshal(signatureBytes, &t.Signature)
	if err != nil {
		return nil, malformed
	}

	return &t, nil
}

type StringTokenVerifier struct {
	verifier TokenVerifier
}

func (stv *StringTokenVerifier) Verify(ctx context.Context, jwt string) (context.Context, error) {
	t, err := ParseJWT(jwt)
	if err != nil {
		return ctx, err
	}
	_, err = stv.verifier.Verify(ctx, t)
	return ctx, err
}

func NewStringTokenVerifier(tv TokenVerifier) *StringTokenVerifier {
	return &StringTokenVerifier{
		verifier: tv,
	}
}

func String(jwt *JWT) (string, error) {
	headerBytes, err := json.Marshal(jwt.Header)
	if err != nil {
		return "", err
	}

	claimsBytes, err := json.Marshal(jwt.Claims)
	if err != nil {
		return "", err
	}

	signatureBytes, err := json.Marshal(jwt.Signature)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s.%s",
		base64.RawURLEncoding.EncodeToString(headerBytes),
		base64.RawURLEncoding.EncodeToString(claimsBytes),
		base64.RawURLEncoding.EncodeToString(signatureBytes),
	), nil
}
