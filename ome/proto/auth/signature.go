package authpb

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func EcdsaJwtSignature(key *ecdsa.PrivateKey, j *JWT) (string, error) {
	claimsBytes, err := json.Marshal(j.Claims)
	if err != nil {
		return "", err
	}

	headerBytes, err := json.Marshal(j.Header)
	if err != nil {
		return "", err
	}

	data := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)

	sha := sha256.New()
	sha.Write([]byte(data))
	hash := sha.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s",
		base64.RawURLEncoding.EncodeToString(r.Bytes()),
		base64.RawURLEncoding.EncodeToString(s.Bytes())), nil
}

func EcdsaJwtSignatureVerify(key *ecdsa.PublicKey, j *JWT) (bool, error) {
	claimsBytes, err := json.Marshal(j.Claims)
	if err != nil {
		return false, fmt.Errorf("could not encode claims: %s", err)
	}

	headerBytes, err := json.Marshal(j.Header)
	if err != nil {
		return false, fmt.Errorf("could not encode header: %s", err)
	}

	data := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)

	sha := sha256.New()
	sha.Write([]byte(data))
	hash := sha.Sum(nil)

	parts := strings.Split(j.Signature, ".")
	r, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, errors.New("token wrong format")
	}

	s, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false, errors.New("token wrong format")
	}

	rInt := new(big.Int)
	rInt.SetBytes(r)
	sInt := new(big.Int)
	sInt.SetBytes(s)

	return ecdsa.Verify(key, hash, rInt, sInt), nil
}
