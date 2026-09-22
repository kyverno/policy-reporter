package telegram

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

func executeTemplate(t *testing.T, result openreports.ResultAdapter) string {
	t.Helper()

	tmpl, err := template.New("telegram").Funcs(template.FuncMap{"escape": escape}).Parse(notificationTempl)
	if err != nil {
		t.Fatalf("failed to parse notification template: %v", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, values{Result: result, Time: time.Now()}); err != nil {
		t.Fatalf("failed to render notification template: %v", err)
	}

	return buffer.String()
}

func Test_template(t *testing.T) {
	t.Parallel()
	t.Run("renders the result description", func(t *testing.T) {
		t.Parallel()
		out := executeTemplate(t, fixtures.CompleteTargetSendResult)

		if !strings.Contains(out, escape(fixtures.CompleteTargetSendResult.Description)) {
			t.Error("expected the rendered message to contain the result description")
		}
	})
	t.Run("no Properties section without properties", func(t *testing.T) {
		t.Parallel()
		out := executeTemplate(t, openreports.ResultAdapter{})

		if strings.Contains(out, "Properties") {
			t.Error("expected no Properties section for a result without properties")
		}
	})
}
