package api

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBookingLockCancellationAndIndependentCourt(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	lock, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(ctx)
	if _, err = lock.Exec(ctx, "SELECT id FROM courts WHERE id=1 FOR UPDATE"); err != nil {
		t.Fatal(err)
	}
	a := &API{DB: pool}
	handler := a.Handler()
	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	day := time.Now().In(loc).AddDate(0, 0, 1)
	start := time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, loc)
	post := func(ctx context.Context, id int) int {
		body := fmt.Sprintf(`{"court_id":%d,"start_at":%q}`, id, start.Format(time.RFC3339))
		request := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader(body)).WithContext(ctx)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}
	waiting, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan int, 1)
	go func() { result <- post(waiting, 1) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var count int
		err = pool.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock'`, pool.Config().ConnConfig.RuntimeParams["application_name"]).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("booking never waited on court lock")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if code := post(ctx, 2); code != 201 {
		t.Fatalf("independent court: %d", code)
	}
	cancel()
	select {
	case code := <-result:
		if code != 503 {
			t.Fatalf("canceled request: %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled booking did not return")
	}
	if err := lock.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings WHERE court_id=1").Scan(&count); err != nil || count != 0 {
		t.Fatalf("canceled write: rows=%d err=%v", count, err)
	}
	if code := post(ctx, 1); code != 201 {
		t.Fatalf("pool/lock did not recover: %d", code)
	}
}

func TestDirectOverlapAndRollback(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	start := time.Now().UTC().Truncate(time.Hour).Add(24 * time.Hour)
	insert := `INSERT INTO bookings(court_id,start_at,end_at,amount_sen) VALUES($1,$2,$3,2500)`
	if _, err := pool.Exec(ctx, insert, 1, start, start.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, insert, 1, start.Add(30*time.Minute), start.Add(90*time.Minute)); err == nil {
		t.Fatal("partial overlap accepted")
	}
	if _, err := pool.Exec(ctx, insert, 1, start.Add(time.Hour), start.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, insert, 2, start, start.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, insert, 1, start, start.Add(time.Hour)); err == nil {
		t.Fatal("overlap accepted")
	}
	if err := tx.Commit(ctx); err == nil {
		t.Fatal("aborted transaction committed")
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM bookings WHERE court_id=2").Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback count=%d err=%v", count, err)
	}
}
