package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aLittlecrocodile/devops-practice/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	port string
	mux  *http.ServeMux
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

		// Track response status
		status := "200"
		defer func() {
			metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		}()

		// Call the actual handler
		next(w, r)
	}
}

// healthz returns 200 if the server is healthy
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// readyz returns 200 if the server is ready
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}
