package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/report"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var filter = report.NewFilter(false, make([]string, 0), make([]string, 0))

func Test_PolicyReportWatcher(t *testing.T) {
	ctx := context.Background()
	stop := make(chan struct{})
	defer close(stop)

	wg := sync.WaitGroup{}
	wg.Add(3)

	store := newStore(3)
	publisher := report.NewEventPublisher()
	publisher.RegisterListener(func(event report.LifecycleEvent) {
		store.Add(event)
		wg.Done()
	})

	kclient, rclient, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), filter, publisher)

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	rclient.Create(ctx, policyReportCRD, metav1.CreateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Update(ctx, policyReportCRD, metav1.UpdateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Delete(ctx, policyReportCRD.Name, metav1.DeleteOptions{})

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
	publisher.RegisterListener(func(event report.LifecycleEvent) {
		store.Add(event)
		wg.Done()
	})

	kclient, _, rclient := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), filter, publisher)

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	rclient.Create(ctx, clusterPolicyReportCRD, metav1.CreateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Update(ctx, clusterPolicyReportCRD, metav1.UpdateOptions{})
	time.Sleep(10 * time.Millisecond)
	rclient.Delete(ctx, clusterPolicyReportCRD.Name, metav1.DeleteOptions{})

	wg.Wait()

	if len(store.List()) != 3 {
		t.Error("Should receive the Added, Updated and Deleted Event")
	}
}

func Test_HasSynced(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	kclient, _, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), filter, report.NewEventPublisher())

	err := client.Run(stop)
	if err != nil {
		t.Fatal(err)
	}

	if client.HasSynced() != true {
		t.Errorf("Should synced")
	}
}
