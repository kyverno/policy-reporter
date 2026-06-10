package webhook

import (
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Host         string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   http.Client
	Keepalive    *v1alpha1.KeepaliveConfig
}

type client struct {
	target.BaseClient
	host         string
	headers      map[string]string
	customFields map[string]string
	client       http.Client
	keepalive    *v1alpha1.KeepaliveConfig
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

	req, err := http.CreateJSONRequest("POST", e.host, http.NewJSONResult(result))
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) SendHeartbeat() {
	payload := map[string]interface{}{
		"event": "heartbeat",
		"time":  time.Now().Format(time.RFC3339),
	}

	// Add keepalive params if configured
	if e.keepalive != nil {
		if len(e.keepalive.Params) > 0 {
			for k, v := range e.keepalive.Params {
				payload[k] = v
			}
		}
	}

	zap.L().Debug("sending heartbeat payload",
		zap.String("target", e.Name()),
		zap.Any("payload", payload))

	req, err := http.CreateJSONRequest("POST", e.host, payload)
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name()+"-heartbeat", resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new webhook client
func NewClient(options Options) target.Client {
	return &client{
		BaseClient:   target.NewBaseClient(options.ClientOptions),
		host:         options.Host,
		headers:      options.Headers,
		customFields: options.CustomFields,
		client:       options.HTTPClient,
		keepalive:    options.Keepalive,
	}
}
