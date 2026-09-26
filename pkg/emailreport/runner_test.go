package emailreport_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	crdfake "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/emailreport"
)

func TestRunnerUsesCurrentEmailReport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	report := newEmailReport("team-a", "daily", "0 8 * * *")
	report.Spec.Type = policyreporterv1alpha1.EmailReportTypeViolations
	client := crdfake.NewSimpleClientset(report)

	var delivered *policyreporterv1alpha1.EmailReport
	runner := emailreport.NewRunner(client.PolicyreporterV1alpha1(), func(_ context.Context, current *policyreporterv1alpha1.EmailReport) error {
		delivered = current.DeepCopy()
		return nil
	})

	require.NoError(t, runner.Run(ctx, "team-a", "daily", report.UID))
	require.NotNil(t, delivered)
	assert.Equal(t, policyreporterv1alpha1.EmailReportTypeViolations, delivered.Spec.Type)
	assert.Equal(t, []string{"team-a@example.test"}, delivered.Spec.To)
}

func TestRunnerRejectsStaleUID(t *testing.T) {
	t.Parallel()
	report := newEmailReport("team-a", "daily", "0 8 * * *")
	client := crdfake.NewSimpleClientset(report)
	runner := emailreport.NewRunner(client.PolicyreporterV1alpha1(), func(context.Context, *policyreporterv1alpha1.EmailReport) error {
		t.Fatal("delivery must not run for a stale UID")
		return nil
	})

	err := runner.Run(context.Background(), "team-a", "daily", types.UID("stale-uid"))
	assert.ErrorContains(t, err, "UID does not match")
}

func TestRunnerReturnsValidationAndDeliveryErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("invalid resource", func(t *testing.T) {
		report := newEmailReport("team-a", "daily", "invalid")
		client := crdfake.NewSimpleClientset(report)
		runner := emailreport.NewRunner(client.PolicyreporterV1alpha1(), func(context.Context, *policyreporterv1alpha1.EmailReport) error {
			t.Fatal("delivery must not run for an invalid resource")
			return nil
		})

		assert.ErrorContains(t, runner.Run(ctx, "team-a", "daily", report.UID), "invalid schedule")
	})

	t.Run("delivery failure", func(t *testing.T) {
		report := &policyreporterv1alpha1.EmailReport{
			ObjectMeta: metav1.ObjectMeta{Namespace: "team-a", Name: "daily", UID: types.UID("uid")},
			Spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeSummary,
				Schedule: "0 8 * * *",
				To:       []string{"team-a@example.test"},
			},
		}
		client := crdfake.NewSimpleClientset(report)
		runner := emailreport.NewRunner(client.PolicyreporterV1alpha1(), func(context.Context, *policyreporterv1alpha1.EmailReport) error {
			return errors.New("smtp unavailable")
		})

		assert.ErrorContains(t, runner.Run(ctx, "team-a", "daily", report.UID), "smtp unavailable")
	})
}
