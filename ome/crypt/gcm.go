package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

func AESGCMEncryptWithSalt(key, salt, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	result := gcm.Seal(nil, salt[:12], data, nil)
	return append(salt[:12], result...), nil
}

func AESGCMEncrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	salt := make([]byte, 12)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, err
	}

	result := gcm.Seal(nil, salt, data, nil)
	return append(salt, result...), nil
}

func AESGCMDecrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, data[:12], data[12:], nil)
}

func AESGCMDecryptWithNonce(key, nonce, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, data, nil)
}
