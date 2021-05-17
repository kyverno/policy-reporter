package kubernetes_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
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
		kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
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
		kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
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
		kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
		kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5)),
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

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))

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

func Test_ResultClient_SkipReportsWithoutResults(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	mapper := NewMapper(k8sCMClient)

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))

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
		"results": []interface{}{},
	}

	var clusterPolicyMap2 = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "clusterpolicy-report",
			"creationTimestamp": "2021-02-23T15:00:00Z",
		},
		"summary": map[string]interface{}{
			"pass":  int64(0),
			"skip":  int64(0),
			"warn":  int64(0),
			"fail":  int64(0),
			"error": int64(0),
		},
		"results": []interface{}{},
	}

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap2})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Result from none empty PolicyReport and ClusterPolicyReport Modify")
	}
}

func Test_ResultClient_SkipCachedResults(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	mapper := NewMapper(k8sCMClient)

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 1, cache.New(cache.DefaultExpiration, time.Minute*5))
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 1, cache.New(cache.DefaultExpiration, time.Minute*5))

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

	var policyMap1 = map[string]interface{}{
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
				"status":  "fail",
				"scored":  true,
				"policy":  "required-label",
				"rule":    "app-label-required",
				"timestamp": map[string]interface{}{
					"seconds": 1614093000,
				},
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
				"status":  "fail",
				"scored":  true,
				"timestamp": map[string]interface{}{
					"seconds": int64(1614093000),
				},
				"policy":    "priority-test",
				"resources": []interface{}{},
			},
		},
	}

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
		"results": []interface{}{},
	}

	var clusterPolicyMap2 = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "clusterpolicy-report",
			"creationTimestamp": "2021-02-23T15:00:00Z",
		},
		"summary": map[string]interface{}{
			"pass":  int64(0),
			"skip":  int64(0),
			"warn":  int64(0),
			"fail":  int64(0),
			"error": int64(0),
		},
		"results": []interface{}{},
	}

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})
	time.Sleep(time.Second * 2)
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	time.Sleep(time.Second * 2)
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap1})
	time.Sleep(time.Second * 2)
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Result from none empty PolicyReport and ClusterPolicyReport Modify")
	}
}

func Test_ResultClient_SkipReportsCleanUpEvents(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	mapper := NewMapper(k8sCMClient)

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))

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

	var policyMap2 = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "policy-report",
			"namespace":         "test",
			"creationTimestamp": "2021-02-23T15:00:00Z",
		},
		"summary": map[string]interface{}{
			"pass":  int64(0),
			"skip":  int64(0),
			"warn":  int64(0),
			"fail":  int64(0),
			"error": int64(0),
		},
		"results": []interface{}{},
	}

	var clusterPolicyMap2 = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "clusterpolicy-report",
			"creationTimestamp": "2021-02-23T15:00:00Z",
		},
		"summary": map[string]interface{}{
			"pass":  int64(0),
			"skip":  int64(0),
			"warn":  int64(0),
			"fail":  int64(0),
			"error": int64(0),
		},
		"results": []interface{}{},
	}

	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})
	fakeAdapter.clusterPolicyWatcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.policyWatcher.Modify(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Results from the initial add events, not from the cleanup modify events")
	}
}

func Test_ResultClient_SkipReportsReconnectEvents(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	fakeAdapter := NewPolicyReportAdapter()

	mapper := NewMapper(k8sCMClient)

	pClient := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))
	cpClient := kubernetes.NewClusterPolicyReportClient(fakeAdapter, report.NewClusterPolicyReportStore(), mapper, time.Now(), 5, cache.New(cache.DefaultExpiration, time.Minute*5))

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
	fakeAdapter.clusterPolicyWatcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.policyWatcher.Add(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Results from the initial add events, not from the restart events")
	}
}
