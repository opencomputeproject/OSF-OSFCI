// Extracted from base/base.go for testing crypto/rand token generation.
// This file mirrors the production functions so tests can validate them
// without requiring the full module dependency graph (viper, smtp, etc).

package main

import (
	"crypto/rand"
	"math/big"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
var simpleLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randAlphaSlashPlus(n int) string {
	b := make([]rune, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}

func randAlpha(n int) string {
	b := make([]rune, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(simpleLetters))))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		b[i] = simpleLetters[idx.Int64()]
	}
	return string(b)
}

// GenerateAccountACKLink generates account verification link
func GenerateAccountACKLink(length int) string {
	return randAlpha(length)
}

// GenerateAuthToken creates auth token for created user
func GenerateAuthToken(TokenType string, length int) string {
	return randAlphaSlashPlus(length)
}

func main() {}
