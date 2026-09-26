package emailreport

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	typedclient "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/targetconfig/v1alpha1"
)

type DeliveryFunc func(context.Context, *policyreporterv1alpha1.EmailReport) error

type Runner struct {
	reports  typedclient.PolicyreporterV1alpha1Interface
	delivery DeliveryFunc
}

func NewRunner(reports typedclient.PolicyreporterV1alpha1Interface, delivery DeliveryFunc) *Runner {
	return &Runner{reports: reports, delivery: delivery}
}

func (r *Runner) Run(ctx context.Context, namespace, name string, uid types.UID) error {
	report, err := r.reports.EmailReports(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if report.UID != uid {
		return fmt.Errorf("EmailReport UID does not match current resource")
	}
	if err := Validate(report.Spec); err != nil {
		return err
	}
	return r.delivery(ctx, report)
}
