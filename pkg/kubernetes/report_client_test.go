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
	testcore "k8s.io/client-go/testing"
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

func Test_FetchPolicyReports(t *testing.T) {
	client, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	watcher := watch.NewFake()
	// Sync error don't break the process
	client.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watcher, errors.New("")))

	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	fakeAdapter.policies = append(fakeAdapter.policies, unstructured.Unstructured{Object: policyMap})

	policies, err := policyClient.FetchPolicyReports()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(policies) != 1 {
		t.Fatal("Expected one Policy")
	}

	expected := kubernetes.NewMapper(configMap.Data).MapPolicyReport(policyMap)
	policy := policies[0]

	if policy.Name != expected.Name {
		t.Errorf("Expected Policy Name %s", expected.Name)
	}
}

func Test_FetchPolicyReportsError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.policyError = errors.New("")

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	_, err = policyClient.FetchPolicyReports()
	if err == nil {
		t.Error("Configured Error should be returned")
	}
}

func Test_FetchClusterPolicyReports(t *testing.T) {
	client, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	watcher := watch.NewFake()
	client.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watcher, nil))

	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	watcher.Modify(configMap)

	fakeAdapter.clusterPolicies = append(fakeAdapter.clusterPolicies, unstructured.Unstructured{Object: clusterPolicyMap})

	policies, err := policyClient.FetchClusterPolicyReports()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(policies) != 1 {
		t.Fatal("Expected one Policy")
	}

	expected := kubernetes.NewMapper(configMap.Data).MapClusterPolicyReport(clusterPolicyMap)
	policy := policies[0]

	if policy.Name != expected.Name {
		t.Errorf("Expected Policy Name %s", expected.Name)
	}
}

func Test_FetchClusterPolicyReportsError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.clusterPolicyError = errors.New("")

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	_, err = policyClient.FetchClusterPolicyReports()
	if err == nil {
		t.Error("Configured Error should be returned")
	}
}

func Test_FetchClusterPolicyResults(t *testing.T) {
	client, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	watcher := watch.NewFake()
	client.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watcher, nil))

	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	watcher.Modify(configMap)

	fakeAdapter.clusterPolicies = append(fakeAdapter.clusterPolicies, unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.policies = append(fakeAdapter.policies, unstructured.Unstructured{Object: policyMap})

	results, err := policyClient.FetchPolicyReportResults()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 Results, got %d", len(results))
	}
}
func Test_FetchPolicyResultsError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.clusterPolicyError = errors.New("")

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	_, err = policyClient.FetchPolicyReportResults()
	if err == nil {
		t.Error("ClusterPolicyFetch Error should be returned by FetchPolicyReportResults")
	}
}

func Test_Watchers(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	policyClient.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	policyClient.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go policyClient.StartWatchPolicyReports()
	go policyClient.StartWatchClusterPolicyReports()

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 2 Results from the Policy and 1 Result from the ClusterPolicy")
	}
}

var notSkippedPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "policy-report",
		"namespace":         "test",
		"creationTimestamp": time.Now().Add(10 * time.Minute).Format("2006-01-02T15:04:05Z"),
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
			"message":  "message",
			"status":   "fail",
			"scored":   true,
			"policy":   "not-skiped-policy-result",
			"rule":     "app-label-required",
			"category": "test",
			"severity": "low",
			"resources": []interface{}{
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Deployment",
					"name":       "nginx",
					"namespace":  "test",
					"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
				},
			},
		},
	},
}

var notSkippedClusterPolicyMap = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "clusterpolicy-report",
		"creationTimestamp": time.Now().Add(10 * time.Minute).Format("2006-01-02T15:04:05Z"),
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
			"message":  "message",
			"status":   "fail",
			"scored":   true,
			"policy":   "not-skiped-cluster-policy-result",
			"rule":     "app-label-required",
			"category": "test",
			"severity": "low",
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

func Test_SkipExisting(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	policyClient.RegisterPolicyResultWatcher(true)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]report.Result, 0, 2)

	policyClient.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go policyClient.StartWatchPolicyReports()
	go policyClient.StartWatchClusterPolicyReports()

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: notSkippedPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: notSkippedClusterPolicyMap})

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive 2 not skipped Result form notSkippedPolicyMap and notSkippedClusterPolicyMap")
	}

	if results[0].Policy != "not-skiped-policy-result" && results[0].Policy != "not-skiped-cluster-policy-result" {
		t.Error("Should be one of 'not-skiped-policy-result', 'not-skiped-cluster-policy-result'")
	}
	if results[1].Policy != "not-skiped-policy-result" && results[1].Policy != "not-skiped-cluster-policy-result" {
		t.Error("Should be one of 'not-skiped-policy-result', 'not-skiped-cluster-policy-result'")
	}
}

func Test_WatcherError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.policyError = errors.New("")
	fakeAdapter.clusterPolicyError = errors.New("")

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	policyClient.RegisterPolicyResultWatcher(false)

	err = policyClient.StartWatchPolicyReports()
	if err == nil {
		t.Error("Shoud stop execution when error is returned")
	}

	err = policyClient.StartWatchClusterPolicyReports()
	if err == nil {
		t.Error("Shoud stop execution when error is returned")
	}
}

func Test_WatchDeleteEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	policyClient.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	policyClient.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go policyClient.StartWatchPolicyReports()
	go policyClient.StartWatchClusterPolicyReports()

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.policyWatcher.Delete(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.clusterPolicyWatcher.Delete(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive initial 3 and no result from deletion")
	}
}

func Test_WatchModifiedEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	policyClient, err := kubernetes.NewPolicyReportClient(
		context.Background(),
		fakeAdapter,
		kubernetes.NewConfigMapAdapter(k8sCMClient),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	policyClient.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(5)

	results := make([]report.Result, 0, 5)
	policyClient.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go policyClient.StartWatchPolicyReports()
	go policyClient.StartWatchClusterPolicyReports()

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	var policyMap2 = map[string]interface{}{
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
				"message":  "message",
				"status":   "fail",
				"scored":   true,
				"policy":   "required-label",
				"rule":     "app-label-required",
				"category": "test",
				"severity": "low",
				"resources": []interface{}{
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Deployment",
						"name":       "nginx",
						"namespace":  "test",
						"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
					},
				},
			},
			map[string]interface{}{
				"message":   "message 2",
				"status":    "fail",
				"scored":    true,
				"policy":    "priority-test",
				"resources": []interface{}{},
			},
			map[string]interface{}{
				"message":   "message 3",
				"status":    "pass",
				"scored":    true,
				"policy":    "priority-test",
				"resources": []interface{}{},
			},
		},
	}
	clusterPolicyMap2 := map[string]interface{}{
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
				"message":  "message",
				"status":   "fail",
				"scored":   true,
				"policy":   "required-label",
				"rule":     "app-label-required",
				"category": "test",
				"severity": "low",
				"resources": []interface{}{
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"name":       "policy-reporter",
						"uid":        "dfd57c50-f30c-4729-b63f-b1954d8988d1",
					},
				},
			},
			map[string]interface{}{
				"message":  "message",
				"status":   "fail",
				"scored":   true,
				"policy":   "required-label",
				"rule":     "app-label-required",
				"category": "test",
				"severity": "low",
				"resources": []interface{}{
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Namespace",
						"name":       "policy-reporter",
						"uid":        "dfd57c50-f30c-4729-b63f-b1754d7988d1",
					},
				},
			},
		},
	}

	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})

	wg.Wait()

	if len(results) != 5 {
		t.Error("Should receive initial 3 and 2 modification")
	}
}
