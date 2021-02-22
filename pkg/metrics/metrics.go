package metrics

type Metrics interface {
	GenerateMetrics() error
}
