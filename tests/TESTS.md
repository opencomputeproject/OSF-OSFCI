# OSFCI Test Suite

This document tracks all test files created as part of the security hardening effort.
Each section corresponds to a PR and lists every test with what it validates.

---

## PR 1 — Crypto-secure Token Generation

**Files:** `tests/base/base.go`, `tests/base/base_test.go`

**What changed:** Replaced `math/rand` (predictable PRNG) with `crypto/rand` (OS CSPRNG) for all token, secret, and validation link generation. Removed the racy `randInit` global variable.

| Test | What it validates |
|------|-------------------|
| `TestRandAlphaSlashPlusLength` | Correct output length for various sizes |
| `TestRandAlphaLength` | Correct output length for various sizes |
| `TestRandAlphaSlashPlusCharset` | Only valid characters produced (1000-char sample) |
| `TestRandAlphaCharset` | Only alphanumeric chars, no `+/` leakage |
| `TestRandAlphaSlashPlusUniqueness` | No collisions across 10k 32-byte tokens |
| `TestRandAlphaUniqueness` | No collisions across 10k 32-byte tokens |
| `TestGenerateAuthTokenLength` | Public API returns correct length (40) |
| `TestGenerateAccountACKLinkLength` | Public API returns correct length (24) |
| `TestGenerateAccountACKLinkNoSpecialChars` | ACK links use safe alphabet only |
| `TestRandAlphaConcurrentSafety` | 10 goroutines calling simultaneously — run with `go test -race` |

**Setup example**

```cd ~/code/osfci
go mod init github.com/arunkoshy/OSF-OSFCI
go mod tidy
```
**Run:**

```bash
cd tests/base && go test -v -race .
```
