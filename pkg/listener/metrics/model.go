package metrics

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/report"
)

type Mode = string

const (
	Simple   Mode = "simple"
	Custom   Mode = "custom"
	Detailed Mode = "detailed"
)

const ReportLabelPrefix = "label:"

var LabelGeneratorMapping = map[string]LabelCallback{
	"namespace": func(m map[string]string, pr report.PolicyReport, _ report.Result) {
		m["namespace"] = pr.Namespace
	},
	"report": func(m map[string]string, pr report.PolicyReport, _ report.Result) {
		m["report"] = pr.Name
	},
	"policy": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["policy"] = r.Policy
	},
	"rule": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["rule"] = r.Rule
	},
	"kind": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["kind"] = r.Resource.Kind
	},
	"name": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["name"] = r.Resource.Name
	},
	"severity": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["severity"] = r.Severity
	},
	"category": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["category"] = r.Category
	},
	"source": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["source"] = r.Source
	},
	"status": func(m map[string]string, _ report.PolicyReport, r report.Result) {
		m["status"] = r.Status
	},
}

func CreateLabelGenerator(labels []string, names []string) LabelGenerator {
	chains := make([]func(map[string]string, report.PolicyReport, report.Result), 0, len(labels))

	for index, label := range labels {
		if strings.HasPrefix(label, ReportLabelPrefix) {
			label = strings.TrimPrefix(label, ReportLabelPrefix)

			chains = append(chains, func(m map[string]string, pr report.PolicyReport, _ report.Result) {
				m[names[index]] = pr.Labels[label]
			})
		}
		if callback, ok := LabelGeneratorMapping[label]; ok {
			chains = append(chains, callback)
		}
	}

	return func(pr report.PolicyReport, r report.Result) map[string]string {
		labels := map[string]string{}
		for _, generate := range chains {
			generate(labels, pr, r)
		}

		return labels
	}
}
