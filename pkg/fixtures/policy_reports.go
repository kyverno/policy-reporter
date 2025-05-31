package fixtures

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

var DefaultMeta = &v1.PartialObjectMetadata{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	TypeMeta: v1.TypeMeta{
		Kind:       "PolicyReport",
		APIVersion: "wgpolicyk8s.io/v1alpha2",
	},
}

var DefaultPolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha1.ReportSummary{
		Pass:  0,
		Skip:  0,
		Warn:  0,
		Fail:  3,
		Error: 0,
	},
	Results: []v1alpha1.ReportResult{
		{
			ID:          "12348",
			Description: "message",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "required-label",
			Rule:        "app-label-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
		{
			ID:          "12346",
			Description: "message 2",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "priority-test",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
		},
		{
			ID:          "12347",
			Description: "message 3",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "required-label",
			Rule:        "app-label-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "name",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988b3",
				},
			},
			Properties: map[string]string{"version": "1.2.0", v1alpha2.ResultIDKey: "123456"},
		},
	},
}

var ScopePolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha1.ReportSummary{
		Pass:  0,
		Skip:  0,
		Warn:  0,
		Fail:  3,
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
			ID:          "12348",
			Description: "message",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "required-label",
			Rule:        "app-label-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Properties:  map[string]string{"version": "1.2.0"},
		},
	},
}

var MultiResourcePolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha1.ReportSummary{
		Pass:  1,
		Skip:  2,
		Warn:  3,
		Fail:  4,
		Error: 5,
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
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx2",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d2",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
	},
}

var MinPolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
}

var EnforceReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:              "policy-report",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Summary: v1alpha1.ReportSummary{
		Fail: 3,
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
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "test",
					UID:        "",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
		{
			Description: "message 3",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "required-label",
			Rule:        "app-label-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "name",
					Namespace:  "test",
					UID:        "",
				},
			},
			Properties: map[string]string{"version": "1.2.0", v1alpha2.ResultIDKey: "123456"},
		},
	},
}

var DefaultClusterMeta = &v1.PartialObjectMetadata{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	TypeMeta: v1.TypeMeta{
		Kind:       "ClusterPolicyReport",
		APIVersion: "wgpolicyk8s.io/v1alpha2",
	},
}

var ClusterPolicyReport = &v1alpha1.ClusterReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha1.ReportSummary{
		Fail: 4,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "cluster-required-label",
			Rule:        "ns-label-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "policy-reporter",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
	},
}

var MinClusterPolicyReport = &v1alpha1.ClusterReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha1.ReportSummary{
		Fail: 4,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "cluster-policy",
			Rule:        "cluster-role",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
		},
	},
}

var PassPolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "pass-policy-report",
		Namespace: "test",
	},
	Summary: v1alpha1.ReportSummary{
		Pass:  1,
		Skip:  0,
		Warn:  0,
		Fail:  0,
		Error: 0,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusPass,
			Scored:      true,
			Policy:      "required-limit",
			Rule:        "resource-limit-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093003},
			Source:      "Kyverno",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}

var EmptyPolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "empty-policy-report",
		Namespace: "test",
	},
	Summary: v1alpha1.ReportSummary{},
}

var PassClusterPolicyReport = &v1alpha1.ClusterReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "pass-cluster-policy-report",
	},
	Summary: v1alpha1.ReportSummary{
		Pass: 1,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusPass,
			Scored:      true,
			Policy:      "cluster-policy-pass",
			Rule:        "cluster-role-pass",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "test",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
		},
	},
}

var EmptyClusterPolicyReport = &v1alpha1.ClusterReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "empty-cluster-policy-report",
	},
	Summary: v1alpha1.ReportSummary{},
}

var KyvernoPolicyReport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kyverno-policy-report",
		Namespace: "kyverno",
	},
	Summary: v1alpha1.ReportSummary{
		Pass:  1,
		Skip:  0,
		Warn:  1,
		Fail:  0,
		Error: 0,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusPass,
			Scored:      true,
			Policy:      "required-limit",
			Rule:        "resource-limit-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093003},
			Source:      "Kyverno",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "kyverno",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
		{
			Description: "message",
			Result:      v1alpha2.StatusWarn,
			Scored:      true,
			Policy:      "required-limit",
			Rule:        "resource-limit-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093003},
			Source:      "Kyverno",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "nginx2",
					Namespace:  "kyverno",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d2",
				},
			},
		},
	},
}

var KyvernoClusterPolicyReport = &v1alpha1.ClusterReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "kyverno-cluster-policy-report",
	},
	Summary: v1alpha1.ReportSummary{
		Fail:  1,
		Warn:  0,
		Error: 0,
		Pass:  0,
	},
	Results: []v1alpha1.ReportResult{
		{
			Description: "message",
			Result:      v1alpha2.StatusFail,
			Scored:      true,
			Policy:      "cluster-required-quota",
			Rule:        "ns-quota-required",
			Timestamp:   v1.Timestamp{Seconds: 1614093000},
			Source:      "Kyverno",
			Category:    "test",
			Severity:    v1alpha2.SeverityHigh,
			Subjects: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "kyverno",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}
