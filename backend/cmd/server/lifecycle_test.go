package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeDrain(t *testing.T) {
	for _, forced := range []bool{false, true} {
		name := "completed"
		if forced {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			entered := make(chan struct{})
			release := make(chan struct{})
			finished := make(chan struct{})
			srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				close(entered)
				defer close(finished)
				select {
				case <-release:
					_, _ = io.WriteString(w, "completed")
				case <-r.Context().Done():
				}
			})}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			drain := time.Second
			if forced {
				drain = 50 * time.Millisecond
			}
			stopped := make(chan error, 1)
			go func() { stopped <- serve(ctx, srv, listener, drain) }()
			response := make(chan string, 1)
			go func() {
				client := &http.Client{Timeout: 3 * time.Second}
				res, err := client.Get("http://" + listener.Addr().String())
				if err != nil {
					response <- ""
					return
				}
				defer res.Body.Close()
				body, _ := io.ReadAll(res.Body)
				response <- string(body)
			}()
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("handler did not enter")
			}
			cancel()
			if !forced {
				select {
				case err := <-stopped:
					t.Fatalf("returned before drain: %v", err)
				default:
				}
				close(release)
			}
			select {
			case err := <-stopped:
				if forced && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("want deadline, got %v", err)
				}
				if !forced && err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("shutdown did not finish")
			}
			select {
			case <-finished:
			default:
				t.Fatal("shutdown returned before handler finished")
			}
			if body := <-response; !forced && body != "completed" {
				t.Fatalf("body=%q", body)
			}
		})
	}
}

func TestServeListenerFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	if err := serve(context.Background(), &http.Server{Handler: http.NotFoundHandler()}, listener, time.Second); err == nil {
		t.Fatal("closed listener must fail")
	}
}
