package kubernetes_test

import (
	"context"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(items ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClient(runtime.NewScheme(), items...)
}

func Test_WatchPolicyReports(t *testing.T) {
	ctx := context.Background()
	dynamic := NewFakeClient()
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(ctx, configMap, metav1.CreateOptions{})

	client := kubernetes.NewPolicyReportAdapter(dynamic, NewMapper(k8sCMClient))

	_, err := client.WatchPolicyReports(ctx)
	if err != nil {
		t.Error("Unexpected WatchError")
	}
}
