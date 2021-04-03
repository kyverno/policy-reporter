package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/helper"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type block struct {
	Type   string  `json:"type"`
	Text   *text   `json:"text,omitempty"`
	Fields []field `json:"fields,omitempty"`
}

type field struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type attachment struct {
	Color  string  `json:"color"`
	Blocks []block `json:"blocks"`
}

type payload struct {
	Username    string       `json:"username,omitempty"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type client struct {
	webhook               string
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func colorFromPriority(p report.Priority) string {
	if p == report.CriticalPriority {
		return "#b80707"
	}
	if p == report.ErrorPriority {
		return "#e20b0b"
	}
	if p == report.WarningPriority {
		return "#f2c744"
	}
	if p == report.InfoPriority {
		return "#36a64f"
	}

	return "#68c2ff"
}

func (s *client) newPayload(result report.Result) payload {
	p := payload{
		Attachments: make([]attachment, 0, 1),
	}

	att := attachment{
		Color:  colorFromPriority(result.Priority),
		Blocks: make([]block, 0),
	}

	policyBlock := block{
		Type:   "section",
		Fields: []field{{Type: "mrkdwn", Text: "*Policy*\n" + result.Policy}},
	}

	if result.Rule != "" {
		policyBlock.Fields = append(policyBlock.Fields, field{Type: "mrkdwn", Text: "*Rule*\n" + result.Rule})
	}

	att.Blocks = append(
		att.Blocks,
		block{Type: "header", Text: &text{Type: "plain_text", Text: "New Policy Report Result"}},
		policyBlock,
	)

	att.Blocks = append(
		att.Blocks,
		block{Type: "section", Text: &text{Type: "mrkdwn", Text: "*Message*\n" + result.Message}},
		block{
			Type: "section",
			Fields: []field{
				{Type: "mrkdwn", Text: "*Priority*\n" + result.Priority.String()},
				{Type: "mrkdwn", Text: "*Status*\n" + result.Status},
			},
		},
	)

	b := block{
		Type:   "section",
		Fields: make([]field, 0, 2),
	}

	if result.Category != "" {
		b.Fields = append(b.Fields, field{Type: "mrkdwn", Text: "*Category*\n" + result.Category})
	}
	if result.Severity != "" {
		b.Fields = append(b.Fields, field{Type: "mrkdwn", Text: "*Severity*\n" + result.Severity})
	}

	if len(b.Fields) > 0 {
		att.Blocks = append(att.Blocks, b)
	}

	res := report.Resource{}

	if len(result.Resources) > 0 {
		res = result.Resources[0]
	}
	if res.Kind != "" {
		att.Blocks = append(
			att.Blocks,
			block{Type: "section", Text: &text{Type: "mrkdwn", Text: "*Resource*"}},
			block{
				Type: "section",
				Fields: []field{
					{Type: "mrkdwn", Text: "*Kind*\n" + res.Kind},
					{Type: "mrkdwn", Text: "*API Version*\n" + res.APIVersion},
				},
			},
			block{
				Type: "section",
				Fields: []field{
					{Type: "mrkdwn", Text: "*Name*\n" + res.Name},
					{Type: "mrkdwn", Text: "*UID*\n" + res.UID},
				},
			},
		)
	}

	if res.Namespace != "" {
		att.Blocks = append(att.Blocks, block{Type: "section", Fields: []field{{Type: "mrkdwn", Text: "*Namespace*\n" + res.Namespace}}})
	}

	p.Attachments = append(p.Attachments, att)

	return p
}

func (s *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(s.minimumPriority) {
		return
	}

	payload := s.newPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] SLACK : %v\n", err.Error())
		return
	}

	req, err := http.NewRequest("POST", s.webhook, body)
	if err != nil {
		log.Printf("[ERROR] SLACK : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := s.client.Do(req)
	helper.HandleHTTPResponse("SLACK", resp, err)
}

func (s *client) SkipExistingOnStartup() bool {
	return s.skipExistingOnStartup
}

func (s *client) Name() string {
	return "Slack"
}

func (s *client) MinimumPriority() string {
	return s.minimumPriority
}

// NewClient creates a new slack.client to send Results to Slack
func NewClient(host, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		host,
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
