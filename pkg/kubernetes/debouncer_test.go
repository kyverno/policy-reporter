package kubernetes_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
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
			Type:            report.Added,
			NewPolicyReport: policyReportCRD,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:            report.Updated,
			NewPolicyReport: minPolicyReportCRD,
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:            report.Updated,
			NewPolicyReport: policyReportCRD,
		})

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to skip the empty modify event")
		}
	})
}
