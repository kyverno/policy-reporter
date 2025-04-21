package listener

import (
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendResults = "send_results_listener"

func NewSendResultListener(targets *target.Collection) report.PolicyReportResultListener {
	return func(rep v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, e bool) {
		clients := targets.SingleSendClients()
		if len(clients) == 0 {
			return
		}

		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult, preExisted bool) {
				defer wg.Done()

				if !result.HasResource() && re.GetScope() != nil {
					result.Resources = []corev1.ObjectReference{*re.GetScope()}
				}

				if (preExisted && target.SkipExistingOnStartup()) || !target.Validate(re, result) {
					return
				}

				target.Send(&payload.PolicyReportResultPayload{Result: result})
			}(t, rep, r, e)
		}

		wg.Wait()
	}
}
