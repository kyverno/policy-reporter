package listener

import (
	"context"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
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

	return func(rep openreports.ReportInterface) {
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
			go func(target target.Client, re openreports.ReportInterface) {
				defer wg.Done()

				filtered := helper.Filter(re.GetResults(), func(result *openreports.ORResultAdapter) bool {
					return target.Validate(re, result)
				})

				target.BatchSend(re, filtered)
			}(t, rep)
		}

		wg.Wait()
	}
}
