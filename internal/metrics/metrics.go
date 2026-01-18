package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal tracks the total number of HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPInflight tracks the number of inflight HTTP requests
	HTTPInflight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_inflight_requests",
			Help: "Number of inflight HTTP requests",
		},
	)
)
