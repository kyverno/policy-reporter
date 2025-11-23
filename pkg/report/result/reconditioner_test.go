package result_test

import (
	"testing"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestReconditioner(t *testing.T) {
	t.Run("prepare with default generator", func(t *testing.T) {
		var report openreports.ReportInterface = &openreports.ReportAdapter{
			Report: &v1alpha1.Report{
				ObjectMeta: v1.ObjectMeta{
					Name:      "policy-report",
					Namespace: "test",
				},
				Summary: v1alpha1.ReportSummary{
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
				Results: []v1alpha1.ReportResult{
					{
						Description: "message",
						Result:      v1alpha2.StatusFail,
						Scored:      true,
						Policy:      "required-label",
						Rule:        "app-label-required",
						Timestamp:   v1.Timestamp{Seconds: 1614093000},
						Source:      "test",
						Category:    "",
						Severity:    v1alpha2.SeverityHigh,
						Properties:  map[string]string{"version": "1.2.0"},
					},
				},
			},
		}

		rec := result.NewReconditioner(nil)

		report = rec.Prepare(report)
		res := report.GetResults()[0]

		assert.Equal(t, "1412073110812056002", res.ID)
		assert.Equal(t, "Other", res.Category)
		assert.Equal(t, *report.GetScope(), res.Subjects[0])
	})

	t.Run("prepare with custom generator", func(t *testing.T) {
		var report openreports.ReportInterface = &openreports.ReportAdapter{
			Report: &v1alpha1.Report{
				ObjectMeta: v1.ObjectMeta{
					Name:      "policy-report",
					Namespace: "test",
				},
				Summary: v1alpha1.ReportSummary{
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
				Results: []v1alpha1.ReportResult{
					{
						Description: "message",
						Result:      v1alpha2.StatusFail,
						Scored:      true,
						Policy:      "required-label",
						Rule:        "app-label-required",
						Timestamp:   v1.Timestamp{Seconds: 1614093000},
						Source:      "test",
						Category:    "",
						Severity:    v1alpha2.SeverityHigh,
						Properties:  map[string]string{"version": "1.2.0"},
					},
				},
			},

			Results: []openreports.ResultAdapter{
				{
					ID: "12348",
					ReportResult: v1alpha1.ReportResult{
						Description: "message",
						Result:      v1alpha2.StatusFail,
						Scored:      true,
						Policy:      "required-label",
						Rule:        "app-label-required",
						Timestamp:   v1.Timestamp{Seconds: 1614093000},
						Source:      "test",
						Category:    "",
						Severity:    v1alpha2.SeverityHigh,
						Properties:  map[string]string{"version": "1.2.0"},
					},
				},
			},
		}

		rec := result.NewReconditioner(map[string]result.ReconditionerConfig{
			"test": {
				IDGenerators: result.NewIDGenerator([]string{"policy", "rule", "resource"}),
			},
		})

		report = rec.Prepare(report)
		res := report.GetResults()[0]

		assert.Equal(t, "17886198668009838131", res.ID)
		assert.Equal(t, "Other", res.Category)
		assert.Equal(t, *report.GetScope(), res.Subjects[0])
	})
	t.Run("prepare with self assigned namespace", func(t *testing.T) {
		var report openreports.ReportInterface = &openreports.ReportAdapter{
			Report: &v1alpha1.Report{
				ObjectMeta: v1.ObjectMeta{
					Name:      "policy-report",
					Namespace: "test",
				},
				Summary: v1alpha1.ReportSummary{
					Pass:  0,
					Skip:  0,
					Warn:  0,
					Fail:  1,
					Error: 0,
				},
				Scope: &corev1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "default",
					Namespace:  "",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
				Results: []v1alpha1.ReportResult{
					{
						Description: "message",
						Result:      v1alpha2.StatusFail,
						Scored:      true,
						Policy:      "required-label",
						Rule:        "label-required",
						Timestamp:   v1.Timestamp{Seconds: 1614093000},
						Source:      "test",
						Category:    "",
						Severity:    v1alpha2.SeverityHigh,
						Properties:  map[string]string{"version": "1.2.0"},
					},
				},
			},
		}

		rec := result.NewReconditioner(map[string]result.ReconditionerConfig{
			"test": {
				SelfassignNamespaces: true,
			},
		})

		report = rec.Prepare(report)
		res := report.GetResults()[0]

		assert.Equal(t, "default", res.GetResource().Namespace)
		assert.Equal(t, "11297973392776184291", res.ID)
		assert.Equal(t, "Other", res.Category)
		assert.Equal(t, *report.GetScope(), res.Subjects[0])
	})

}
