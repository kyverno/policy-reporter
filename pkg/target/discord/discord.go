package discord

import (
	"net/http"
	"strings"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

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

var colors = map[report.Priority]string{
	report.DebugPriority:    "12370112",
	report.InfoPriority:     "3066993",
	report.WarningPriority:  "15105570",
	report.CriticalPriority: "15158332",
	report.ErrorPriority:    "15158332",
}

func newPayload(result *report.Result) payload {
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
		embedFields = append(embedFields, embedField{"Severity", result.Severity, true})
	}

	if result.HasResource() {
		embedFields = append(embedFields, embedField{"Kind", result.Resource.Kind, true})
		embedFields = append(embedFields, embedField{"Name", result.Resource.Name, true})
		if result.Resource.Namespace != "" {
			embedFields = append(embedFields, embedField{"Namespace", result.Resource.Namespace, true})
		}
		embedFields = append(embedFields, embedField{"API Version", result.Resource.APIVersion, true})
	}

	for property, value := range result.Properties {
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

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	target.BaseClient
	webhook string
	client  httpClient
}

func (d *client) Send(result *report.Result) {
	req, err := helper.CreateJSONRequest(d.Name(), "POST", d.webhook, newPayload(result))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := d.client.Do(req)
	helper.ProcessHTTPResponse(d.Name(), resp, err)
}

func (d *client) Name() string {
	return "Discord"
}

// NewClient creates a new loki.client to send Results to Discord
func NewClient(webhook, minimumPriority string, sources []string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		webhook,
		httpClient,
	}
}
