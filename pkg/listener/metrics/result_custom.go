package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func RegisterCustomResultGauge(name string, labelNames []string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "Gauge of Results by Policy",
	}, labelNames)
}

type LabelGenerator = func(report.PolicyReport, report.Result) map[string]string
type LabelCallback = func(map[string]string, report.PolicyReport, report.Result)

func CreateCustomResultMetricsListener(
	filter *report.ResultFilter,
	gauge *prometheus.GaugeVec,
	labelGenerator LabelGenerator,
) report.PolicyReportListener {
	var newReport report.PolicyReport
	var oldReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		oldReport = event.OldPolicyReport

		switch event.Type {
		case report.Added:
			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(newReport, result)).Inc()
			}
		case report.Updated:
			for _, result := range oldReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(oldReport, result)).Dec()
			}

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(newReport, result)).Inc()
			}
		case report.Deleted:
			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(newReport, result)).Dec()
			}
		}
	}
}
