package listener

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendResults = "send_results_listener"

func NewSendResultListener(clients []target.Client) report.PolicyReportResultListener {
	return func(rep report.PolicyReport, r report.Result, e bool) {
		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re report.PolicyReport, result report.Result, preExisted bool) {
				defer wg.Done()

				if (preExisted && target.SkipExistingOnStartup()) || !target.Validate(re, result) {
					return
				}

				target.Send(result)
			}(t, rep, r, e)
		}

		wg.Wait()
	}
}
