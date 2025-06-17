package listener

import (
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendResults = "send_results_listener"

func NewSendResultListener(targets *target.Collection) report.PolicyReportResultListener {
	return func(rep openreports.ReportInterface, r openreports.ORResultAdapter, e bool) {
		clients := targets.SingleSendClients()
		if len(clients) == 0 {
			return
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re openreports.ReportInterface, result openreports.ORResultAdapter, preExisted bool) {
				defer wg.Done()

				if !result.HasResource() && re.GetScope() != nil {
					result.Subjects = []corev1.ObjectReference{*re.GetScope()}
				}

				if (preExisted && target.SkipExistingOnStartup()) || !target.Validate(re, result) {
					return
				}

				target.Send(result)
			}(t, rep, r, e)
		}

		wg.Wait()
	}
}
