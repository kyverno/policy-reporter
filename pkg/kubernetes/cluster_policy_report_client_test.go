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

func Test_FetchClusterPolicyReports(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	fakeAdapter.clusterPolicies = append(fakeAdapter.clusterPolicies, unstructured.Unstructured{Object: clusterPolicyMap})

	policies, err := client.FetchClusterPolicyReports()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(policies) != 1 {
		t.Fatal("Expected one Policy")
	}

	expected := kubernetes.NewMapper(configMap.Data, nil).MapClusterPolicyReport(clusterPolicyMap)
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

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	_, err := client.FetchClusterPolicyReports()
	if err == nil {
		t.Error("Configured Error should be returned")
	}
}

func Test_FetchClusterPolicyResults(t *testing.T) {
	fakeClient, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	watcher := watch.NewFake()
	fakeClient.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watcher, nil))

	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	fakeAdapter.clusterPolicies = append(fakeAdapter.clusterPolicies, unstructured.Unstructured{Object: clusterPolicyMap})

	results, err := client.FetchPolicyResults()
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 Results, got %d", len(results))
	}
}
func Test_FetchClusterPolicyResultsError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.clusterPolicyError = errors.New("")

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	_, err := client.FetchPolicyResults()
	if err == nil {
		t.Error("ClusterPolicyFetch Error should be returned by FetchPolicyResults")
	}
}

func Test_ClusterPolicyWatcher(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(1)

	results := make([]report.Result, 0, 1)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()

	if len(results) != 1 {
		t.Error("Should receive 1 Result from the ClusterPolicy")
	}
}

func Test_ClusterPolicyWatcherTwice(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	go client.StartWatching()

	time.Sleep(10 * time.Millisecond)

	err := client.StartWatching()
	if err == nil {
		t.Error("Second StartWatching call should return immediately with error")
	}
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

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(true)

	wg := sync.WaitGroup{}
	wg.Add(1)

	results := make([]report.Result, 0, 1)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: notSkippedClusterPolicyMap})

	wg.Wait()

	if len(results) != 1 {
		t.Error("Should receive one not skipped Result form notSkippedClusterPolicyMap")
	}

	if results[0].Policy != "not-skiped-cluster-policy-result" {
		t.Error("Should be 'not-skiped-cluster-policy-result'")
	}
}

func Test_WatcherError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()
	fakeAdapter.clusterPolicyError = errors.New("")

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	err := client.StartWatching()
	if err == nil {
		t.Error("Shoud stop execution when error is returned")
	}
}

func Test_WatchDeleteEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(1)

	results := make([]report.Result, 0, 1)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Delete(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()

	if len(results) != 1 {
		t.Error("Should receive initial 1 and no result from deletion")
	}
}

func Test_WatchDelayEvents(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(2)

	client.RegisterCallback(func(e watch.EventType, r report.ClusterPolicyReport, o report.ClusterPolicyReport) {
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: minClusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Delete(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()
}

func Test_WatchDelayEventsWithoutClearEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	client.RegisterCallback(func(e watch.EventType, r report.ClusterPolicyReport, o report.ClusterPolicyReport) {
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Delete(&unstructured.Unstructured{Object: clusterPolicyMap})

	wg.Wait()
}

func Test_WatchModifiedEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	client := kubernetes.NewClusterPolicyReportClient(
		fakeAdapter,
		report.NewClusterPolicyReportStore(),
		NewMapper(k8sCMClient),
		time.Now(),
	)

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]report.Result, 0, 2)
	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

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
				"severity": "high",
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

	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive initial 1 and 1 modification")
	}
}
