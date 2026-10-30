package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // Embed venue timezone rules in minimal deployment images.

	"courtflow/internal/api"

	"github.com/jackc/pgx/v5/pgxpool"
)

// main assembles the database pool and HTTP server from environment settings.
// Schema migrations are a separate operator step; startup never changes the schema.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("invalid database configuration")
		os.Exit(1)
	}
	// Bound database usage per API process without keeping idle connections open.
	cfg.MaxConns = 8
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	// Pool creation does not prove database readiness; connections are opened as needed.
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("database setup failed")
		os.Exit(1)
	}
	defer pool.Close()
	// Origins must match browser Origin headers exactly, including scheme and port.
	origins := map[string]bool{}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o != "" {
			origins[o] = true
		}
	}
	a := &api.API{DB: pool, Origins: origins}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           a.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8192,
	}
	// On termination, stop accepting requests and give active handlers up to ten
	// seconds to finish. Use a fresh context because the signal context is canceled.
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	slog.Info("API listening", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed")
		os.Exit(1)
	}
}
