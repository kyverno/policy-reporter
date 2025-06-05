package metrics

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"openreports.io/apis/openreports.io/v1alpha1"
)

type Mode = string

const (
	Simple   Mode = "simple"
	Custom   Mode = "custom"
	Detailed Mode = "detailed"
)

const (
	ReportLabelPrefix    = "label:"
	ReportPropertyPrefix = "property:"
)

var LabelGeneratorMapping = map[string]LabelCallback{
	"namespace": func(m map[string]string, pr openreports.ReportInterface, _ v1alpha1.ReportResult) {
		m["namespace"] = pr.GetNamespace()
	},
	"report": func(m map[string]string, pr openreports.ReportInterface, _ v1alpha1.ReportResult) {
		m["report"] = pr.GetName()
	},
	"policy": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["policy"] = r.Policy
	},
	"rule": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["rule"] = r.Rule
	},
	"kind": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		if !r.HasResource() {
			m["kind"] = ""
			return
		}

		m["kind"] = r.GetResource().Kind
	},
	"name": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		if !r.HasResource() {
			m["name"] = ""
			return
		}

		m["name"] = r.GetResource().Name
	},
	"severity": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["severity"] = string(r.Severity)
	},
	"category": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["category"] = r.Category
	},
	"source": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["source"] = r.Source
	},
	"status": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["status"] = string(r.Result)
	},
	"message": func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
		m["message"] = r.Description
	},
}

func CreateLabelGenerator(labels []string, names []string) LabelGenerator {
	chains := make([]LabelCallback, 0, len(labels))

	for index, label := range labels {
		if strings.HasPrefix(label, ReportLabelPrefix) {
			label := strings.TrimPrefix(label, ReportLabelPrefix)
			lIndex := index

			chains = append(chains, func(m map[string]string, pr openreports.ReportInterface, _ v1alpha1.ReportResult) {
				m[names[lIndex]] = pr.GetLabels()[label]
			})
		} else if strings.HasPrefix(label, ReportPropertyPrefix) {
			label := strings.TrimPrefix(label, ReportPropertyPrefix)
			pIndex := index

			chains = append(chains, func(m map[string]string, _ openreports.ReportInterface, r v1alpha1.ReportResult) {
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

	return func(pr openreports.ReportInterface, r v1alpha1.ReportResult) map[string]string {
		labels := map[string]string{}
		for _, generate := range chains {
			generate(labels, pr, r)
		}

		return labels
	}
}
