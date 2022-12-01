package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	dto "github.com/prometheus/client_model/go"
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

				decreaseOrDelete(gauge, labelGenerator(oldReport, result))
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

				decreaseOrDelete(gauge, labelGenerator(newReport, result))
			}
		}
	}
}

func decreaseOrDelete(vec *prometheus.GaugeVec, labels map[string]string) {
	m := &dto.Metric{}

	err := vec.With(labels).Write(m)
	if err != nil {
		vec.With(labels).Dec()
		return
	}

	if *m.Gauge.Value == 1 {
		vec.Delete(labels)
	} else {
		vec.With(labels).Dec()
	}
}
