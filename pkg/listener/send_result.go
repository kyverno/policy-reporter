package listener

import (
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendResults = "send_results_listener"

func NewSendResultListener(clients []target.Client, mapper report.Mapper) report.PolicyReportResultListener {
	return func(rep v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, e bool) {
		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult, preExisted bool) {
				defer wg.Done()

				if result.Result == v1alpha2.StatusFail {
					result.Priority = mapper.ResolvePriority(result.Policy, result.Severity)
				}

				if !result.HasResource() && re.GetScope() != nil {
					result.Resources = []corev1.ObjectReference{*re.GetScope()}
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
