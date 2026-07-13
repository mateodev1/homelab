package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

// AuthMiddleware gates access to the backend based on requireAuth, the
// symmetric dev/prod switch the caller derives from ENV.
//
// OPTIONS preflight and requests to /api/health always pass through unchanged.
// When requireAuth is false (dev mode), every other request also passes
// through WITHOUT any credential check — dev is fully open.
// When requireAuth is true (prod mode), a valid "Authorization: Bearer <token>"
// header is required on every non-exempt request. Two credential types are
// accepted in parallel:
//
//  1. An exact match against apiKey — a static M2M client credential. Callers
//     (cmd/api/main.go) only pass a non-empty apiKey when ENV=production.
//  2. A JWT validated by validator (Auth0 access token: RS256 signature via
//     JWKS, issuer, audience, expiry) — checked for real users via the
//     frontend.
//
// The dev/prod gate is owned by requireAuth so dev is open and prod is
// authenticated symmetrically; apiKey acceptance remains env-gated by the
// caller passing a non-empty value only in production.
func AuthMiddleware(apiKey string, validator *JWTValidator, requireAuth bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		if !requireAuth {
			next.ServeHTTP(w, r)
			return
		}

		const prefix = "Bearer "
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, prefix) {
			unauthorized(w)
			return
		}
		token := strings.TrimPrefix(auth, prefix)

		if apiKey != "" && token == apiKey {
			next.ServeHTTP(w, r)
			return
		}
		if validator.Valid(token) {
			next.ServeHTTP(w, r)
			return
		}

		unauthorized(w)
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
