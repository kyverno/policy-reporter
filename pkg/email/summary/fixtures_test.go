package summary_test

import (
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	v1alpha2client "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

var filter = email.NewFilter(validate.RuleSets{}, validate.RuleSets{})

func NewFakeClient() (v1alpha2client.Wgpolicyk8sV1alpha2Interface, v1alpha2client.PolicyReportInterface, v1alpha2client.ClusterPolicyReportInterface) {
	client := fake.NewSimpleClientset().Wgpolicyk8sV1alpha2()

	return client, client.PolicyReports("test"), client.ClusterPolicyReports()
}
