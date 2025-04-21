package payload

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

var notificationTempl = `*\[Policy Reporter\] \[{{ .Result.Severity }}\] {{ escape (or .Result.Policy .Result.Rule) }}*
{{- if .Resource }}

*Resource*: {{ .Resource.Kind }} {{ if .Resource.Namespace }}{{ escape .Resource.Namespace }}/{{ end }}{{ escape .Resource.Name }}

{{- end }}

*Status*: {{ escape .Result.Result }}
*Time*: {{ escape (.Time.Format "02 Jan 06 15:04 MST") }}

{{ if .Result.Category }}*Category*: {{ escape .Result.Category }}{{ end }}
{{ if .Result.Policy }}*Rule*: {{ escape .Result.Rule }}{{ end }}
*Source*: {{ escape .Result.Source }}

*Message*:

{{ escape .Result.Message }}

*Properties*:
{{ range $key, $value := .Result.Properties }}• *{{ escape $key }}*: {{ escape $value }}
{{ end }}
`

var replacer = strings.NewReplacer(
	"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
	"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
	"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
	"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
)

func escape(text interface{}) string {
	return replacer.Replace(fmt.Sprintf("%v", text))
}

type values struct {
	Result   v1alpha2.PolicyReportResult
	Time     time.Time
	Resource *corev1.ObjectReference
	Props    map[string]string
	Priority string
}

func (s *PolicyReportResultPayload) ToTelegram(chatID string) (string, error) {
	var textBuffer bytes.Buffer

	ttmpl, err := template.New("telegram").Funcs(template.FuncMap{"escape": escape}).Parse(notificationTempl)
	if err != nil {
		return "", err
	}

	var res *corev1.ObjectReference
	if s.Result.HasResource() {
		res = s.Result.GetResource()
	}

	err = ttmpl.Execute(&textBuffer, values{
		Result:   s.Result,
		Time:     time.Now(),
		Resource: res,
	})
	if err != nil {
		return "", err
	}

	return textBuffer.String(), nil
}
