package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var (
	PolicyReports        = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "policyreports"}
	ClusterPolicyReports = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "clusterpolicyreports"}
)

type PolicyReportAdapter interface {
	ListClusterPolicyReports() (*unstructured.UnstructuredList, error)
	ListPolicyReports() (*unstructured.UnstructuredList, error)
	WatchClusterPolicyReports() (watch.Interface, error)
	WatchPolicyReports() (watch.Interface, error)
}

type k8sPolicyReportAdapter struct {
	client dynamic.Interface
}

func (k *k8sPolicyReportAdapter) ListClusterPolicyReports() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(ClusterPolicyReports).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) ListPolicyReports() (*unstructured.UnstructuredList, error) {
	return k.client.Resource(PolicyReports).List(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) WatchClusterPolicyReports() (watch.Interface, error) {
	return k.client.Resource(ClusterPolicyReports).Watch(context.Background(), metav1.ListOptions{})
}

func (k *k8sPolicyReportAdapter) WatchPolicyReports() (watch.Interface, error) {
	return k.client.Resource(PolicyReports).Watch(context.Background(), metav1.ListOptions{})
}

func NewPolicyReportAdapter(dynamic dynamic.Interface) PolicyReportAdapter {
	return &k8sPolicyReportAdapter{dynamic}
}
