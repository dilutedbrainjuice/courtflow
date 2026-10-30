package api

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSlotValidation(t *testing.T) {
	c := court{Open: 8, Close: 23, Zone: "Asia/Kuala_Lumpur"}
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		offset time.Duration
		want   bool
	}{{time.Hour, true}, {time.Minute, false}, {-time.Hour, false}, {16 * time.Hour, false}, {31 * 24 * time.Hour, false}} {
		if got := validSlot(now.Add(tc.offset), now, c); got != tc.want {
			t.Errorf("offset %v got %v", tc.offset, got)
		}
	}
}
func TestIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL absent; integration not run")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host != "127.0.0.1" && cfg.ConnConfig.Host != "localhost" {
		t.Fatal("integration requires loopback database")
	}
	ctx := context.Background()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	sql, err := os.ReadFile("../../migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			t.Fatal(err)
		}
	}
	a := &API{DB: pool, Origins: map[string]bool{"http://localhost:3000": true}}
	h := a.Handler()
	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	day := time.Now().In(loc).AddDate(0, 0, 1)
	start := time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, loc)
	post := func(body string) int {
		r := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	body := func(id int, s time.Time) string {
		return fmt.Sprintf(`{"court_id":%d,"start_at":%q}`, id, s.Format(time.RFC3339))
	}
	var wg sync.WaitGroup
	codes := make(chan int, 24)
	for range 24 {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- post(body(1, start)) }()
	}
	wg.Wait()
	close(codes)
	success, conflict := 0, 0
	for code := range codes {
		switch code {
		case 201:
			success++
		case 409:
			conflict++
		default:
			t.Errorf("unexpected %d", code)
		}
	}
	if success != 1 || conflict != 23 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	var count int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM bookings WHERE court_id=1").Scan(&count); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	for _, b := range []string{body(1, start.Add(time.Hour)), body(2, start)} {
		if code := post(b); code != 201 {
			t.Errorf("adjacent/different court: %d", code)
		}
	}
	for _, b := range []string{`{"court_id":1,"start_at":"bad"}`, strings.TrimSuffix(body(3, start), "}") + `,"price_sen":1}`, body(1, start.Add(30*time.Minute)), body(1, start.AddDate(0, 0, 31)), body(1, time.Date(day.Year(), day.Month(), day.Day(), 2, 0, 0, 0, loc)), `{}`} {
		if code := post(b); code != 400 {
			t.Errorf("invalid input %s returned %d", b, code)
		}
	}
	for _, path := range []string{"/healthz", "/api/v1/facilities", "/api/v1/facilities/1/courts", "/api/v1/courts/1/availability?date=" + day.Format("2006-01-02")} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthz", nil)
	r.Header.Set("Origin", "https://untrusted.invalid")
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Error("origin accepted")
	}
	a.mu.Lock()
	a.attempts = 120
	a.window = time.Now()
	a.mu.Unlock()
	if code := post(body(3, start)); code != 429 {
		t.Errorf("rate limit %d", code)
	}
}
