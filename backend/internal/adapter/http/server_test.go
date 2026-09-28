package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerAllowsConfiguredFrontendOrigin(t *testing.T) {
	e := NewServer(nil, nil, nil, "http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/token", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestServerRejectsUnknownFrontendOrigin(t *testing.T) {
	e := NewServer(nil, nil, nil, "http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/token", nil)
	req.Header.Set("Origin", "https://attacker.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Access-Control-Allow-Origin = %q", got)
	}
}

func TestServerAllowsAdminUpdatePreflight(t *testing.T) {
	e := NewServer(nil, nil, nil, "http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/admin/clients/client-1", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("missing allowed methods")
	}
}
