package result_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestReconditioner(t *testing.T) {
	t.Run("prepare with default generator", func(t *testing.T) {
		var report v1alpha2.ReportInterface = &v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:      "policy-report",
				Namespace: "test",
			},
			Summary: v1alpha2.PolicyReportSummary{
				Pass:  0,
				Skip:  0,
				Warn:  0,
				Fail:  1,
				Error: 0,
			},
			Scope: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "nginx",
				Namespace:  "test",
				UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
			},
			Results: []v1alpha2.PolicyReportResult{
				{
					ID:         "12348",
					Message:    "message",
					Result:     v1alpha2.StatusFail,
					Scored:     true,
					Policy:     "required-label",
					Rule:       "app-label-required",
					Timestamp:  v1.Timestamp{Seconds: 1614093000},
					Source:     "test",
					Category:   "",
					Severity:   v1alpha2.SeverityHigh,
					Properties: map[string]string{"version": "1.2.0"},
				},
			},
		}

		rec := result.NewReconditioner(nil)

		report = rec.Prepare(report)
		res := report.GetResults()[0]

		if res.ID != "1412073110812056002" {
			t.Errorf("result id should be generated from default generator: %s", res.ID)
		}
		if res.Category != "Other" {
			t.Error("result category should default to Other")
		}
		if res.Priority != v1alpha2.ErrorPriority {
			t.Error("result prioriry should be mapped")
		}
		if len(res.Resources) == 0 || res.Resources[0] != *report.GetScope() {
			t.Error("result resource should be mapped to scope")
		}
	})

	t.Run("prepare with custom generator", func(t *testing.T) {
		var report v1alpha2.ReportInterface = &v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:      "policy-report",
				Namespace: "test",
			},
			Summary: v1alpha2.PolicyReportSummary{
				Pass:  0,
				Skip:  0,
				Warn:  0,
				Fail:  1,
				Error: 0,
			},
			Scope: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "nginx",
				Namespace:  "test",
				UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
			},
			Results: []v1alpha2.PolicyReportResult{
				{
					ID:         "12348",
					Message:    "message",
					Result:     v1alpha2.StatusFail,
					Scored:     true,
					Policy:     "required-label",
					Rule:       "app-label-required",
					Timestamp:  v1.Timestamp{Seconds: 1614093000},
					Source:     "test",
					Category:   "",
					Severity:   v1alpha2.SeverityHigh,
					Properties: map[string]string{"version": "1.2.0"},
				},
			},
		}

		rec := result.NewReconditioner(map[string]result.IDGenerator{
			"test": result.NewIDGenerator([]string{"resource", "policy", "rule", "result"}),
		})

		report = rec.Prepare(report)
		res := report.GetResults()[0]

		if res.ID != "12714703365089292087" {
			t.Errorf("result id should be generated from custom generator: %s", res.ID)
		}
		if res.Category != "Other" {
			t.Error("result category should default to Other")
		}
		if res.Priority != v1alpha2.ErrorPriority {
			t.Error("result prioriry should be mapped")
		}
		if len(res.Resources) == 0 || res.Resources[0] != *report.GetScope() {
			t.Error("result resource should be mapped to scope")
		}
	})
}
