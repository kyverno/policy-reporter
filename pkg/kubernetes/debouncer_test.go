package kubernetes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_Debouncer(t *testing.T) {
	t.Parallel()
	t.Run("Skip Empty Update", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := kubernetes.NewDebouncer(200*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to skip the empty modify event")
		}
	})

	t.Run("Execute Empty Updates after waiting time", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := kubernetes.NewDebouncer(10*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(5 * time.Millisecond)

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to execute empty update event after wait time")
		}
	})

	t.Run("Coalesce consecutive empty updates into the latest revision", func(t *testing.T) {
		t.Parallel()

		events := make(chan report.LifecycleEvent, 2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			events <- event
		})

		debouncer := kubernetes.NewDebouncer(50*time.Millisecond, publisher)

		first := fixtures.MinPolicyReport.Report.DeepCopy()
		first.ResourceVersion = "1"
		second := first.DeepCopy()
		second.ResourceVersion = "2"

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: first},
		})
		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: second},
		})

		select {
		case event := <-events:
			if version := event.PolicyReport.GetResourceVersion(); version != "2" {
				t.Errorf("Expected the latest revision to be published, got %q", version)
			}
		case <-time.After(time.Second):
			t.Fatal("Expected the debounced update to be published")
		}

		select {
		case event := <-events:
			t.Errorf("Expected a single coalesced event, got an extra revision %q", event.PolicyReport.GetResourceVersion())
		case <-time.After(100 * time.Millisecond):
		}
	})

	t.Run("Skip Empty Update when delete event follows directly", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := kubernetes.NewDebouncer(200*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Deleted,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to skip the empty modify event")
		}
	})
}
