package listener

import (
	"sync"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func NewSendResultListener(clients []target.Client) report.PolicyReportResultListener {
	return func(r *report.Result, e bool) {
		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, result *report.Result, preExisted bool) {
				defer wg.Done()

				if (preExisted && target.SkipExistingOnStartup()) || !target.Validate(result) {
					return
				}

				target.Send(result)
			}(t, r, e)
		}

		wg.Wait()
	}
}
