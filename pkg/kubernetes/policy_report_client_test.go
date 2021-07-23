package kubernetes_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_PolicyWatcher(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]report.Result, 0, 3)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive 2 Results from the Policy")
	}
}

func Test_PolicyWatcherTwice(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	go client.StartWatching()

	time.Sleep(10 * time.Millisecond)

	err := client.StartWatching()
	if err == nil {
		t.Error("Second StartWatching call should return immediately with error")
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

func Test_PolicySkipExisting(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(true)

	wg := sync.WaitGroup{}
	wg.Add(1)

	results := make([]report.Result, 0, 1)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: notSkippedPolicyMap})

	wg.Wait()

	if len(results) != 1 {
		t.Error("Should receive one not skipped Result form notSkippedPolicyMap")
	}

	if results[0].Policy != "not-skiped-policy-result" {
		t.Error("Should be 'not-skiped-policy-result'")
	}
}

func Test_PolicyWatcherError(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	fakeAdapter.Error = errors.New("")

	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	err := client.StartWatching()
	if err == nil {
		t.Error("Shoud stop execution when error is returned")
	}
}

func Test_PolicyWatchDeleteEvent(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(2)

	results := make([]report.Result, 0, 2)

	client.RegisterPolicyResultCallback(func(r report.Result, b bool) {
		results = append(results, r)
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Delete(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()

	if len(results) != 2 {
		t.Error("Should receive initial 2 and no result from deletion")
	}
}

func Test_PolicyWatchModifiedEvent(t *testing.T) {
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

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})

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
				"severity": "medium",
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

	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap2})

	wg.Wait()

	if len(results) != 3 {
		t.Error("Should receive initial 2 and 1 modification")
	}
}

func Test_PolicyDelayReset(t *testing.T) {
	_, k8sCMClient := newFakeAPI()
	k8sCMClient.Create(context.Background(), configMap, metav1.CreateOptions{})

	fakeAdapter := NewPolicyReportAdapter(NewMapper(k8sCMClient))
	client := kubernetes.NewPolicyReportClient(fakeAdapter, report.NewPolicyReportStore(), time.Now(), cache.New(cache.DefaultExpiration, time.Minute*5))

	client.RegisterPolicyResultWatcher(false)

	wg := sync.WaitGroup{}
	wg.Add(2)

	client.RegisterCallback(func(e watch.EventType, r report.PolicyReport, o report.PolicyReport) {
		wg.Done()
	})

	go client.StartWatching()

	fakeAdapter.Watcher.Add(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: minPolicyMap})
	fakeAdapter.Watcher.Modify(&unstructured.Unstructured{Object: policyMap})
	fakeAdapter.Watcher.Delete(&unstructured.Unstructured{Object: policyMap})

	wg.Wait()
}
