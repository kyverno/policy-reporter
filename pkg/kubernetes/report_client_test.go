package kubernetes_test

import (
	"context"
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
	List    []report.PolicyReport
	Watcher *watch.FakeWatcher
	Error   error
	mapper  kubernetes.Mapper
}

func (f *fakeClient) ListPolicyReports() ([]report.PolicyReport, error) {
	return f.List, f.Error
}

func (f *fakeClient) ListClusterPolicyReports() ([]report.PolicyReport, error) {
	return f.List, f.Error
}

func (f *fakeClient) WatchPolicyReports() (chan kubernetes.WatchEvent, error) {
	channel := make(chan kubernetes.WatchEvent)

	go func() {
		for result := range f.Watcher.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				report := f.mapper.MapPolicyReport(item.Object)
				channel <- kubernetes.WatchEvent{report, result.Type}
			}
		}
	}()

	return channel, f.Error
}

func NewPolicyReportAdapter(mapper kubernetes.Mapper) *fakeClient {
	return &fakeClient{
		List:    make([]report.PolicyReport, 0),
		Watcher: watch.NewFake(),
		mapper:  mapper,
	}
}

func NewMapper(k8sCMClient v1.ConfigMapInterface) kubernetes.Mapper {
	return kubernetes.NewMapper(make(map[string]string), kubernetes.NewConfigMapAdapter(k8sCMClient))
}

func Test_ResultClient_RegisterPolicyResultWatcher(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Result from all PolicyReports")
	}
}

func Test_ResultClient_SkipCachedResults(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

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

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap1})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Result from none empty PolicyReport and ClusterPolicyReport Modify")
	}
}

func Test_ResultClient_SkipReportsCleanUpEvents(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

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

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap2})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap2})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Results from the initial add events, not from the cleanup modify events")
	}
}

func Test_ResultClient_SkipReportsReconnectEvents(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(3)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})
	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: clusterPolicyMap})

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive 3 Results from the initial add events, not from the restart events")
	}
}
