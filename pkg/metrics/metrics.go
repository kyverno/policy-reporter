package metrics

// Metrics interface for Prometheus Metrics used for the Resolver
type Metrics interface {
	// GenerateMetrics for Prometheus
	GenerateMetrics() error
}
