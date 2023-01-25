package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	kclient, rclient, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, publisher)

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	rclient.Create(ctx, fixtures.DefaultPolicyReport, metav1.CreateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Update(ctx, fixtures.DefaultPolicyReport, metav1.UpdateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Delete(ctx, fixtures.DefaultPolicyReport.Name, metav1.DeleteOptions{})

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

	kclient, _, rclient := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, publisher)

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	rclient.Create(ctx, fixtures.ClusterPolicyReport, metav1.CreateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Update(ctx, fixtures.ClusterPolicyReport, metav1.UpdateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Delete(ctx, fixtures.ClusterPolicyReport.Name, metav1.DeleteOptions{})

	wg.Wait()

	if len(store.List()) != 3 {
		t.Error("Should receive the Added, Updated and Deleted Event")
	}
}

func Test_HasSynced(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	kclient, _, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, filter, report.NewEventPublisher())

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	if client.HasSynced() != true {
		t.Errorf("Should synced")
	}
}
