package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// book validates a guest request and immediately confirms one simulated booking.
// The client supplies the court and start time; the database supplies price and ID.
func (a *API) book(w http.ResponseWriter, r *http.Request) {
	// All visitors to this API instance share a 120-attempt window. Count invalid
	// requests too, and reset the window on the first attempt after a minute.
	// The mutex protects only this counter, not the reservation itself.
	a.mu.Lock()
	if time.Since(a.window) > time.Minute {
		a.window = time.Now()
		a.attempts = 0
	}
	a.attempts++
	limited := a.attempts > 120
	a.mu.Unlock()
	if limited {
		w.Header().Set("Retry-After", "60")
		fail(w, http.StatusTooManyRequests, "demo booking limit reached; retry later")
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "use application/json")
		return
	}
	var input struct {
		CourtID int       `json:"court_id"`
		Start   time.Time `json:"start_at"`
	}
	// Accept one JSON object of at most 1024 bytes. Reject unknown fields
	// (including client-supplied prices) and require EOF after the first value.
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil || d.Decode(&struct{}{}) != io.EOF || input.CourtID < 1 {
		fail(w, http.StatusBadRequest, "invalid booking request")
		return
	}
	result, status, message := a.reserve(r, input.CourtID, input.Start)
	if status != http.StatusCreated {
		fail(w, status, message)
		return
	}
	JSON(w, status, result)
}

// reserve owns the complete transaction so commit/rollback and pool release happen
// before book writes any response. Cooperating writers serialize on the court row;
// PostgreSQL's exclusion constraint still protects against other database writers.
func (a *API) reserve(r *http.Request, courtID int, start time.Time) (map[string]any, int, string) {
	started := time.Now()
	tx, err := a.DB.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		a.databaseError(r, "begin", started, err)
		return nil, http.StatusServiceUnavailable, "booking unavailable"
	}
	// Request cancellation must not prevent rollback from releasing the court lock.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			a.databaseError(r, "rollback", started, err)
		}
	}()
	// Lock only the court, not the joined facility. Different courts stay independent.
	var c court
	err = tx.QueryRow(r.Context(), `
		SELECT c.price_sen,f.opens,f.closes,f.timezone
		FROM courts c JOIN facilities f ON f.id=c.facility_id
		WHERE c.id=$1
		FOR UPDATE OF c
	`, courtID).Scan(&c.Price, &c.Open, &c.Close, &c.Zone)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, http.StatusNotFound, "court not found"
	}
	if err != nil {
		a.databaseError(r, "court_lock", started, err)
		return nil, http.StatusServiceUnavailable, "booking unavailable"
	}
	if !validSlot(start, time.Now(), c) {
		return nil, http.StatusBadRequest, "choose an hourly slot within opening hours and the next 30 days"
	}
	// A separate READ COMMITTED statement sees commits made while the lock waited.
	// Combining this read with lock acquisition could use the earlier snapshot.
	var occupied bool
	err = tx.QueryRow(r.Context(), `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE court_id=$1 AND status='confirmed' AND start_at < $3 AND end_at > $2
		)
	`, courtID, start, start.Add(time.Hour)).Scan(&occupied)
	if err != nil {
		a.databaseError(r, "overlap", started, err)
		return nil, http.StatusServiceUnavailable, "booking unavailable"
	}
	if occupied {
		return nil, http.StatusConflict, "slot already booked"
	}
	var id string
	var amount int
	err = tx.QueryRow(r.Context(), `
		INSERT INTO bookings(court_id,start_at,end_at,amount_sen) SELECT id,$2,$3,price_sen
		FROM courts
		WHERE id=$1
		RETURNING id,amount_sen
	`, courtID, start, start.Add(time.Hour)).Scan(&id, &amount)
	// SQLSTATE 23P01 is the exclusion-constraint violation for an occupied interval.
	// Translate this expected contention into a conflict, not a database outage.
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23P01" {
		return nil, http.StatusConflict, "slot already booked"
	}
	if err != nil {
		a.databaseError(r, "insert_scan", started, err)
		return nil, http.StatusServiceUnavailable, "booking unavailable"
	}
	// A failed commit acknowledgment is not proof that no booking was persisted.
	if err = tx.Commit(r.Context()); err != nil {
		a.databaseError(r, "commit", started, err)
		return nil, http.StatusServiceUnavailable, "booking outcome uncertain; refresh availability"
	}
	// The INSERT has committed before this response is written. A lost response
	// can leave a confirmed booking even if the browser reports a network error.
	return map[string]any{"id": id, "court_id": courtID, "start_at": start.UTC(), "end_at": start.Add(time.Hour).UTC(), "amount_sen": amount, "status": "confirmed", "payment": "simulated"}, http.StatusCreated, ""
}
