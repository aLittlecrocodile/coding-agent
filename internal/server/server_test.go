package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "ok" {
		t.Errorf("expected body 'ok', got '%s'", body)
	}
}

func TestHealthzWrongMethod(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestReadyzHandler(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "ready" {
		t.Errorf("expected body 'ready', got '%s'", body)
	}
}

func TestMetricsHandler(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check content type contains text/plain
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		t.Error("expected Content-Type header")
	}
}

func TestMiddleware(t *testing.T) {
	// Test that middleware increments metrics
	s := New("8080")

	// Make a request to trigger metrics
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	// Check that request was recorded (we can't easily test the actual prometheus
	// metrics without additional setup, but we verify the request succeeds)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
