package kubernetes_test

import (
	"context"
	"fmt"
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

	kclient, rclient, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), 100*time.Millisecond, filter)

	group := client.WatchPolicyReports(ctx)
	store := newStore(3)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		reportID := <-group.ChannelAdded()
		eventChan, err := group.Listen(reportID)
		if err != nil {
			t.Error(err)
		}

		for event := range eventChan {
			fmt.Printf("%v\n", event.Type)
			store.Add(event)
			wg.Done()
		}
	}()

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

	kclient, _, rclient := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), 100*time.Millisecond, filter)

	group := client.WatchPolicyReports(ctx)
	store := newStore(3)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		reportID := <-group.ChannelAdded()
		eventChan, err := group.Listen(reportID)
		if err != nil {
			t.Error(err)
		}

		for event := range eventChan {
			store.Add(event)
			wg.Done()
		}
	}()

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

func Test_GetFoundResources(t *testing.T) {
	ctx := context.Background()

	kclient, _, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), 100*time.Millisecond, filter)

	client.WatchPolicyReports(ctx)

	time.Sleep(1 * time.Second)

	if len(client.GetFoundResources()) != 2 {
		t.Errorf("Should find PolicyReport and ClusterPolicyReport Resource")
	}
}

func Test_GetFoundResourcesWihDisabledClusterReports(t *testing.T) {
	ctx := context.Background()

	kclient, _, _ := NewFakeCilent()
	client := kubernetes.NewPolicyReportClient(kclient, NewMapper(), 100*time.Millisecond, report.NewFilter(true, make([]string, 0), make([]string, 0)))

	client.WatchPolicyReports(ctx)

	time.Sleep(1 * time.Second)

	if len(client.GetFoundResources()) != 1 {
		t.Errorf("Should find only PolicyReport Resource")
	}
}
