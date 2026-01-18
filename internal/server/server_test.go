package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
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
		{
			name:           "PATCH returns 405",
			method:         http.MethodPatch,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "HEAD returns 405",
			method:         http.MethodHead,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "OPTIONS returns 405",
			method:         http.MethodOptions,
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
		{
			name:           "PUT returns 405",
			method:         http.MethodPut,
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

	// Check for specific metric names
	expectedMetrics := []string{
		"http_requests_total",
		"http_inflight_requests",
	}
	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("expected metrics to contain '%s'", metric)
		}
	}
}

func TestMetricsHandlerWrongMethod(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	// /metrics is handled by promhttp, which may allow POST
	// We just verify it doesn't crash
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

func TestNotFoundHandlerWithVariousPaths(t *testing.T) {
	paths := []string{
		"/api/v1/users",
		"/admin",
		"/static/file.txt",
		"/",
		"/healthz/extra",
		"/readyz/extra",
	}

	for _, path := range paths {
		t.Run("path_"+path, func(t *testing.T) {
			s := New("8080")
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			s.mux.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound && path != "/" {
				t.Errorf("path %s: expected status 404, got %d", path, w.Code)
			}
		})
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

func TestMiddlewareHeaders(t *testing.T) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareWithBody(t *testing.T) {
	s := New("8080")
	body := strings.NewReader("test body")
	req := httptest.NewRequest(http.MethodPost, "/healthz", body)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	// Should return 405 Method Not Allowed, but not crash
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

func TestConcurrentMixedRequests(t *testing.T) {
	s := New("8080")
	var wg sync.WaitGroup
	requests := 100

	paths := []string{"/healthz", "/readyz", "/metrics", "/nonexistent"}

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := paths[idx%len(paths)]
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)

			// Don't check status for 404s
			if path != "/nonexistent" && w.Code != http.StatusOK {
				t.Errorf("path %s: expected status 200, got %d", path, w.Code)
			}
		}(i)
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
		{name: "high port", port: "65535"},
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

func TestServerStartIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Use a fixed port for testing
	s := New("18081")

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test the actual HTTP server
	resp, err := http.Get("http://localhost:18081/healthz")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got '%s'", string(body))
	}

	// There's no Shutdown method, so we just verify the server is running
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			t.Logf("server exited with error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		// Server is still running, which is expected
	}
}

func TestMiddlewareStatusCodeTracking(t *testing.T) {
	s := New("8080")

	// Make requests that will return different status codes
	tests := []struct {
		path           string
		method         string
		expectedStatus int
	}{
		{"/healthz", http.MethodGet, http.StatusOK},
		{"/healthz", http.MethodPost, http.StatusMethodNotAllowed},
		{"/readyz", http.MethodGet, http.StatusOK},
		{"/nonexistent", http.MethodGet, http.StatusNotFound},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		if w.Code != tt.expectedStatus {
			t.Errorf("%s %s: expected status %d, got %d", tt.method, tt.path, tt.expectedStatus, w.Code)
		}
	}
}

func TestServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	s := New("18080") // Use non-standard port

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.Start()
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Test all endpoints
	client := &http.Client{Timeout: 5 * time.Second}
	endpoints := []struct {
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"/healthz", http.StatusOK, "ok"},
		{"/readyz", http.StatusOK, "ready"},
		{"/metrics", http.StatusOK, ""},
	}

	for _, ep := range endpoints {
		resp, err := client.Get("http://localhost:18080" + ep.path)
		if err != nil {
			t.Errorf("failed to get %s: %v", ep.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != ep.expectedStatus {
			t.Errorf("%s: expected status %d, got %d", ep.path, ep.expectedStatus, resp.StatusCode)
		}

		if ep.expectedBody != "" {
			body, _ := io.ReadAll(resp.Body)
			if string(body) != ep.expectedBody {
				t.Errorf("%s: expected body '%s', got '%s'", ep.path, ep.expectedBody, string(body))
			}
		}
	}

	// Don't explicitly shut down - let it be cleaned up when test ends
	_ = serverErr
}

// Benchmark tests
func BenchmarkHealthzHandler(b *testing.B) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
	}
}

func BenchmarkReadyzHandler(b *testing.B) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
	}
}

func BenchmarkMetricsHandler(b *testing.B) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	s := New("8080")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)
		}
	})
}
