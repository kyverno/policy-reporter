package metrics

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type Mode = string

const (
	Simple   Mode = "simple"
	Custom   Mode = "custom"
	Detailed Mode = "detailed"
)

const ReportLabelPrefix = "label:"
const ReportPropertyPrefix = "property:"

var LabelGeneratorMapping = map[string]LabelCallback{
	"namespace": func(m map[string]string, pr v1alpha2.ReportInterface, _ v1alpha2.PolicyReportResult) {
		m["namespace"] = pr.GetNamespace()
	},
	"report": func(m map[string]string, pr v1alpha2.ReportInterface, _ v1alpha2.PolicyReportResult) {
		m["report"] = pr.GetName()
	},
	"policy": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["policy"] = r.Policy
	},
	"rule": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["rule"] = r.Rule
	},
	"kind": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		if !r.HasResource() {
			m["kind"] = ""
			return
		}

		m["kind"] = r.GetResource().Kind
	},
	"name": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		if !r.HasResource() {
			m["name"] = ""
			return
		}

		m["name"] = r.GetResource().Name
	},
	"severity": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["severity"] = string(r.Severity)
	},
	"category": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["category"] = r.Category
	},
	"source": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["source"] = r.Source
	},
	"status": func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		m["status"] = string(r.Result)
	},
}

func CreateLabelGenerator(labels []string, names []string) LabelGenerator {
	chains := make([]LabelCallback, 0, len(labels))

	for index, label := range labels {
		if strings.HasPrefix(label, ReportLabelPrefix) {
			label := strings.TrimPrefix(label, ReportLabelPrefix)
			lIndex := index

			chains = append(chains, func(m map[string]string, pr v1alpha2.ReportInterface, _ v1alpha2.PolicyReportResult) {
				m[names[lIndex]] = pr.GetLabels()[label]
			})
		} else if strings.HasPrefix(label, ReportPropertyPrefix) {
			label := strings.TrimPrefix(label, ReportPropertyPrefix)
			pIndex := index

			chains = append(chains, func(m map[string]string, _ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
				val := ""

				if r.Properties != nil {
					val = r.Properties[label]
				}

				m[names[pIndex]] = val
			})
		} else if callback, ok := LabelGeneratorMapping[label]; ok {
			chains = append(chains, callback)
		}
	}

	return func(pr v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) map[string]string {
		labels := map[string]string{}
		for _, generate := range chains {
			generate(labels, pr, r)
		}

		return labels
	}
}
