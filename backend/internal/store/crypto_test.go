package store_test

import (
	"strings"
	"testing"

	"github.com/mateo/homelab/backend/internal/store"
)

func TestResolveEncryptionKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secretsKey  string
		apiKey      string
		wantDerived bool
		wantErr     bool
	}{
		{name: "prefers dedicated key", secretsKey: "dedicated", apiKey: "fallback", wantDerived: false},
		{name: "falls back to api key", secretsKey: "", apiKey: "fallback", wantDerived: true},
		{name: "no key material", secretsKey: "", apiKey: "", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key, derived, err := store.ResolveEncryptionKey(tt.secretsKey, tt.apiKey)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(key) != store.KeySize {
				t.Fatalf("expected key length %d, got %d", store.KeySize, len(key))
			}
			if derived != tt.wantDerived {
				t.Fatalf("expected derived=%v, got %v", tt.wantDerived, derived)
			}
		})
	}
}

func TestResolveEncryptionKey_Deterministic(t *testing.T) {
	t.Parallel()

	k1, _, err := store.ResolveEncryptionKey("same-input", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	k2, _, err := store.ResolveEncryptionKey("same-input", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(k1) != string(k2) {
		t.Fatalf("expected deterministic key derivation")
	}

	k3, _, err := store.ResolveEncryptionKey("different-input", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(k1) == string(k3) {
		t.Fatalf("expected different inputs to produce different keys")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	key, _, err := store.ResolveEncryptionKey("round-trip-key", "")
	if err != nil {
		t.Fatalf("resolve key: %v", err)
	}
	aead, err := store.NewGCMCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	tests := []string{
		"",
		"simple-value",
		"value with spaces",
		`value with "quotes" and \backslashes\`,
		"multi\nline\nvalue",
	}

	for _, plaintext := range tests {
		encrypted, err := store.EncryptSecretValue(aead, plaintext)
		if err != nil {
			t.Fatalf("encrypt(%q): %v", plaintext, err)
		}
		if strings.Contains(encrypted, plaintext) && plaintext != "" {
			t.Fatalf("ciphertext must not contain plaintext")
		}
		decrypted, err := store.DecryptSecretValue(aead, encrypted)
		if err != nil {
			t.Fatalf("decrypt(%q): %v", encrypted, err)
		}
		if decrypted != plaintext {
			t.Fatalf("round trip mismatch: want %q, got %q", plaintext, decrypted)
		}
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	t.Parallel()

	key, _, _ := store.ResolveEncryptionKey("nonce-key", "")
	aead, err := store.NewGCMCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}

	a, err := store.EncryptSecretValue(aead, "same-plaintext")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	b, err := store.EncryptSecretValue(aead, "same-plaintext")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if a == b {
		t.Fatalf("expected different ciphertexts due to random nonce, got identical")
	}
}

func TestNewGCMCipher_WrongKeySize(t *testing.T) {
	t.Parallel()

	if _, err := store.NewGCMCipher([]byte("too-short")); err == nil {
		t.Fatalf("expected error for wrong key size")
	}
}
