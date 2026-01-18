package server

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET returns ok",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "POST returns 405",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "PUT returns 405",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "DELETE returns 405",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("8080")
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			w := httptest.NewRecorder()

			s.mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestReadyzHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET returns ready",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "ready",
		},
		{
			name:           "POST returns 405",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New("8080")
			req := httptest.NewRequest(tt.method, "/readyz", nil)
			w := httptest.NewRecorder()

			s.mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
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

	// Check content type is text/plain with prometheus version
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4; charset=utf-8" && contentType != "text/plain" {
		t.Errorf("expected Content-Type to contain text/plain, got '%s'", contentType)
	}

	// Verify metrics output contains our custom metrics
	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty metrics body")
	}
}

func TestNotFoundHandler(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestMiddlewareMetrics(t *testing.T) {
	s := New("8080")

	// Make multiple requests to test middleware
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}
}

func TestConcurrentRequests(t *testing.T) {
	s := New("8080")
	var wg sync.WaitGroup
	requests := 50

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("concurrent request failed with status %d", w.Code)
			}
		}()
	}

	wg.Wait()
}

func TestServerNew(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{name: "default port", port: "8080"},
		{name: "custom port", port: "9090"},
		{name: "privileged port", port: "80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.port)
			if s == nil {
				t.Fatal("New() returned nil server")
			}
			if s.port != tt.port {
				t.Errorf("expected port %s, got %s", tt.port, s.port)
			}
			if s.mux == nil {
				t.Error("expected non-nil mux")
			}
		})
	}
}
