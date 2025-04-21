package discord

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"go.uber.org/zap"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Webhook      string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	webhook      string
	customFields map[string]string
	client       http.Client
}

func (d *client) Send(result payload.Payload) {
	if len(d.customFields) > 0 {
		if err := result.AddCustomFields(d.customFields); err != nil {
			zap.L().Error(d.Name()+": Error adding custom fields", zap.Error(err))
			return
		}
	}
	req, err := http.CreateJSONRequest("POST", d.webhook, result.ToDiscord())
	if err != nil {
		return
	}

	resp, err := d.client.Do(req)
	http.ProcessHTTPResponse(d.Name(), resp, err)
}

func (d *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (d *client) BatchSend(_ v1alpha2.ReportInterface, _ []payload.Payload) {}

func (d *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new loki.client to send Results to Discord
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.CustomFields,
		options.HTTPClient,
	}
}
