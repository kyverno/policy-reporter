package kubernetes_test

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewFakeCilent() (*fake.Clientset, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset()

	return client, client.Wgpolicyk8sV1alpha2().PolicyReports("test"), client.Wgpolicyk8sV1alpha2().ClusterPolicyReports()
}

func NewMapper() kubernetes.Mapper {
	return kubernetes.NewMapper(make(map[string]string))
}

type store struct {
	store []report.LifecycleEvent
	rwm   *sync.RWMutex
}

func (s *store) Add(r report.LifecycleEvent) {
	s.rwm.Lock()
	s.store = append(s.store, r)
	s.rwm.Unlock()
}

func (s *store) Get(index int) report.LifecycleEvent {
	return s.store[index]
}

func (s *store) List() []report.LifecycleEvent {
	return s.store
}

func newStore(size int) *store {
	return &store{
		store: make([]report.LifecycleEvent, 0, size),
		rwm:   &sync.RWMutex{},
	}
}

var priorityMap = map[string]string{
	"priority-test": "warning",
}

var policyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass:  1,
		Skip:  2,
		Warn:  3,
		Fail:  4,
		Error: 5,
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
			Rule:      "app-label-required",
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
			Properties: map[string]string{"version": "1.2.0", kubernetes.ResultIDKey: "123456"},
		},
	},
}

var multiResourcePolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Pass:  1,
		Skip:  2,
		Warn:  3,
		Fail:  4,
		Error: 5,
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

var minPolicyReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:      "policy-report",
		Namespace: "test",
	},
}

var enforceReportCRD = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "policy-report",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail: 3,
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
					UID:        "",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
		{
			Message:   "message 3",
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
					Name:       "name",
					Namespace:  "test",
					UID:        "",
				},
			},
			Properties: map[string]string{"version": "1.2.0", kubernetes.ResultIDKey: "123456"},
		},
	},
}

var clusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail: 4,
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
			Properties: map[string]string{"version": "1.2.0"},
		},
	},
}

var minClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail: 4,
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

var multiResourceClusterPolicyReportCRD = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name: "cluster-policy-report",
	},
	Summary: v1alpha2.PolicyReportSummary{
		Fail: 4,
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
				{
					APIVersion: "v1",
					Kind:       "Namespace",
					Name:       "policy-reporter-check",
					Namespace:  "test",
					UID:        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
			Properties: map[string]string{"version": "1.2.0"},
		},
	},
}

var result1ID string = report.GeneratePolicyReportResultID(
	string(policyReportCRD.Results[0].Resources[0].UID),
	policyReportCRD.Results[0].Resources[0].Name,
	policyReportCRD.Results[0].Policy,
	policyReportCRD.Results[0].Rule,
	string(policyReportCRD.Results[0].Result),
	policyReportCRD.Results[0].Message,
	policyReportCRD.Results[0].Category,
)

var result2ID string = report.GeneratePolicyReportResultID(
	"",
	"",
	policyReportCRD.Results[1].Policy,
	policyReportCRD.Results[1].Rule,
	string(policyReportCRD.Results[1].Result),
	policyReportCRD.Results[1].Message,
	policyReportCRD.Results[1].Category,
)

var result3ID string = "123456"

var cresult1ID string = report.GeneratePolicyReportResultID(
	string(clusterPolicyReportCRD.Results[0].Resources[0].UID),
	clusterPolicyReportCRD.Results[0].Resources[0].Name,
	clusterPolicyReportCRD.Results[0].Policy,
	clusterPolicyReportCRD.Results[0].Rule,
	string(clusterPolicyReportCRD.Results[0].Result),
	clusterPolicyReportCRD.Results[0].Message,
	clusterPolicyReportCRD.Results[0].Category,
)
