package listener_test

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailPodResult},
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
