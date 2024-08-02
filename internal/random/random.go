// Package random provides common generators of random data.
package random

import (
	"crypto/rand"

	"github.com/google/uuid"
)

func String(length int) string {
	return fromCharset(length, "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
}

func Code(length int) string {
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

func UuidString() string {
	return uuid.NewString()
}
