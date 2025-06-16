package discord

import (
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Webhook      string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type payload struct {
	Content string  `json:"content"`
	Embeds  []embed `json:"embeds"`
}

type embed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       string       `json:"color"`
	Fields      []embedField `json:"fields"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

var colors = map[v1alpha1.ResultSeverity]string{
	openreports.SeverityInfo:     "12370112",
	openreports.SeverityLow:      "3066993",
	openreports.SeverityMedium:   "15105570",
	openreports.SeverityHigh:     "15158332",
	openreports.SeverityCritical: "15158332",
}

func newPayload(result *openreports.ORResultAdapter, customFields map[string]string) payload {
	color, exists := colors[result.Severity]
	if !exists {
		color = "0"
	}

	embedFields := make([]embedField, 0)

	embedFields = append(embedFields, embedField{"Policy", result.Policy, true})

	if result.Rule != "" {
		embedFields = append(embedFields, embedField{"Rule", result.Rule, true})
	}

	if result.Category != "" {
		embedFields = append(embedFields, embedField{"Category", result.Category, true})
	}
	if result.Severity != "" {
		embedFields = append(embedFields, embedField{"Severity", string(result.Severity), true})
	}

	if result.HasResource() {
		res := result.GetResource()

		embedFields = append(embedFields, embedField{"Kind", res.Kind, true})
		embedFields = append(embedFields, embedField{"Name", res.Name, true})
		if res.Namespace != "" {
			embedFields = append(embedFields, embedField{"Namespace", res.Namespace, true})
		}
		if res.APIVersion != "" {
			embedFields = append(embedFields, embedField{"API Version", res.APIVersion, true})
		}
	}

	for property, value := range result.Properties {
		embedFields = append(embedFields, embedField{helper.Title(property), value, true})
	}

	for property, value := range customFields {
		embedFields = append(embedFields, embedField{helper.Title(property), value, true})
	}

	embeds := make([]embed, 0, 1)
	embeds = append(embeds, embed{
		Title:       "New Policy Report Result",
		Description: result.Description,
		Color:       color,
		Fields:      embedFields,
	})

	return payload{
		Content: "",
		Embeds:  embeds,
	}
}

type client struct {
	target.BaseClient
	webhook      string
	customFields map[string]string
	client       http.Client
}

func (d *client) Send(result *openreports.ORResultAdapter) {
	req, err := http.CreateJSONRequest("POST", d.webhook, newPayload(result, d.customFields))
	if err != nil {
		return
	}

	resp, err := d.client.Do(req)
	http.ProcessHTTPResponse(d.Name(), resp, err)
}

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
