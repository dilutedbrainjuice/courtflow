package api

import (
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
	c, err := a.getCourt(r.Context(), input.CourtID)
	if errors.Is(err, pgx.ErrNoRows) {
		fail(w, http.StatusNotFound, "court not found")
		return
	}
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	if !validSlot(input.Start, time.Now(), c) {
		fail(w, http.StatusBadRequest, "choose an hourly slot within opening hours and the next 30 days")
		return
	}
	var id string
	var amount int
	// This single statement reads the current price and inserts the reservation
	// atomically. The exclusion constraint arbitrates concurrent overlapping
	// requests; checking availability before inserting would leave a race.
	err = a.DB.QueryRow(r.Context(), `
		INSERT INTO bookings(court_id,start_at,end_at,amount_sen) SELECT id,$2,$3,price_sen
		FROM courts
		WHERE id=$1
		RETURNING id,amount_sen
	`, input.CourtID, input.Start, input.Start.Add(time.Hour)).Scan(&id, &amount)
	// SQLSTATE 23P01 is the exclusion-constraint violation for an occupied interval.
	// Translate this expected contention into a conflict, not a database outage.
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23P01" {
		fail(w, 409, "slot already booked")
		return
	}
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "booking unavailable")
		return
	}
	// The INSERT has committed before this response is written. A lost response
	// can leave a confirmed booking even if the browser reports a network error.
	JSON(w, http.StatusCreated, map[string]any{"id": id, "court_id": input.CourtID, "start_at": input.Start.UTC(), "end_at": input.Start.Add(time.Hour).UTC(), "amount_sen": amount, "status": "confirmed", "payment": "simulated"})
}
