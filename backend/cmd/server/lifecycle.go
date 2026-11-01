package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

// serve joins HTTP serving and request draining before its caller closes the pool.
// The signal context triggers drain without canceling accepted work. After the
// drain deadline, cancel work and allow one second for handler cleanup; a stuck
// handler is reported so main can terminate without an unbounded pool.Close.
func serve(ctx context.Context, srv *http.Server, listener net.Listener, drain time.Duration) error {
	work, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()
	srv.BaseContext = func(net.Listener) context.Context { return work }
	handler := srv.Handler
	// The gate prevents Add racing Wait after forced shutdown closes admission.
	var active sync.WaitGroup
	var gate sync.Mutex
	closing := false
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gate.Lock()
		if closing {
			gate.Unlock()
			http.Error(w, "server stopping", http.StatusServiceUnavailable)
			return
		}
		active.Add(1)
		gate.Unlock()
		defer active.Done()
		handler.ServeHTTP(w, r)
	})
	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()
	var serveErr error
	select {
	case serveErr = <-served:
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), drain)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdownCtx)
	gate.Lock()
	closing = true
	gate.Unlock()
	// Shutdown does not forcibly close active connections on its own deadline.
	if shutdownErr != nil {
		cancelWork()
		_ = srv.Close()
	}
	if serveErr == nil {
		serveErr = <-served
	}
	done := make(chan struct{})
	go func() { active.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		return errHandlersStuck
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, shutdownErr)
}

var errHandlersStuck = errors.New("handlers did not stop after forced shutdown")
