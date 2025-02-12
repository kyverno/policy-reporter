package listener

import (
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const SendResults = "send_results_listener"

func NewSendResultListener(targets *target.Collection) report.PolicyReportResultListener {
	return func(rep v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
		clients := targets.SingleSendClients()
		wg := &sync.WaitGroup{}
		wg.Add(len(clients))

		for _, t := range clients {
			go func(target target.Client, re v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) {
				defer wg.Done()

				if !result.HasResource() && re.GetScope() != nil {
					result.Resources = []corev1.ObjectReference{*re.GetScope()}
				}

				existing := target.Cache().GetResults(re.GetID())
				for _, r := range re.GetResults() {
					preExisted := time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos)).Local().Before(target.CreationTimestamp())
					if (preExisted && target.SkipExistingOnStartup()) || !target.Validate(re, result) {
						continue
					}

					if helper.Contains(r.GetID(), existing) {
						continue
					}
					target.Send(result)
				}

				target.Cache().AddReport(re)
			}(t, rep, r)
		}

		wg.Wait()
	}
}
