package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/mateo/homelab/backend/internal/handler"
	"github.com/mateo/homelab/backend/internal/service"
	"github.com/mateo/homelab/backend/internal/store"

	// Register the modernc SQLite driver.
	_ "modernc.org/sqlite"
)

type sqlHealthChecker struct {
	db *sql.DB
}

func (c sqlHealthChecker) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func main() {
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "/data/homelab.db")
	env := envOr("ENV", "development")
	apiKey := os.Getenv("API_KEY")
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	auth0Audience := os.Getenv("AUTH0_AUDIENCE")

	secretsEncryptionKey := os.Getenv("SECRETS_ENCRYPTION_KEY")

	if env == "production" && apiKey == "" {
		log.Fatalf("API_KEY must be set when ENV=production")
	}
	if auth0Domain == "" || auth0Audience == "" {
		log.Fatalf("AUTH0_DOMAIN and AUTH0_AUDIENCE must always be set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	issuer := "https://" + auth0Domain + "/"
	jwksURL := issuer + ".well-known/jwks.json"
	validator, err := handler.NewJWTValidator(ctx, jwksURL, issuer, auth0Audience)
	if err != nil {
		log.Fatalf("handler.NewJWTValidator: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("sql.Open(%q): %v", dbPath, err)
	}
	defer func() { _ = db.Close() }()

	if err := store.Migrate(db); err != nil {
		log.Fatalf("store.Migrate: %v", err)
	}

	// Wire: store → service → handler.
	s := store.New(db)
	svc := service.NewTodoService(s)
	projectSvc := service.NewProjectService(s)

	encryptionKey, derivedFromAPIKey, err := store.ResolveEncryptionKey(secretsEncryptionKey, apiKey)
	if err != nil {
		if env == "production" {
			log.Fatalf("store.ResolveEncryptionKey: %v", err)
		}
		// Dev/test runtime with neither SECRETS_ENCRYPTION_KEY nor API_KEY set:
		// fall back to a fixed, clearly-insecure dev key so local development
		// keeps working without extra setup. Never used when ENV=production.
		log.Printf("WARNING: no SECRETS_ENCRYPTION_KEY or API_KEY set; using an " +
			"insecure fixed development encryption key for secrets at rest. This " +
			"is only acceptable outside production.")
		encryptionKey, _, err = store.ResolveEncryptionKey("insecure-development-only-key", "")
		if err != nil {
			log.Fatalf("store.ResolveEncryptionKey dev fallback: %v", err)
		}
	}
	if derivedFromAPIKey {
		// SECRETS_ENCRYPTION_KEY is unset — warn loudly so operators can set a
		// dedicated key. This is production runtime only (cmd/api/main.go is
		// never invoked from tests), so the warning never fires in test runs.
		log.Printf("WARNING: SECRETS_ENCRYPTION_KEY is not set; deriving the secrets " +
			"encryption key from API_KEY for backward compatibility. Set " +
			"SECRETS_ENCRYPTION_KEY explicitly to decouple secret encryption from " +
			"the API authentication credential.")
	}
	aead, err := store.NewGCMCipher(encryptionKey)
	if err != nil {
		log.Fatalf("store.NewGCMCipher: %v", err)
	}
	secretSvc := service.NewSecretService(s, aead)

	todoHandler := handler.NewTodoHandler(svc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	secretHandler := handler.NewSecretHandler(secretSvc, apiKey)
	healthHandler := handler.NewHealthHandler(sqlHealthChecker{db: db})

	mux := http.NewServeMux()
	todoHandler.Register(mux)
	projectHandler.Register(mux)
	secretHandler.Register(mux)
	healthHandler.Register(mux)

	requireAuth := env == "production"
	// requireAuth already encodes "prod"; in prod the Fatal above guarantees
	// apiKey != "", and in dev apiKey is ignored by AuthMiddleware. Pass it
	// through unconditionally — there's nothing left to gate.
	protected := handler.AuthMiddleware(apiKey, validator, requireAuth, mux)
	chain := handler.RecoveryMiddleware(handler.LoggingMiddleware(handler.CORSMiddleware(protected)))

	addr := ":" + port
	log.Printf("backend listening on %s (db: %s, env: %s)", addr, dbPath, env)
	if err := http.ListenAndServe(addr, chain); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

// envOr returns the environment variable value or the fallback if unset/empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
