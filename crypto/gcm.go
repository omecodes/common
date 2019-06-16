package crypto2

import (
	"crypto/aes"
	"crypto/cipher"
)

func GCMEncrypt(key, salt, data []byte) ([]byte, error) {
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

func GCMDecrypt(key, data []byte) ([]byte, error) {
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
