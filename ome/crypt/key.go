package crypt

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	// "github.com/omecodes/common/utils/jcon"
	"golang.org/x/crypto/pbkdf2"
	"hash"
)

const (
	// PBKDF2Iterations set to 10 000 which is the acceptable number of pbkdf2 iterations for year 2020
	PBKDF2Iterations = 10000
)

// Info about encrypted key parameter
type Info struct {
	Iterations   int    `json:"iterations"`
	Salt         string `json:"salt"`
	Length       int    `json:"length"`
	Hash         string `json:"hash"`
	Alg          string `json:"alg"`
	EncryptedKey string `json:"encrypted_key"`
}

// Generate generates random password of size length. Then encrypted with a key derived from password using pbkdf2.
//The derivation parameters are put in an Info object
func Generate(phrase string, length int) ([]byte, *Info, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return nil, nil, err
	}

	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	info := &Info{
		Iterations:   0,
		Salt:         "",
		Length:       0,
		Hash:         "",
		EncryptedKey: "",
	}

	info.Salt = base64.RawStdEncoding.EncodeToString(salt)
	info.Iterations = PBKDF2Iterations
	info.Length = 32
	info.Hash = "sha512"
	info.Alg = "aes-gcm-256"

	k := pbkdf2.Key([]byte(phrase), salt, PBKDF2Iterations, 32, sha512.New)

	encryptedKeyBytes, err := AESGCMEncrypt(k, bytes)
	if err != nil {
		return nil, nil, err
	}

	info.EncryptedKey = base64.RawStdEncoding.EncodeToString(encryptedKeyBytes)
	return bytes, info, nil
}

// Reveal decrypts info.Encrypted with a key built from phrase and info content.
func Reveal(phrase string, info *Info) ([]byte, error) {
	salt, err := base64.RawStdEncoding.DecodeString(info.Salt)
	if err != nil {
		return nil, err
	}

	var hf func() hash.Hash
	switch info.Hash {
	case "sha512":
		hf = sha512.New

	case "sha256":
		hf = sha256.New

	default:
		hf = sha1.New
	}

	k := pbkdf2.Key([]byte(phrase), salt, PBKDF2Iterations, 32, hf)
	encryptedKeyBytes, err := base64.RawStdEncoding.DecodeString(info.EncryptedKey)
	if err != nil {
		return nil, err
	}

	switch info.Alg {
	default:
		return AESGCMDecrypt(k, encryptedKeyBytes)
	}
}
