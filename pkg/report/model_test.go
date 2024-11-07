package report_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var preport = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: make([]v1alpha2.PolicyReportResult, 0),
	Summary: v1alpha2.PolicyReportSummary{},
}

var creport = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "cpolr-test",
		CreationTimestamp: v1.Now(),
	},
	Results: make([]v1alpha2.PolicyReportResult, 0),
	Summary: v1alpha2.PolicyReportSummary{},
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
	preport1 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha2.PolicyReportResult{fixtures.FailResult},
		Summary: v1alpha2.PolicyReportSummary{},
	}
	preport2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailPodResult},
		Summary: v1alpha2.PolicyReportSummary{},
	}

	diff := report.FindNewResults(preport2, preport1)
	assert.Len(t, diff, 1, "should only return one new result")
	assert.Equal(t, fixtures.FailPodResult.GetResource().UID, diff[0].GetResource().UID, "should only return the new result2")

	diff2 := report.FindNewResults(preport2, nil)
	assert.Len(t, diff2, 2, "should return all results in the new report")
}
