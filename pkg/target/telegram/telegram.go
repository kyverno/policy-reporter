package telegram

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

var replacer = strings.NewReplacer(
	"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
	"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
	"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
	"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
)

func escape(text interface{}) string {
	return replacer.Replace(fmt.Sprintf("%v", text))
}

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

type Payload struct {
	Text                  string `json:"text,omitempty"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	ChatID                string `json:"chat_id,omitempty"`
}

type values struct {
	Result   openreports.ResultAdapter
	Time     time.Time
	Resource *corev1.ObjectReference
	Props    map[string]string
	Priority string
}

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	ChatID       string
	Host         string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	chatID       string
	host         string
	headers      map[string]string
	customFields map[string]string
	client       http.Client
}

func (e *client) Send(report openreports.ReportInterface, result openreports.ResultAdapter) {
	if len(e.customFields) > 0 {
		props := make(map[string]string, 0)

		for property, value := range e.customFields {
			props[property] = value
		}

		for property, value := range result.Properties {
			props[property] = value
		}

		result.Properties = props
	}

	payload := Payload{
		ParseMode:             "MarkdownV2",
		DisableWebPagePreview: true,
		ChatID:                e.chatID,
	}

	var textBuffer bytes.Buffer

	ttmpl, err := template.New("telegram").Funcs(template.FuncMap{"escape": escape}).Parse(notificationTempl)
	if err != nil {
		zap.L().Error(e.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	var res *corev1.ObjectReference
	if result.HasResource() {
		res = result.GetResource()
	}

	err = ttmpl.Execute(&textBuffer, values{
		Result:   result,
		Time:     time.Now(),
		Resource: res,
	})
	if err != nil {
		zap.L().Error(e.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	payload.Text = textBuffer.String()

	req, err := http.CreateJSONRequest("POST", e.host, payload)
	if err != nil {
		zap.L().Error(e.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) CleanUp(_ context.Context, _ openreports.ReportInterface) {}

func (e *client) Reset(_ context.Context) error {
	return nil
}

func (e *client) BatchSend(_ openreports.ReportInterface, _ []openreports.ResultAdapter) {}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.ChatID,
		options.Host,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
