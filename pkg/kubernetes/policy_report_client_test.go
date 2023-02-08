package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

var filter = report.NewFilter(false, validate.RuleSets{})

func Test_PolicyReportWatcher(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	wg := sync.WaitGroup{}
	wg.Add(3)

	store := newStore(3)
	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(event report.LifecycleEvent) {
		store.Add(event)
		wg.Done()
	})

	restClient, polrClient, _ := NewFakeCilent()

	queue := kubernetes.NewQueue(
		cache.New(),
		kubernetes.NewDebouncer(0, publisher),
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test-queue"),
		restClient.Wgpolicyk8sV1alpha2(),
	)

	kclient, rclient, _ := NewFakeMetaCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, queue)

	go func() {
		err := client.Run(stop)
		if err != nil {
			t.Error(err)
		}
	}()

	polrClient.Create(ctx, fixtures.DefaultPolicyReport, metav1.CreateOptions{})

	rclient.CreateFake(fixtures.DefaultMeta, metav1.CreateOptions{})
	time.Sleep(1 * time.Second)

	rclient.UpdateFake(fixtures.DefaultMeta, metav1.UpdateOptions{})
	time.Sleep(1 * time.Second)

	polrClient.Delete(ctx, fixtures.DefaultPolicyReport.Name, metav1.DeleteOptions{})
	rclient.Delete(ctx, fixtures.DefaultMeta.Name, metav1.DeleteOptions{})

	wg.Wait()

	if len(store.List()) != 3 {
		t.Error("Should receive the Added, Updated and Deleted Event")
	}
}

func Test_ClusterPolicyReportWatcher(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)
	wg := sync.WaitGroup{}
	wg.Add(3)

	store := newStore(3)
	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(event report.LifecycleEvent) {
		store.Add(event)
		wg.Done()
	})

	restClient, _, polrClient := NewFakeCilent()

	queue := kubernetes.NewQueue(
		cache.New(),
		kubernetes.NewDebouncer(0, publisher),
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test-queue"),
		restClient.Wgpolicyk8sV1alpha2(),
	)

	kclient, _, rclient := NewFakeMetaCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, queue)

	go func() {
		err := client.Run(stop)
		if err != nil {
			t.Error(err)
		}
	}()

	polrClient.Create(ctx, fixtures.ClusterPolicyReport, metav1.CreateOptions{})

	rclient.CreateFake(fixtures.DefaultClusterMeta, metav1.CreateOptions{})
	time.Sleep(1 * time.Second)

	rclient.UpdateFake(fixtures.DefaultClusterMeta, metav1.UpdateOptions{})
	time.Sleep(1 * time.Second)

	polrClient.Delete(ctx, fixtures.ClusterPolicyReport.Name, metav1.DeleteOptions{})
	rclient.Delete(ctx, fixtures.ClusterPolicyReport.Name, metav1.DeleteOptions{})

	wg.Wait()

	if len(store.List()) != 3 {
		t.Error("Should receive the Added, Updated and Deleted Event")
	}
}

func Test_HasSynced(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	restClient, _, _ := NewFakeCilent()

	queue := kubernetes.NewQueue(
		cache.New(),
		kubernetes.NewDebouncer(0, report.NewEventPublisher()),
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test-queue"),
		restClient.Wgpolicyk8sV1alpha2(),
	)

	kclient, _, _ := NewFakeMetaCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, queue)

	err := client.Sync(stop)
	if err != nil {
		t.Error(err)
	}

	if client.HasSynced() != true {
		t.Errorf("Should synced")
	}
}
