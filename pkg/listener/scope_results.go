package listener

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendScopeResults = "send_scope_results_listener"

func NewSendScopeResultsListener(targets *target.Collection) report.ScopeResultsListener {
	return func(rep openreports.ReportInterface, r []openreports.ORResultAdapter, e bool) {
		clients := targets.BatchSendClients()
		if len(clients) == 0 {
			return
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re openreports.ReportInterface, results []openreports.ORResultAdapter, preExisted bool) {
				defer wg.Done()

				filtered := helper.Filter(results, func(result openreports.ORResultAdapter) bool {
					return target.Validate(re, result)
				})

				if len(filtered) == 0 || preExisted && target.SkipExistingOnStartup() {
					return
				}

				target.BatchSend(re, filtered)
			}(t, rep, r, e)
		}

		wg.Wait()
	}
}
