package crypto2

import (
	"crypto/rand"
	"io"
)

var verificationCodeDigits = []rune("0123456789abcdefghijklmnopqrstuv")


func GenerateVerificationCode(max int) (string, error)  {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = byte(verificationCodeDigits[int(b[i])%len(verificationCodeDigits)])
	}
	return string(b), nil
}
