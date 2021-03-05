package kubernetes_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type fakeClient struct {
	policies             []unstructured.Unstructured
	clusterPolicies      []unstructured.Unstructured
	policyWatcher        *watch.FakeWatcher
	clusterPolicyWatcher *watch.FakeWatcher
	policyError          error
	clusterPolicyError   error
}

func (f *fakeClient) ListClusterPolicyReports() (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{
		Items: f.clusterPolicies,
	}, f.clusterPolicyError
}

func (f *fakeClient) ListPolicyReports() (*unstructured.UnstructuredList, error) {
	return &unstructured.UnstructuredList{
		Items: f.policies,
	}, f.policyError
}

func (f *fakeClient) WatchClusterPolicyReports() (watch.Interface, error) {
	return f.clusterPolicyWatcher, f.clusterPolicyError
}

func (f *fakeClient) WatchPolicyReports() (watch.Interface, error) {
	return f.policyWatcher, f.policyError
}

func NewPolicyReportAdapter() *fakeClient {
	return &fakeClient{
		policies:             make([]unstructured.Unstructured, 0),
		clusterPolicies:      make([]unstructured.Unstructured, 0),
		policyWatcher:        watch.NewFake(),
		clusterPolicyWatcher: watch.NewFake(),
	}
}

func NewMapper(k8sCMClient v1.ConfigMapInterface) kubernetes.Mapper {
	return kubernetes.NewMapper(make(map[string]string), kubernetes.NewConfigMapAdapter(k8sCMClient))
}

func Test_ResultClient_FetchPolicyResults(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	mapper := NewMapper(k8sCMClient)

	client := kubernetes.NewPolicyResultClient(
		kubernetes.NewPolicyReportClient(fakeAdapter, mapper, time.Now()),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, mapper, time.Now()),
	)

	fakeAdapter.policies = append(fakeAdapter.policies, unstructured.Unstructured{Object: policyMap})
	fakeAdapter.clusterPolicies = append(fakeAdapter.clusterPolicies, unstructured.Unstructured{Object: clusterPolicyMap})

	results, err := client.FetchPolicyResults()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 Results, got %d", len(results))
	}
}

func Test_ResultClient_FetchPolicyResultsPolicyReportError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.policyError = errors.New("")

	mapper := NewMapper(k8sCMClient)

	client := kubernetes.NewPolicyResultClient(
		kubernetes.NewPolicyReportClient(fakeAdapter, mapper, time.Now()),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, mapper, time.Now()),
	)

	_, err := client.FetchPolicyResults()
	if err == nil {
		t.Error("PolicyFetch Error should be returned by FetchPolicyResults")
	}
}

func Test_ResultClient_FetchPolicyResultsClusterPolicyReportError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.clusterPolicyError = errors.New("")

	mapper := NewMapper(k8sCMClient)

	client := kubernetes.NewPolicyResultClient(
		kubernetes.NewPolicyReportClient(fakeAdapter, mapper, time.Now()),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, mapper, time.Now()),
	)

	_, err := client.FetchPolicyResults()
	if err == nil {
		t.Error("ClusterPolicyFetch Error should be returned by FetchPolicyResults")
	}
}

func Test_ResultClient_RegisterPolicyResultWatcher(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	mapper := NewMapper(k8sCMClient)

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, mapper, time.Now())
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, mapper, time.Now())

	client := kubernetes.NewPolicyResultClient(pClient, cpClient)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go pClient.StartWatching()
	go cpClient.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Result from all PolicyReports")
	}
}
