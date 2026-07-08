package handler

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates Auth0-issued access tokens: RS256 signature against a
// remote JWKS (cached and rotated in the background by keyfunc), plus issuer
// and audience claims.
type JWTValidator struct {
	keyfunc  keyfunc.Keyfunc
	issuer   string
	audience string
}

// NewJWTValidator builds a JWTValidator backed by the JWK Set at jwksURL. The
// returned keyfunc keeps a background refresh goroutine alive for the
// lifetime of ctx.
func NewJWTValidator(ctx context.Context, jwksURL, issuer, audience string) (*JWTValidator, error) {
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("keyfunc.NewDefaultCtx(%q): %w", jwksURL, err)
	}
	return &JWTValidator{keyfunc: k, issuer: issuer, audience: audience}, nil
}

// Valid reports whether tokenString is a well-formed JWT, signed by a key
// present in the JWKS, not expired, and carrying the expected issuer and
// audience.
func (v *JWTValidator) Valid(tokenString string) bool {
	if v == nil || tokenString == "" {
		return false
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)

	token, err := parser.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, v.keyfunc.Keyfunc)
	if err != nil {
		return false
	}
	return token.Valid
}
