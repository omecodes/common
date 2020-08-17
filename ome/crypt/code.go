package crypt

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

var verificationCodeDigits = []rune("0123456789abcdefghijklmnopqrstuv")

func GenerateVerificationCode(max int) (string, error) {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "", nil
	}

	if err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = byte(verificationCodeDigits[int(b[i])%len(verificationCodeDigits)])
	}
	return string(b), nil
}

func RandomCode(max int) (string, error) {
	buff := make([]byte, max)
	_, err := rand.Read(buff)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buff), nil
}
