package listener_test

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var scopereport1 = &openreports.ORReportAdapter{
	Report: &v1alpha1.Report{
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
		Results: []v1alpha1.ReportResult{*fixtures.FailResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1},
	},
}

var preport1 = &openreports.ORReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{*fixtures.FailResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1},
	},
}

var preport2 = &openreports.ORReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{*fixtures.FailPodResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1, Pass: 1},
	},
}

var preport3 = &openreports.ORReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{},
	},
}

var creport = &openreports.ORClusterReportAdapter{
	ClusterReport: &v1alpha1.ClusterReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{*fixtures.FailResult.ReportResult, *fixtures.FailPodResult.ReportResult},
	},
}
