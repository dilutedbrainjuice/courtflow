package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type API struct {
	DB       *pgxpool.Pool
	Origins  map[string]bool
	mu       sync.Mutex
	window   time.Time
	attempts int
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]string{"error": message})
}
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { JSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /api/v1/facilities", a.facilities)
	m.HandleFunc("GET /api/v1/facilities/{id}/courts", a.courts)
	m.HandleFunc("GET /api/v1/courts/{id}/availability", a.availability)
	m.HandleFunc("POST /api/v1/bookings", a.book)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if origin := r.Header.Get("Origin"); origin != "" {
			if !a.Origins[origin] {
				fail(w, 403, "origin not allowed")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(204)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		m.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (a *API) facilities(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(r.Context(), "SELECT id,name,timezone FROM facilities ORDER BY id")
	if err != nil {
		fail(w, 503, "database unavailable")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var name, tz string
		if rows.Scan(&id, &name, &tz) != nil {
			fail(w, 500, "read failed")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "timezone": tz})
	}
	if rows.Err() != nil {
		fail(w, 503, "read failed")
		return
	}
	JSON(w, 200, out)
}
func (a *API) courts(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		fail(w, 400, "invalid facility")
		return
	}
	rows, err := a.DB.Query(r.Context(), "SELECT id,name,price_sen FROM courts WHERE facility_id=$1 ORDER BY id", id)
	if err != nil {
		fail(w, 503, "database unavailable")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, price int
		var name string
		if rows.Scan(&id, &name, &price) != nil {
			fail(w, 500, "read failed")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "price_sen": price})
	}
	if rows.Err() != nil {
		fail(w, 503, "read failed")
		return
	}
	JSON(w, 200, out)
}

type court struct {
	Price, Open, Close int
	Zone               string
}

func (a *API) getCourt(ctx context.Context, id int) (court, error) {
	var c court
	err := a.DB.QueryRow(ctx, "SELECT c.price_sen,f.opens,f.closes,f.timezone FROM courts c JOIN facilities f ON f.id=c.facility_id WHERE c.id=$1", id).Scan(&c.Price, &c.Open, &c.Close, &c.Zone)
	return c, err
}
func validSlot(t, now time.Time, c court) bool {
	loc, err := time.LoadLocation(c.Zone)
	if err != nil {
		return false
	}
	v := t.In(loc)
	return t.After(now) && !t.After(now.AddDate(0, 0, 30)) && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 && v.Hour() >= c.Open && v.Hour() < c.Close
}
func (a *API) availability(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		fail(w, 400, "invalid court")
		return
	}
	c, err := a.getCourt(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		fail(w, 404, "court not found")
		return
	}
	if err != nil {
		fail(w, 503, "database unavailable")
		return
	}
	loc, err := time.LoadLocation(c.Zone)
	if err != nil {
		fail(w, 500, "invalid venue timezone")
		return
	}
	day, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get("date"), loc)
	now := time.Now()
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	if err != nil || day.Before(today) || day.After(today.AddDate(0, 0, 30)) {
		fail(w, 400, "date must be within 30 days")
		return
	}
	rows, err := a.DB.Query(r.Context(), "SELECT start_at,end_at FROM bookings WHERE court_id=$1 AND status='confirmed' AND start_at < $3 AND end_at > $2", id, day, day.AddDate(0, 0, 1))
	if err != nil {
		fail(w, 503, "database unavailable")
		return
	}
	defer rows.Close()
	var occupied [][2]time.Time
	for rows.Next() {
		var s, e time.Time
		if rows.Scan(&s, &e) != nil {
			fail(w, 500, "read failed")
			return
		}
		occupied = append(occupied, [2]time.Time{s, e})
	}
	if rows.Err() != nil {
		fail(w, 503, "read failed")
		return
	}
	slots := []map[string]any{}
	for h := c.Open; h < c.Close; h++ {
		s := day.Add(time.Duration(h) * time.Hour)
		e := s.Add(time.Hour)
		available := validSlot(s, now, c)
		for _, b := range occupied {
			if s.Before(b[1]) && e.After(b[0]) {
				available = false
			}
		}
		slots = append(slots, map[string]any{"start_at": s.UTC(), "end_at": e.UTC(), "available": available, "price_sen": c.Price})
	}
	JSON(w, 200, slots)
}
func (a *API) book(w http.ResponseWriter, r *http.Request) {
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
		fail(w, 429, "demo booking limit reached; retry later")
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		fail(w, 415, "use application/json")
		return
	}
	var input struct {
		CourtID int       `json:"court_id"`
		Start   time.Time `json:"start_at"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil || d.Decode(&struct{}{}) != io.EOF || input.CourtID < 1 {
		fail(w, 400, "invalid booking request")
		return
	}
	c, err := a.getCourt(r.Context(), input.CourtID)
	if errors.Is(err, pgx.ErrNoRows) {
		fail(w, 404, "court not found")
		return
	}
	if err != nil {
		fail(w, 503, "database unavailable")
		return
	}
	if !validSlot(input.Start, time.Now(), c) {
		fail(w, 400, "choose an hourly slot within opening hours and the next 30 days")
		return
	}
	var id string
	var amount int
	err = a.DB.QueryRow(r.Context(), "INSERT INTO bookings(court_id,start_at,end_at,amount_sen) SELECT id,$2,$3,price_sen FROM courts WHERE id=$1 RETURNING id,amount_sen", input.CourtID, input.Start, input.Start.Add(time.Hour)).Scan(&id, &amount)
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23P01" {
		fail(w, 409, "slot already booked")
		return
	}
	if err != nil {
		fail(w, 503, "booking unavailable")
		return
	}
	JSON(w, 201, map[string]any{"id": id, "court_id": input.CourtID, "start_at": input.Start.UTC(), "end_at": input.Start.Add(time.Hour).UTC(), "amount_sen": amount, "status": "confirmed", "payment": "simulated"})
}
