package random

import (
	"crypto/rand"
)

var UseTestCode = false

func String(length int) string {
	if UseTestCode {
		return "abcdef123456"
	}
	return fromCharset(length, "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
}

func Code(length int) string {
	if UseTestCode {
		return "1234"
	}

	return fromCharset(length, "0123456789")
}

func fromCharset(length int, charset string) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}
