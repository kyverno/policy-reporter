package report_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var preport = &openreports.ReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: make([]v1alpha1.ReportResult, 0),
		Summary: v1alpha1.ReportSummary{},
	},
}

var creport = &openreports.ClusterReportAdapter{
	ClusterReport: &v1alpha1.ClusterReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Results: make([]v1alpha1.ReportResult, 0),
		Summary: v1alpha1.ReportSummary{},
	},
}

func Test_Events(t *testing.T) {
	t.Run("Event.String", func(t *testing.T) {
		assert.Equal(t, "add", report.Added.String(), "Unexpected type conversion")
		assert.Equal(t, "update", report.Updated.String(), "Unexpected type conversion")
		assert.Equal(t, "delete", report.Deleted.String(), "Unexpected type conversion")
		assert.Equal(t, "unknown", report.Event(4).String(), "Unexpected type conversion")
	})
}

func Test_GetType(t *testing.T) {
	assert.Equal(t, report.PolicyReportType, report.GetType(preport), "expected type")
	assert.Equal(t, report.ClusterPolicyReportType, report.GetType(creport), "expected type")
}

func Test_FindNewEvents(t *testing.T) {
	preport1 := &openreports.ReportAdapter{
		Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{
				Name:              "polr-test",
				Namespace:         "test",
				CreationTimestamp: v1.Now(),
			},
			Results: []v1alpha1.ReportResult{fixtures.FailResult.ReportResult},
			Summary: v1alpha1.ReportSummary{},
		},
		Results: []openreports.ORResultAdapter{fixtures.FailResult},
	}
	preport2 := &openreports.ReportAdapter{
		Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{
				Name:              "polr-test",
				Namespace:         "test",
				CreationTimestamp: v1.Now(),
			},
			Results: []v1alpha1.ReportResult{fixtures.FailResult.ReportResult, fixtures.FailPodResult.ReportResult},
			Summary: v1alpha1.ReportSummary{},
		},
		Results: []openreports.ORResultAdapter{fixtures.FailResult, fixtures.FailPodResult},
	}

	diff := report.FindNewResults(preport2, preport1)
	assert.Len(t, diff, 1, "should only return one new result")
	assert.Equal(t, fixtures.FailPodResult.GetResource().UID, diff[0].GetResource().UID, "should only return the new result2")

	diff2 := report.FindNewResults(preport2, nil)
	assert.Len(t, diff2, 2, "should return all results in the new report")
}
