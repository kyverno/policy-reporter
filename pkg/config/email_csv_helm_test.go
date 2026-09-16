package config_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func TestEmailAttachmentHelmConfiguration(t *testing.T) {
	helm, err := exec.LookPath("helm")
	if err != nil {
		t.Skip("helm is required for the rendered Secret integration test")
	}
	for _, mode := range []string{"default", "csv", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			args := []string{"template", "test", "../../charts/policy-reporter", "--show-only", "templates/config-email-reports-secret.yaml", "--set", "emailReports.summary.enabled=true", "--set", "emailReports.violations.enabled=true"}
			if mode != "default" {
				format := "csv"
				if mode == "invalid" {
					format = "xlsx"
				}
				args = append(args, "--set-string", "emailReports.summary.attachmentFormat="+format, "--set-string", "emailReports.violations.attachmentFormat=csv", "--set-string", "emailReports.summary.channels[0].to[0]=channel@example.test", "--set-string", "emailReports.summary.channels[0].attachmentFormat=csv", "--set-string", "emailReports.summary.channels[1].to[0]=default@example.test")
			}
			output, err := exec.Command(helm, args...).CombinedOutput()
			require.NoError(t, err, string(output))
			var secret corev1.Secret
			require.NoError(t, yaml.Unmarshal(output, &secret))
			require.NotEmpty(t, secret.Data["config.yaml"])
			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, secret.Data["config.yaml"], 0o600))
			command := createCMD()
			require.NoError(t, command.Flags().Set("config", path))
			c, err := config.Load(command)
			if mode == "invalid" {
				require.ErrorContains(t, err, "emailReports.summary.attachmentFormat")
				return
			}
			require.NoError(t, err)
			if mode == "default" {
				require.Empty(t, c.EmailReports.Summary.AttachmentFormat)
				require.Empty(t, c.EmailReports.Violations.AttachmentFormat)
				return
			}
			require.Equal(t, "csv", c.EmailReports.Summary.AttachmentFormat)
			require.Equal(t, "csv", c.EmailReports.Violations.AttachmentFormat)
			require.Equal(t, "csv", c.EmailReports.Summary.Channels[0].AttachmentFormat)
			require.Equal(t, []string{"channel@example.test"}, c.EmailReports.Summary.Channels[0].To)
			require.Empty(t, c.EmailReports.Summary.Channels[1].AttachmentFormat)
		})
	}
}
