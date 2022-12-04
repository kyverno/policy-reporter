package violations_test

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/validate"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var filter = email.NewFilter(validate.RuleSets{}, validate.RuleSets{})

func NewFakeCilent() (v1alpha2client.Wgpolicyk8sV1alpha2Interface, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset().Wgpolicyk8sV1alpha2()

	return client, client.PolicyReports("test"), client.ClusterPolicyReports()
}

var PolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass:  0,
		Skip:  0,
		Warn:  0,
		Fail:  3,
		Error: 0,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "required-label",
			Rule:      "app-label-required",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
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
			Message:   "message 2",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "priority-test",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
		},
		{
			Message:   "message 3",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "required-label",
			Rule:      "",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Deployment",
					Name:       "name",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988b3",
				},
			},
		},
	},
}

var KyvernoPolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kyverno-policy-report",
		Namespace: "kyverno",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass:  1,
		Skip:  0,
		Warn:  1,
		Fail:  0,
		Error: 0,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusPass,
			Scored:    true,
			Policy:    "required-limit",
			Rule:      "resource-limit-required",
			Timestamp: v1.Timestamp{Seconds: 1614093003},
			Source:    "Kyverno",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
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
			Message:   "message",
			Result:    v1alpha2.StatusWarn,
			Scored:    true,
			Policy:    "required-limit",
			Rule:      "resource-limit-required",
			Timestamp: v1.Timestamp{Seconds: 1614093003},
			Source:    "Kyverno",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
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

var PassPolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "pass-policy-report",
		Namespace: "test",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass:  1,
		Skip:  0,
		Warn:  0,
		Fail:  0,
		Error: 0,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusPass,
			Scored:    true,
			Policy:    "required-limit",
			Rule:      "resource-limit-required",
			Timestamp: v1.Timestamp{Seconds: 1614093003},
			Source:    "Kyverno",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
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

var EmptyPolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "empty-policy-report",
		Namespace: "test",
	},
	Summary: v1alpha2.PolicyReportSummary{},
}

var ClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail:  1,
		Warn:  0,
		Error: 0,
		Pass:  1,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "cluster-required-label",
			Rule:      "ns-label-required",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "policy-reporter",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
		{
			Message:   "message",
			Result:    v1alpha2.StatusPass,
			Scored:    true,
			Policy:    "cluster-required-label",
			Rule:      "ns-label-required",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
				{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "test",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}

var KyvernoClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "kyverno-cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail:  1,
		Warn:  0,
		Error: 0,
		Pass:  0,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "cluster-required-quota",
			Rule:      "ns-quota-required",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "Kyverno",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
			Resources: []corev1.ObjectReference{
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

var MinClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail: 1,
		Pass: 1,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusFail,
			Scored:    true,
			Policy:    "cluster-policy",
			Rule:      "cluster-role",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
		},
	},
}

var PassClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "pass-cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass: 1,
	},
	Results: []v1alpha2.PolicyReportResult{
		{
			Message:   "message",
			Result:    v1alpha2.StatusPass,
			Scored:    true,
			Policy:    "cluster-policy-pass",
			Rule:      "cluster-role-pass",
			Timestamp: v1.Timestamp{Seconds: 1614093000},
			Source:    "test",
			Category:  "test",
			Severity:  v1alpha2.SeverityHigh,
		},
	},
}

var EmptyClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "empty-cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{},
}
