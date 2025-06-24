package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func RegisterCustomResultGauge(name string, labelNames []string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "Gauge of Results by Policy",
	}, labelNames)
}

type (
	LabelGenerator = func(openreports.ReportInterface, openreports.ResultAdapter) map[string]string
	LabelCallback  = func(map[string]string, openreports.ReportInterface, openreports.ResultAdapter)
)

func CreateCustomResultMetricsListener(
	filter *report.ResultFilter,
	gauge *prometheus.GaugeVec,
	labelGenerator LabelGenerator,
) report.PolicyReportListener {
	cache := NewCache(filter, labelGenerator)

	return func(event report.LifecycleEvent) {
		newReport := event.PolicyReport

		switch event.Type {
		case report.Added:
			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(newReport, result)).Inc()
			}

			cache.AddReport(newReport)
		case report.Updated:
			items := cache.GetReportLabels(newReport.GetID())
			for _, item := range items {
				decreaseOrDelete(gauge, item.Labels, item.Value)
			}

			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(labelGenerator(newReport, result)).Inc()
			}

			cache.AddReport(newReport)
		case report.Deleted:
			items := cache.GetReportLabels(newReport.GetID())
			for _, item := range items {
				if len(item.Labels) > 0 {
					decreaseOrDelete(gauge, item.Labels, item.Value)
				}
			}

			cache.Remove(newReport.GetID())
		}
	}
}

func decreaseOrDelete(vec *prometheus.GaugeVec, labels map[string]string, gauge float64) {
	m := &dto.Metric{}

	err := vec.With(labels).Write(m)
	if err != nil {
		vec.With(labels).Sub(gauge)
		return
	}

	if *m.Gauge.Value-gauge == 0 {
		vec.Delete(labels)
	} else {
		vec.With(labels).Sub(gauge)
	}
}
