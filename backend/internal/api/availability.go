package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// availability offers one-hour sessions for a court on a venue-local calendar date.
// Results are a snapshot: only a later booking INSERT can reserve a session.
func (a *API) availability(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		fail(w, http.StatusBadRequest, "invalid court")
		return
	}
	c, err := a.getCourt(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		fail(w, http.StatusNotFound, "court not found")
		return
	}
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	loc, err := time.LoadLocation(c.Zone)
	if err != nil {
		fail(w, http.StatusInternalServerError, "invalid venue timezone")
		return
	}
	// Interpret the requested date at venue midnight, rather than UTC midnight.
	// Date validation admits day 30; validSlot still checks each start against
	// the exact booking horizon, so some sessions on that date may be unavailable.
	day, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get("date"), loc)
	now := time.Now()
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	if err != nil || day.Before(today) || day.After(today.AddDate(0, 0, 30)) {
		fail(w, http.StatusBadRequest, "date must be within 30 days")
		return
	}
	// Read confirmed intervals intersecting [day, next local midnight).
	// The strict inequalities exclude bookings that only touch a day boundary.
	rows, err := a.DB.Query(r.Context(), `
		SELECT start_at,end_at
		FROM bookings
		WHERE court_id=$1 AND status='confirmed' AND start_at < $3 AND end_at > $2
	`, id, day, day.AddDate(0, 0, 1))
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer rows.Close()
	var occupied [][2]time.Time
	for rows.Next() {
		var s, e time.Time
		if rows.Scan(&s, &e) != nil {
			fail(w, http.StatusInternalServerError, "read failed")
			return
		}
		occupied = append(occupied, [2]time.Time{s, e})
	}
	if rows.Err() != nil {
		fail(w, http.StatusServiceUnavailable, "read failed")
		return
	}
	slots := []map[string]any{}
	for h := c.Open; h < c.Close; h++ {
		s := day.Add(time.Duration(h) * time.Hour)
		e := s.Add(time.Hour)
		available := validSlot(s, now, c)
		// [start, end) intervals overlap only when each starts before the other ends.
		// A session ending exactly when another starts remains bookable.
		for _, b := range occupied {
			if s.Before(b[1]) && e.After(b[0]) {
				available = false
			}
		}
		slots = append(slots, map[string]any{"start_at": s.UTC(), "end_at": e.UTC(), "available": available, "price_sen": c.Price})
	}
	JSON(w, http.StatusOK, slots)
}
