// Package api serves the public CourtFlow demo endpoints using a shared pgx pool.
// Handlers validate requests and execute SQL directly; PostgreSQL enforces booking overlap.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// API holds the dependencies and attempt counter shared by concurrent handlers.
// The caller creates and closes DB. Set DB and Origins before serving requests,
// then leave both fields unchanged. Do not copy an API after it is first used.
type API struct {
	DB      *pgxpool.Pool
	Origins map[string]bool
	// mu protects window and attempts only; database constraints protect bookings.
	mu       sync.Mutex
	window   time.Time
	attempts int
}

// JSON sets the response content type and status, then encodes value.
// Encoding or write errors cannot change the status once headers have been sent.
func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// fail returns a public error message without exposing database or driver details.
func fail(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]string{"error": message})
}

// Handler returns the routes wrapped in origin filtering and a database-work deadline.
// Requests without Origin are allowed: this browser policy is not authentication.
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	// Liveness stays independent of database connectivity.
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	m.HandleFunc("GET /api/v1/facilities", a.facilities)
	m.HandleFunc("GET /api/v1/facilities/{id}/courts", a.courts)
	m.HandleFunc("GET /api/v1/courts/{id}/availability", a.availability)
	m.HandleFunc("POST /api/v1/bookings", a.book)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if origin := r.Header.Get("Origin"); origin != "" {
			if !a.Origins[origin] {
				fail(w, http.StatusForbidden, "origin not allowed")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Handlers pass this context to pgx. This bounds database work; transport
		// read/write timeouts are configured separately on the HTTP server.
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		m.ServeHTTP(w, r.WithContext(ctx))
	})
}
