package listener_test

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
)

var scopereport1 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.NewTime(time.Now().Add(time.Hour)),
	},
	Scope: &corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "test",
		Namespace:  "test",
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var preport1 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var preport2 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailPodResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1, Pass: 1},
}

var preport3 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{},
}

var creport = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "cpolr-test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailPodResult},
}
