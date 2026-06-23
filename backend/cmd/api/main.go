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

	todoHandler := handler.NewTodoHandler(svc)
	healthHandler := handler.NewHealthHandler(sqlHealthChecker{db: db})

	mux := http.NewServeMux()
	todoHandler.Register(mux)
	healthHandler.Register(mux)
	chain := handler.RecoveryMiddleware(handler.LoggingMiddleware(handler.CORSMiddleware(mux)))

	addr := ":" + port
	log.Printf("backend listening on %s (db: %s)", addr, dbPath)
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
