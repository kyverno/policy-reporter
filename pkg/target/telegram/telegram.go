package telegram

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
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

type Payload struct {
	Text                  string `json:"text,omitempty"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	ChatID                string `json:"chat_id,omitempty"`
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

func (e *client) Send(result payload.Payload) {
	if len(e.customFields) > 0 {
		if err := result.AddCustomFields(e.customFields); err != nil {
			zap.L().Error(e.Name()+": Error adding custom fields", zap.Error(err))
			return
		}
	}

	payload := Payload{
		ParseMode:             "MarkdownV2",
		DisableWebPagePreview: true,
		ChatID:                e.chatID,
	}

	payloadText, err := result.ToTelegram(e.chatID)
	if err != nil {
		zap.L().Error(e.Name()+": PUSH FAILED", zap.Error(err))
		fmt.Println(err)
		return
	}

	payload.Text = payloadText

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

func (e *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (e *client) Reset(_ context.Context) error {
	return nil
}

func (e *client) BatchSend(_ v1alpha2.ReportInterface, _ []payload.Payload) {}

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
