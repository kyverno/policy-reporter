package violations_test

import (
	"github.com/openreports/reports-api/pkg/client/clientset/versioned/fake"
	"github.com/openreports/reports-api/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

var (
	filter = email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{})
	logger = zap.NewNop()
)

func NewFakeClient() (*fake.Clientset, v1alpha1.ReportInterface, v1alpha1.ClusterReportInterface) {
	client := fake.NewSimpleClientset()

	return client, client.OpenreportsV1alpha1().Reports("test"), client.OpenreportsV1alpha1().ClusterReports()
}
