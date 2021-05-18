package kubernetes_test

import (
	"context"
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var (
	policyReportAlphaV1 = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha1",
		Resource: "clusterpolicyreports",
	}
)

func newUnstructured(apiVersion, kind, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}

func NewFakeClient(items ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClient(runtime.NewScheme(), items...)
}

func Test_WatchPolicyReports(t *testing.T) {
	dynamic := NewFakeClient()
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	client := kubernetes.NewPolicyReportAdapter(dynamic, NewMapper(k8sCMClient))

	_, err := client.WatchPolicyReports()
	if err != nil {
		t.Error("Unexpected WatchError")
	}
}
