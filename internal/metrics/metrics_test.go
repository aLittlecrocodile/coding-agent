package metrics

import (
	"testing"
)

func TestMetricsAreDefined(t *testing.T) {
	// Verify that metrics are properly initialized and not nil
	if HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal metric should not be nil")
	}

	if HTTPInflight == nil {
		t.Error("HTTPInflight metric should not be nil")
	}
}

func TestMetricsCanBeRecorded(t *testing.T) {
	// Test that we can record metrics with labels
	// This should not panic
	HTTPRequestsTotal.WithLabelValues("GET", "/healthz", "200").Inc()
	HTTPInflight.Inc()
	HTTPInflight.Dec()

	// Test with various label combinations
	labelSets := []struct {
		method, path, status string
	}{
		{"GET", "/healthz", "200"},
		{"GET", "/readyz", "200"},
		{"POST", "/healthz", "405"},
		{"GET", "/metrics", "200"},
	}

	for _, labels := range labelSets {
		HTTPRequestsTotal.WithLabelValues(labels.method, labels.path, labels.status).Inc()
	}

	// If we got here without panic, the test passes
}

func TestMetricsConcurrency(t *testing.T) {
	// Test that metrics are safe for concurrent use
	done := make(chan struct{})

	// Spawn multiple goroutines recording metrics
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Inc()
				HTTPInflight.Inc()
				HTTPInflight.Dec()
			}
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
