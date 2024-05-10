package discord

import (
	"context"
	"strings"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
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

var colors = map[v1alpha2.Priority]string{
	v1alpha2.DebugPriority:    "12370112",
	v1alpha2.InfoPriority:     "3066993",
	v1alpha2.WarningPriority:  "15105570",
	v1alpha2.CriticalPriority: "15158332",
	v1alpha2.ErrorPriority:    "15158332",
}

func newPayload(result v1alpha2.PolicyReportResult, customFields map[string]string) payload {
	color := colors[result.Priority]

	embedFields := make([]embedField, 0)

	embedFields = append(embedFields, embedField{"Policy", result.Policy, true})

	if result.Rule != "" {
		embedFields = append(embedFields, embedField{"Rule", result.Rule, true})
	}

	embedFields = append(embedFields, embedField{"Priority", result.Priority.String(), true})

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
		embedFields = append(embedFields, embedField{strings.Title(property), value, true})
	}

	for property, value := range customFields {
		embedFields = append(embedFields, embedField{strings.Title(property), value, true})
	}

	embeds := make([]embed, 0, 1)
	embeds = append(embeds, embed{
		Title:       "New Policy Report Result",
		Description: result.Message,
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

func (d *client) Send(result v1alpha2.PolicyReportResult) {
	req, err := http.CreateJSONRequest("POST", d.webhook, newPayload(result, d.customFields))
	if err != nil {
		return
	}

	resp, err := d.client.Do(req)
	http.ProcessHTTPResponse(d.Name(), resp, err)
}

func (d *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (d *client) BatchSend(_ v1alpha2.ReportInterface, _ []v1alpha2.PolicyReportResult) {}

func (d *client) SupportsBatchSend() bool {
	return false
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
