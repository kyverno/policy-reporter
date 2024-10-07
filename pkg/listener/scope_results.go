package listener

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendScopeResults = "send_scope_results_listener"

func NewSendScopeResultsListener(targets *target.Collection) report.ScopeResultsListener {
	return func(rep v1alpha2.ReportInterface, r []v1alpha2.PolicyReportResult, e bool) {
		clients := targets.BatchSendClients()

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult, preExisted bool) {
				defer wg.Done()

				filtered := helper.Filter(results, func(result v1alpha2.PolicyReportResult) bool {
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
