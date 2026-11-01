package api

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type requestIDKey struct{}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// databaseError records safe error classes and cumulative pool counters, never raw
// driver messages (which can contain SQL values or connection details). elapsed_ms
// is time since reserve began, not the duration of the named phase alone.
func (a *API) databaseError(r *http.Request, phase string, started time.Time, err error) {
	code := ""
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		code = pe.Code
	}
	class := "database"
	if errors.Is(err, context.DeadlineExceeded) {
		class = "deadline"
	}
	if errors.Is(err, context.Canceled) {
		class = "canceled"
	}
	stats := a.DB.Stat()
	slog.Error("database operation failed", "request_id", requestID(r.Context()), "phase", phase,
		"elapsed_ms", time.Since(started).Milliseconds(), "class", class, "sqlstate", code,
		"pool_total", stats.TotalConns(), "pool_acquired", stats.AcquiredConns(),
		"pool_idle", stats.IdleConns(), "pool_acquire_count", stats.AcquireCount(),
		"pool_acquire_ms", stats.AcquireDuration().Milliseconds(),
		"pool_empty_acquire_count", stats.EmptyAcquireCount(), "pool_canceled_acquire_count", stats.CanceledAcquireCount())
}

// responseStatus observes ordinary JSON responses; Unwrap preserves access through
// http.ResponseController without claiming streaming/hijacking interfaces.
type responseStatus struct {
	http.ResponseWriter
	status int
}

func (w *responseStatus) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *responseStatus) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}
func (w *responseStatus) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// observe generates correlation IDs independently of untrusted client headers.
// Route patterns exclude query strings and concrete path values from logs.
func observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		recorder := &responseStatus{ResponseWriter: w}
		started := time.Now()
		defer func() {
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			slog.Info("request completed", "request_id", id, "method", r.Method,
				"route", r.Pattern, "status", status, "elapsed_ms", time.Since(started).Milliseconds())
		}()
		next.ServeHTTP(recorder, r)
	})
}
