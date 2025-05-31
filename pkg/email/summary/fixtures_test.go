package summary_test

import (
	"go.uber.org/zap"
	"openreports.io/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/validate"
	"openreports.io/pkg/client/clientset/versioned/fake"
)

var (
	filter = email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{})
	logger = zap.NewNop()
)

func NewFakeClient() (v1alpha1.OpenreportsV1alpha1Interface, v1alpha1.ReportInterface, v1alpha1.ClusterReportInterface) {
	client := fake.NewSimpleClientset().OpenreportsV1alpha1()

	return client, client.Reports("test"), client.ClusterReports()
}
