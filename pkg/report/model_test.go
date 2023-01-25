package report_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		if report.Added.String() != "add" {
			t.Errorf("Unexpected type conversion, expected %s go %s", "add", report.Added.String())
		}
		if report.Updated.String() != "update" {
			t.Errorf("Unexpected type conversion, expected %s go %s", "update", report.Updated.String())
		}
		if report.Deleted.String() != "delete" {
			t.Errorf("Unexpected type conversion, expected %s go %s", "delete", report.Deleted.String())
		}
		if report.Event(4).String() != "unknown" {
			t.Errorf("Unexpected type conversion, expected %s go %s", "unknown", report.Event(4).String())
		}
	})
}

func Test_GetType(t *testing.T) {
	pr := report.GetType(preport)
	if pr != report.PolicyReportType {
		t.Fatal("expected type policy report")
	}

	cpr := report.GetType(creport)
	if cpr != report.ClusterPolicyReportType {
		t.Fatal("expected type cluster policy report")
	}
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
	if len(diff) != 1 {
		t.Fatal("should only return one new result")
	}
	if diff[0].GetResource().UID != fixtures.FailPodResult.GetResource().UID {
		t.Fatal("should only return the new result2")
	}

	diff2 := report.FindNewResults(preport2, nil)
	if len(diff2) != 2 {
		t.Fatal("should return all results in the new report")
	}
}
