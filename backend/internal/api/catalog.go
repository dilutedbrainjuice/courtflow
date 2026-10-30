package api

import (
	"net/http"
	"strconv"
)

// facilities lists public venue identities and their display timezones, ordered by ID.
func (a *API) facilities(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(r.Context(), `
		SELECT id,name,timezone
		FROM facilities
		ORDER BY id
	`)
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer rows.Close()
	// Initialize an empty slice so a catalog with no rows is encoded as [], not null.
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var name, tz string
		if rows.Scan(&id, &name, &tz) != nil {
			fail(w, http.StatusInternalServerError, "read failed")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "timezone": tz})
	}
	if rows.Err() != nil {
		fail(w, http.StatusServiceUnavailable, "read failed")
		return
	}
	JSON(w, http.StatusOK, out)
}

// courts lists names and hourly prices for a positive facility ID.
// An unknown facility returns an empty list, just like a facility with no courts.
func (a *API) courts(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		fail(w, http.StatusBadRequest, "invalid facility")
		return
	}
	rows, err := a.DB.Query(r.Context(), `
		SELECT id,name,price_sen
		FROM courts
		WHERE facility_id=$1
		ORDER BY id
	`, id)
	if err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	defer rows.Close()
	// Initialize an empty slice so a catalog with no rows is encoded as [], not null.
	out := []map[string]any{}
	for rows.Next() {
		var id, price int
		var name string
		if rows.Scan(&id, &name, &price) != nil {
			fail(w, http.StatusInternalServerError, "read failed")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "price_sen": price})
	}
	if rows.Err() != nil {
		fail(w, http.StatusServiceUnavailable, "read failed")
		return
	}
	JSON(w, http.StatusOK, out)
}
