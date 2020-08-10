package apppb

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func AuthChallenge(id string, secret string) (challenge string, nonce string) {
	nonceBytes := make([]byte, 16)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(nonceBytes)

	challenge = hex.EncodeToString(h.Sum(nil))
	nonce = hex.EncodeToString(nonceBytes)
	return
}
