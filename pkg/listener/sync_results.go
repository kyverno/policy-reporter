package listener

import (
	"context"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendSyncResults = "send_sync_results_listener"

func NewSendSyncResultsListener(targets *target.Collection) report.SyncResultsListener {
	ready := make(chan bool)
	ok := false
	go func() {
		ok = targets.Reset(context.Background())
		if ok {
			close(ready)
		}
	}()

	return func(rep v1alpha2.ReportInterface) {
		clients := targets.SyncClients()
		if len(clients) == 0 {
			return
		}

		if !ok {
			<-ready
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface) {
				defer wg.Done()

				filtered := helper.Filter(re.GetResults(), func(result v1alpha2.PolicyReportResult) bool {
					return target.Validate(re, result)
				})

				resultsToSend := []payload.Payload{}
				for _, r := range filtered {
					resultsToSend = append(resultsToSend, &payload.PolicyReportResultPayload{Result: r})
				}

				target.BatchSend(re, resultsToSend)
			}(t, rep)
		}

		wg.Wait()
	}
}
