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
		debouncer := kubernetes.NewDebouncer(200 * time.Millisecond)

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			for event := range debouncer.ReportChan() {
				wg.Done()
				if len(event.NewPolicyReport.Results) == 0 {
					t.Error("Expected to skip the empty modify event")
				}
			}
		}()

		debouncer.Add(report.LifecycleEvent{
			Type:            report.Added,
			NewPolicyReport: mapper.MapPolicyReport(policyMap),
		})

		debouncer.Add(report.LifecycleEvent{
			Type:            report.Updated,
			NewPolicyReport: mapper.MapPolicyReport(minPolicyMap),
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:            report.Updated,
			NewPolicyReport: mapper.MapPolicyReport(policyMap),
		})

		wg.Wait()

		debouncer.Close()
	})
}
