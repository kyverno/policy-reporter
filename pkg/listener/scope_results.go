package listener

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendScopeResults = "send_scope_results_listener"

func NewSendScopeResultsListener(targets *target.Collection) report.ScopeResultsListener {
	return func(rep v1alpha2.ReportInterface, r []v1alpha2.PolicyReportResult) {
		clients := targets.BatchSendClients()

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
				defer wg.Done()

				filtered := helper.Filter(results, func(result v1alpha2.PolicyReportResult) bool {
					return target.Validate(re, result)
				})

				if len(filtered) == 0 {
					return
				}

				var resultsToSend []v1alpha2.PolicyReportResult
				existing := target.Cache().GetResults(re.GetID())

				for _, r := range filtered {
					preExisted := time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos)).Before(target.CreationTimestamp())
					if preExisted && target.SkipExistingOnStartup() {
						continue
					}
					if helper.Contains(r.GetID(), existing) {
						continue
					}

					resultsToSend = append(resultsToSend, r)
				}
				target.Cache().AddReport(re)
				if len(resultsToSend) > 0 {
					target.BatchSend(re, resultsToSend)
				}
			}(t, rep, r)
		}

		wg.Wait()
	}
}
