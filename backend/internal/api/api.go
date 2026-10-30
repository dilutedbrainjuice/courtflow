package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// API owns the shared database pool, allowed browser origins and booking limiter.
// Configure DB and Origins before serving requests; do not mutate them afterward.
type API struct {
	DB       *pgxpool.Pool
	Origins  map[string]bool
	mu       sync.Mutex
	window   time.Time
	attempts int
}

// JSON writes a JSON response after committing the supplied HTTP status.
func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func fail(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]string{"error": message})
}

// Handler registers routes and applies common headers, CORS and request deadlines.
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
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
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		m.ServeHTTP(w, r.WithContext(ctx))
	})
}
