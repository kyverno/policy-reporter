package webhook

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Host       string
	Headers    map[string]string
	HTTPClient http.Client
}

type client struct {
	target.BaseClient
	host    string
	headers map[string]string
	client  http.Client
}

func (e *client) Send(result report.Result) {
	req, err := http.CreateJSONRequest(e.Name(), "POST", e.host, http.NewJSONResult(result))
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Headers,
		options.HTTPClient,
	}
}
