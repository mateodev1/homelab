package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testIssuer   = "https://test-tenant.example.com/"
	testAudience = "https://api.example.com"
	testKID      = "test-key-1"
)

// newTestJWKSServer serves a JWK Set JSON document containing pub, so a real
// keyfunc.Keyfunc (fetched over HTTP, cached, refreshed in the background)
// can be exercised in tests instead of mocking keyfunc itself.
func newTestJWKSServer(t *testing.T, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()

	jwk := map[string]any{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": testKID,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
	jwks := map[string]any{"keys": []any{jwk}}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, kid, issuer, audience string, expiresIn time.Duration) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func newTestValidator(t *testing.T, jwksURL string) *JWTValidator {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	v, err := NewJWTValidator(ctx, jwksURL, testIssuer, testAudience)
	if err != nil {
		t.Fatalf("NewJWTValidator: %v", err)
	}
	return v
}

func TestJWTValidator_ValidToken_Passes(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	token := signTestToken(t, key, testKID, testIssuer, testAudience, time.Hour)

	if !v.Valid(token) {
		t.Fatal("expected valid token to pass")
	}
}

func TestJWTValidator_WrongAudience_Fails(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	token := signTestToken(t, key, testKID, testIssuer, "https://wrong-audience.example.com", time.Hour)

	if v.Valid(token) {
		t.Fatal("expected token with wrong audience to fail")
	}
}

func TestJWTValidator_WrongIssuer_Fails(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	token := signTestToken(t, key, testKID, "https://wrong-issuer.example.com/", testAudience, time.Hour)

	if v.Valid(token) {
		t.Fatal("expected token with wrong issuer to fail")
	}
}

func TestJWTValidator_InvalidSignature_Fails(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	// Signed with a key that isn't present in the JWKS (same kid, different
	// key material) — signature verification must fail.
	token := signTestToken(t, otherKey, testKID, testIssuer, testAudience, time.Hour)

	if v.Valid(token) {
		t.Fatal("expected token signed by an unregistered key to fail")
	}
}

func TestJWTValidator_ExpiredToken_Fails(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	token := signTestToken(t, key, testKID, testIssuer, testAudience, -time.Hour)

	if v.Valid(token) {
		t.Fatal("expected expired token to fail")
	}
}

func TestJWTValidator_MalformedToken_Fails(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)

	if v.Valid("not-a-jwt") {
		t.Fatal("expected malformed token to fail")
	}
	if v.Valid("") {
		t.Fatal("expected empty token to fail")
	}
}

func TestJWTValidator_NilValidator_AlwaysFails(t *testing.T) {
	var v *JWTValidator
	if v.Valid("anything") {
		t.Fatal("expected nil validator to always fail")
	}
}

func TestAuthMiddleware_ValidJWT_PassesThrough(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	token := signTestToken(t, key, testKID, testIssuer, testAudience, time.Hour)

	called := false
	h := AuthMiddleware("", v, true, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called for a valid JWT")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidJWTAndNoAPIKeyMatch_Unauthorized(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)
	badToken := signTestToken(t, key, testKID, testIssuer, "wrong-audience", time.Hour)

	h := AuthMiddleware("secret", v, true, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler should not be called")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+badToken)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_APIKeyIgnoredWhenNotProduction(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestJWKSServer(t, &key.PublicKey)
	defer srv.Close()

	v := newTestValidator(t, srv.URL)

	// Caller passes apiKey="" with requireAuth=true to simulate a prod build
	// where the static M2M credential is not configured: even a request
	// carrying what would be the correct API_KEY value must fall through to
	// JWT validation and fail, since there is no valid JWT here either.
	h := AuthMiddleware("", v, true, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler should not be called")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
