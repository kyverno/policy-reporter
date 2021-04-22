package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type PolicyReportAdapter interface {
	ListClusterPolicyReports() (*unstructured.UnstructuredList, error)
	ListPolicyReports() (*unstructured.UnstructuredList, error)
	WatchClusterPolicyReports() (watch.Interface, error)
	WatchPolicyReports() (watch.Interface, error)
}

type k8sPolicyReportAdapter struct {
	client               dynamic.Interface
	policyReports        schema.GroupVersionResource
	clusterPolicyReports schema.GroupVersionResource
}

func (k *k8sPolicyReportAdapter) ListClusterPolicyReports() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(k.clusterPolicyReports).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) ListPolicyReports() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(k.policyReports).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) WatchClusterPolicyReports() (watch.Interface, error) {
	return k.client.Resource(k.clusterPolicyReports).Watch(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) WatchPolicyReports() (watch.Interface, error) {
	return k.client.Resource(k.policyReports).Watch(context.Background(), metav1.ListOptions{})
}

// NewPolicyReportAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyReportAdapter(dynamic dynamic.Interface, version string) PolicyReportAdapter {
	if version == "" {
		version = "v1alpha1"
	}

	return &k8sPolicyReportAdapter{
		client: dynamic,
		policyReports: schema.GroupVersionResource{
			Group:    "wgpolicyk8s.io",
			Version:  version,
			Resource: "policyreports",
		},
		clusterPolicyReports: schema.GroupVersionResource{
			Group:    "wgpolicyk8s.io",
			Version:  version,
			Resource: "clusterpolicyreports",
		},
	}
}
