package listener_test

import (
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

var result1 = &report.Result{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "Best Practices",
	Severity: report.High,
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

var result2 = &report.Result{
	ID:       "124",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Pass,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	},
}

var preport1 = &report.PolicyReport{
	ID:        report.GeneratePolicyReportID("polr-test", "test"),
	Name:      "polr-test",
	Namespace: "test",
	Results: map[string]*report.Result{
		result1.GetIdentifier(): result1,
	},
	Summary:           &report.Summary{Fail: 1},
	CreationTimestamp: time.Now(),
}

var preport2 = &report.PolicyReport{
	ID:        report.GeneratePolicyReportID("polr-test", "test"),
	Name:      "polr-test",
	Namespace: "test",
	Results: map[string]*report.Result{
		result1.GetIdentifier(): result1,
		result2.GetIdentifier(): result2,
	},
	Summary:           &report.Summary{Fail: 1, Pass: 1},
	CreationTimestamp: time.Now(),
}

var creport = &report.PolicyReport{
	Name:              "cpolr-test",
	Summary:           &report.Summary{},
	CreationTimestamp: time.Now(),
}
