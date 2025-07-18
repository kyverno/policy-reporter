package listener_test

import (
	"time"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var scopereport1 = &openreports.ReportAdapter{
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
		Results: []v1alpha1.ReportResult{fixtures.FailResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1},
	},
	Results: []openreports.ResultAdapter{fixtures.FailResult},
}

var preport1 = &openreports.ReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{fixtures.FailResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1},
	},
	Results: []openreports.ResultAdapter{fixtures.FailResult},
}

var preport2 = &openreports.ReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{fixtures.FailPodResult.ReportResult},
		Summary: v1alpha1.ReportSummary{Fail: 1, Pass: 1},
	},
	Results: []openreports.ResultAdapter{fixtures.FailPodResult},
}

var preport3 = &openreports.ReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{},
	},
}

var creport = &openreports.ClusterReportAdapter{
	ClusterReport: &v1alpha1.ClusterReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Results: []v1alpha1.ReportResult{fixtures.FailResult.ReportResult, fixtures.FailPodResult.ReportResult},
	},
	Results: []openreports.ResultAdapter{fixtures.FailResult, fixtures.FailPodResult},
}
