package doauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenerateVerifier creates a cryptographically strong random string to be used as a PKCE code verifier.
// The verifier must be between 43 and 128 characters long.
func GenerateVerifier() (string, error) {
	// 32 bytes of randomness results in a 43-character base64 string.
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// GenerateChallenge creates a PKCE code challenge from a verifier using the S256 method.
// S256(verifier) = BASE64URL-ENCODE(SHA256(ASCII(verifier)))
func GenerateChallenge(verifier string) string {
	sha := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sha[:])
}
