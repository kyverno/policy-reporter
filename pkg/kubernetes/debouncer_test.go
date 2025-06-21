package kubernetes_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_Debouncer(t *testing.T) {
	t.Run("Skip Empty Update", func(t *testing.T) {
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(event report.LifecycleEvent) {
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
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(event report.LifecycleEvent) {
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

	t.Run("Skip Empty Update when delete event follows directly", func(t *testing.T) {
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(event report.LifecycleEvent) {
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
