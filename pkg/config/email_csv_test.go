package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func TestEmailAttachmentConfiguration(t *testing.T) {
	for _, format := range []string{"", "csv", "CSV", "xlsx", "unknown"} {
		for _, reportType := range []string{"summary", "violations"} {
			for _, channel := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/channel=%t", reportType, format, channel), func(t *testing.T) {
					path := filepath.Join(t.TempDir(), "config.yaml")
					field := fmt.Sprintf("    attachmentFormat: %q\n", format)
					if channel {
						field = fmt.Sprintf("    channels:\n    - attachmentFormat: %q\n", format)
					}
					require.NoError(t, os.WriteFile(path, []byte("emailReports:\n  "+reportType+":\n"+field), 0o600))
					command := createCMD()
					require.NoError(t, command.Flags().Set("config", path))
					c, err := config.Load(command)
					if format != "" && format != "csv" {
						require.ErrorContains(t, err, "attachmentFormat")
						if channel {
							require.ErrorContains(t, err, "channels[0]")
						}
						return
					}
					require.NoError(t, err)
					selected, other := c.EmailReports.Summary, c.EmailReports.Violations
					if reportType == "violations" {
						selected, other = other, selected
					}
					require.Empty(t, other.AttachmentFormat)
					if channel {
						require.Empty(t, selected.AttachmentFormat)
						require.Equal(t, format, selected.Channels[0].AttachmentFormat)
					} else {
						require.Equal(t, format, selected.AttachmentFormat)
					}
				})
			}
		}
	}
}
