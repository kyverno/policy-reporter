package kubernetes_test

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

var policyReportSchema = schema.GroupVersionResource{
	Group:    "wgpolicyk8s.io",
	Version:  "v1alpha2",
	Resource: "policyreports",
}

var clusterPolicyReportSchema = schema.GroupVersionResource{
	Group:    "wgpolicyk8s.io",
	Version:  "v1alpha2",
	Resource: "clusterpolicyreports",
}

var gvrToListKind = map[schema.GroupVersionResource]string{
	policyReportSchema:        "PolicyReportList",
	clusterPolicyReportSchema: "ClusterPolicyReportList",
}

func NewFakeCilent() (dynamic.Interface, dynamic.ResourceInterface) {
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToListKind)

	return client, client.Resource(policyReportSchema).Namespace("test")
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

var policyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "policy-report",
		"namespace":         "test",
		"creationTimestamp": "2021-02-23T15:00:00Z",
	},
	"summary": map[string]interface{}{
		"pass":  int64(1),
		"skip":  int64(2),
		"warn":  int64(3),
		"fail":  int64(4),
		"error": int64(5),
	},
	"results": []interface{}{
		map[string]interface{}{
			"message": "message",
			"result":  "fail",
			"scored":  true,
			"policy":  "required-label",
			"rule":    "app-label-required",
			"timestamp": map[string]interface{}{
				"seconds": int64(1614093000),
			},
			"source":   "test",
			"category": "test",
			"severity": "high",
			"resources": []interface{}{
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Deployment",
					"name":       "nginx",
					"namespace":  "test",
					"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
			"properties": map[string]interface{}{
				"version": "1.2.0",
			},
		},
		map[string]interface{}{
			"message": "message 2",
			"result":  "fail",
			"scored":  true,
			"timestamp": map[string]interface{}{
				"seconds": int64(1614093000),
			},
			"policy":    "priority-test",
			"resources": []interface{}{},
		},
	},
}

var minPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":      "policy-report",
		"namespace": "test",
	},
	"results": []interface{}{},
}

var clusterPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "clusterpolicy-report",
		"creationTimestamp": "2021-02-23T15:00:00Z",
	},
	"summary": map[string]interface{}{
		"pass":  int64(1),
		"skip":  int64(2),
		"warn":  int64(3),
		"fail":  int64(4),
		"error": int64(5),
	},
	"results": []interface{}{
		map[string]interface{}{
			"message":   "message",
			"result":    "fail",
			"scored":    true,
			"policy":    "cluster-required-label",
			"rule":      "ns-label-required",
			"category":  "test",
			"severity":  "high",
			"timestamp": map[string]interface{}{"seconds": ""},
			"resources": []interface{}{
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"name":       "policy-reporter",
					"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}

var priorityMap = map[string]string{
	"priority-test": "warning",
}

var result1ID string = report.GeneratePolicyReportResultID("dfd57c50-f30c-4729-b63f-b1954d8988d1", "required-label", "app-label-required", "fail", "message")
var result2ID string = report.GeneratePolicyReportResultID("", "priority-test", "", "fail", "message 2")
