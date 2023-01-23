package report_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var result1 = v1alpha2.PolicyReportResult{
	ID:       "16097155368874536783",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Category: "resources",
	Severity: v1alpha2.SeverityHigh,
	Scored:   true,
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
}

var result2 = v1alpha2.PolicyReportResult{
	ID:       "2",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Category: "resources",
	Scored:   true,
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
}

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
		Results: []v1alpha2.PolicyReportResult{result1},
		Summary: v1alpha2.PolicyReportSummary{},
	}
	preport2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha2.PolicyReportResult{result1, result2},
		Summary: v1alpha2.PolicyReportSummary{},
	}

	diff := report.FindNewResults(preport2, preport1)
	if len(diff) != 1 {
		t.Fatal("should only return one new result")
	}
	if diff[0].GetResource().UID != result2.GetResource().UID {
		t.Fatal("should only return the new result2")
	}

	diff2 := report.FindNewResults(preport2, nil)
	if len(diff2) != 2 {
		t.Fatal("should return all results in the new report")
	}
}
