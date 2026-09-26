package emailreport_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/emailreport"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    policyreporterv1alpha1.EmailReportSpec
		wantErr string
	}{
		{
			name: "summary",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeSummary,
				Schedule: "0 8 * * *",
				To:       []string{"team@example.test"},
			},
		},
		{
			name: "violations",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeViolations,
				Schedule: "*/5 * * * *",
				To:       []string{"team@example.test"},
			},
		},
		{
			name: "unknown type",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     "details",
				Schedule: "0 8 * * *",
				To:       []string{"team@example.test"},
			},
			wantErr: "unsupported report type",
		},
		{
			name: "invalid schedule",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeSummary,
				Schedule: "definitely-not-cron",
				To:       []string{"team@example.test"},
			},
			wantErr: "invalid schedule",
		},
		{
			name: "no recipients",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeSummary,
				Schedule: "0 8 * * *",
			},
			wantErr: "at least one recipient",
		},
		{
			name: "blank recipient",
			spec: policyreporterv1alpha1.EmailReportSpec{
				Type:     policyreporterv1alpha1.EmailReportTypeSummary,
				Schedule: "0 8 * * *",
				To:       []string{"  "},
			},
			wantErr: "recipient must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := emailreport.Validate(tt.spec)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
