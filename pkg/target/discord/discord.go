package discord

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/helper"
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

func newPayload(result report.Result) payload {
	var color string
	switch result.Priority {
	case report.ErrorPriority:
		color = "15158332"
	case report.WarningPriority:
		color = "15105570"
	case report.InfoPriority:
		color = "3066993"
	case report.DebugPriority:
		color = "12370112"
	}

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
	res := report.Resource{}

	if len(result.Resources) > 0 {
		res = result.Resources[0]
	}
	if res.Kind != "" {
		embedFields = append(embedFields, embedField{"Kind", res.Kind, true})
		embedFields = append(embedFields, embedField{"Name", res.Name, true})
		if res.Namespace != "" {
			embedFields = append(embedFields, embedField{"Namespace", res.Namespace, true})
		}
		embedFields = append(embedFields, embedField{"API Version", res.APIVersion, true})
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
	webhook               string
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func (d *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(d.minimumPriority) {
		return
	}

	payload := newPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] DISCORD : %v\n", err.Error())
		return
	}

	req, err := http.NewRequest("POST", d.webhook, body)
	if err != nil {
		log.Printf("[ERROR] DISCORD : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := d.client.Do(req)
	helper.HandleHTTPResponse("DISCORD", resp, err)
}

func (d *client) SkipExistingOnStartup() bool {
	return d.skipExistingOnStartup
}

func (d *client) Name() string {
	return "Discord"
}

func (d *client) MinimumPriority() string {
	return d.minimumPriority
}

// NewClient creates a new loki.client to send Results to Loki
func NewClient(webhook, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		webhook,
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
