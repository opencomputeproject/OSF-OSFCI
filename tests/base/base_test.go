package main

import (
	"strings"
	"testing"
)

func TestRandAlphaSlashPlusLength(t *testing.T) {
	for _, length := range []int{0, 1, 10, 20, 40, 64} {
		result := randAlphaSlashPlus(length)
		if len(result) != length {
			t.Errorf("randAlphaSlashPlus(%d) returned length %d", length, len(result))
		}
	}
}

func TestRandAlphaLength(t *testing.T) {
	for _, length := range []int{0, 1, 10, 20, 40, 64} {
		result := randAlpha(length)
		if len(result) != length {
			t.Errorf("randAlpha(%d) returned length %d", length, len(result))
		}
	}
}

func TestRandAlphaSlashPlusCharset(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"
	result := randAlphaSlashPlus(1000)
	for _, c := range result {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("randAlphaSlashPlus produced invalid character: %c", c)
		}
	}
}

func TestRandAlphaCharset(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := randAlpha(1000)
	for _, c := range result {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("randAlpha produced invalid character: %c", c)
		}
	}
	// Ensure no slash or plus characters appear (those belong to the other alphabet)
	if strings.ContainsAny(result, "+/") {
		t.Error("randAlpha should not produce + or / characters")
	}
}

func TestRandAlphaSlashPlusUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		token := randAlphaSlashPlus(32)
		if seen[token] {
			t.Fatalf("randAlphaSlashPlus produced duplicate token on iteration %d", i)
		}
		seen[token] = true
	}
}

func TestRandAlphaUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		token := randAlpha(32)
		if seen[token] {
			t.Fatalf("randAlpha produced duplicate token on iteration %d", i)
		}
		seen[token] = true
	}
}

func TestGenerateAuthTokenLength(t *testing.T) {
	token := GenerateAuthToken("mac", 40)
	if len(token) != 40 {
		t.Errorf("GenerateAuthToken returned length %d, want 40", len(token))
	}
}

func TestGenerateAccountACKLinkLength(t *testing.T) {
	link := GenerateAccountACKLink(24)
	if len(link) != 24 {
		t.Errorf("GenerateAccountACKLink returned length %d, want 24", len(link))
	}
}

func TestGenerateAccountACKLinkNoSpecialChars(t *testing.T) {
	// ACK links use randAlpha which should not contain + or /
	for i := 0; i < 100; i++ {
		link := GenerateAccountACKLink(24)
		if strings.ContainsAny(link, "+/") {
			t.Errorf("GenerateAccountACKLink should not contain + or /: got %s", link)
		}
	}
}

func TestRandAlphaConcurrentSafety(t *testing.T) {
	// Run with -race flag to detect data races:
	//   go test -race ./tests/base/
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = randAlpha(32)
				_ = randAlphaSlashPlus(32)
			}
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
