package api

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestDiagnostics(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(previous)
	handler := (&API{}).Handler()
	request := httptest.NewRequest("GET", "/healthz?secret=must-not-log", nil)
	request.Header.Set("X-Request-ID", "untrusted-client-id")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	id := response.Header().Get("X-Request-ID")
	if id == "" || id == "untrusted-client-id" {
		t.Fatalf("invalid generated ID: %q", id)
	}
	output := logs.String()
	if !strings.Contains(output, id) || !strings.Contains(output, `"route":"GET /healthz"`) || !strings.Contains(output, `"status":200`) {
		t.Fatalf("missing correlation/route/status: %s", output)
	}
	if strings.Contains(output, "must-not-log") || strings.Contains(output, "untrusted-client-id") {
		t.Fatal("untrusted data leaked into logs")
	}
}
