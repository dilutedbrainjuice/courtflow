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

func (a *API) book(w http.ResponseWriter, r *http.Request) {
	// This process-wide abuse cap counts attempts, including invalid requests.
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
	// Read the price in the INSERT so the browser cannot choose the amount.
	// PostgreSQL arbitrates concurrent overlap; a preflight read cannot guarantee it.
	err = a.DB.QueryRow(r.Context(), `
		INSERT INTO bookings(court_id,start_at,end_at,amount_sen) SELECT id,$2,$3,price_sen
		FROM courts
		WHERE id=$1
		RETURNING id,amount_sen
	`, input.CourtID, input.Start, input.Start.Add(time.Hour)).Scan(&id, &amount)
	// Exclusion violations mean another booking already occupies this interval.
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23P01" {
		fail(w, 409, "slot already booked")
		return
	}
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "booking unavailable")
		return
	}
	JSON(w, http.StatusCreated, map[string]any{"id": id, "court_id": input.CourtID, "start_at": input.Start.UTC(), "end_at": input.Start.Add(time.Hour).UTC(), "amount_sen": amount, "status": "confirmed", "payment": "simulated"})
}
