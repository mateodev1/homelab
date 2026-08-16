package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// KeySize is the AES-256 key length in bytes.
const KeySize = 32

// ErrNoKeyMaterial is returned when neither SECRETS_ENCRYPTION_KEY nor a
// fallback API key is available to derive an encryption key from.
var ErrNoKeyMaterial = errors.New("no key material available to derive secrets encryption key")

// ResolveEncryptionKey derives a stable 32-byte AES-256 key from the
// provided inputs. The raw input material is always normalised through
// SHA-256 so any non-empty string of any length can be used as key
// material.
//
// Preference order:
//  1. secretsKeyEnv (SECRETS_ENCRYPTION_KEY) — explicit, dedicated key.
//  2. apiKey (API_KEY) — backward-compatible fallback for existing
//     deployments that never set SECRETS_ENCRYPTION_KEY. derived=true is
//     returned so the caller can log a warning in non-test runtime.
//
// Returns ErrNoKeyMaterial if both inputs are empty — callers must supply
// at least one so secrets can be encrypted at rest.
func ResolveEncryptionKey(secretsKeyEnv, apiKey string) (key []byte, derived bool, err error) {
	switch {
	case secretsKeyEnv != "":
		sum := sha256.Sum256([]byte("homelab-secrets-key:" + secretsKeyEnv))
		return sum[:], false, nil
	case apiKey != "":
		sum := sha256.Sum256([]byte("homelab-secrets-key-derived-from-api-key:" + apiKey))
		return sum[:], true, nil
	default:
		return nil, false, ErrNoKeyMaterial
	}
}

// NewGCMCipher builds an AES-256-GCM AEAD from a 32-byte key. This is the
// test-safe constructor seam: tests can call it directly with a fixed key
// without going through environment variables or ResolveEncryptionKey.
func NewGCMCipher(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("store.NewGCMCipher: key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("store.NewGCMCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("store.NewGCMCipher: %w", err)
	}
	return gcm, nil
}

// EncryptSecretValue encrypts plaintext with AES-256-GCM and returns a
// base64-encoded blob of nonce||ciphertext. The plaintext is never
// returned in errors or logged.
func EncryptSecretValue(aead cipher.AEAD, plaintext string) (string, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("store.EncryptSecretValue: generate nonce: %w", err)
	}
	sealed := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptSecretValue reverses EncryptSecretValue.
func DecryptSecretValue(aead cipher.AEAD, encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("store.DecryptSecretValue: decode: %w", err)
	}
	nonceSize := aead.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("store.DecryptSecretValue: ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("store.DecryptSecretValue: %w", err)
	}
	return string(plaintext), nil
}
