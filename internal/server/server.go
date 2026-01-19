package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/aLittlecrocodile/devops-practice/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	port         string
	mux          *http.ServeMux
	server       *http.Server
	shutdownOnce sync.Once
}

// New creates a new HTTP server
func New(port string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		port: port,
		mux:  mux,
	}
	s.registerRoutes()
	return s
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.middleware(s.healthz))
	s.mux.HandleFunc("/readyz", s.middleware(s.readyz))
	s.mux.Handle("/metrics", promhttp.Handler())
}

// middleware wraps handlers with common logic (metrics, inflight tracking)
func (s *Server) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.HTTPInflight.Inc()
		defer metrics.HTTPInflight.Dec()

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w}
		next(rw, r)

		// Track response status
		status := rw.status
		if status == 0 {
			status = http.StatusOK
		}
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", status)).Inc()
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// healthz returns 200 if the server is healthy
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// readyz returns 200 if the server is ready
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	log.Printf("Server starting on %s", addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		if s.server != nil {
			log.Printf("Server shutting down")
			_ = s.server.Shutdown(ctx)
		}
	})
	return nil
}
