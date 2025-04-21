package webhook

import (
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"go.uber.org/zap"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Host         string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
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

	resultBody := result.Body()

	req, err := http.CreateJSONRequest("POST", e.host, resultBody)
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
