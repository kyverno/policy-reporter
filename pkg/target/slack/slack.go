package slack

import (
	"net/http"
	"strings"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
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
	target.BaseClient
	webhook string
	client  httpClient
}

var colors = map[report.Priority]string{
	report.DebugPriority:    "#68c2ff",
	report.InfoPriority:     "#36a64f",
	report.WarningPriority:  "#f2c744",
	report.CriticalPriority: "#b80707",
	report.ErrorPriority:    "#e20b0b",
}

func (s *client) newPayload(result *report.Result) payload {
	p := payload{
		Attachments: make([]attachment, 0, 1),
	}

	att := attachment{
		Color:  colors[result.Priority],
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

	res := &report.Resource{}
	if result.HasResource() {
		res = result.Resource
	}

	if res.UID != "" {
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

	if len(result.Properties) > 0 {
		att.Blocks = append(
			att.Blocks,
			block{Type: "section", Text: &text{Type: "mrkdwn", Text: "*Properties*"}},
		)

		propBlock := block{
			Type:   "section",
			Fields: []field{},
		}

		for property, value := range result.Properties {
			propBlock.Fields = append(propBlock.Fields, field{Type: "mrkdwn", Text: "*" + strings.Title(property) + "*\n" + value})
		}

		att.Blocks = append(att.Blocks, propBlock)
	}

	p.Attachments = append(p.Attachments, att)

	return p
}

func (s *client) Send(result *report.Result) {
	req, err := helper.CreateJSONRequest(s.Name(), "POST", s.webhook, s.newPayload(result))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := s.client.Do(req)
	helper.ProcessHTTPResponse(s.Name(), resp, err)
}

func (s *client) Name() string {
	return "Slack"
}

// NewClient creates a new slack.client to send Results to Slack
func NewClient(host, minimumPriority string, sources []string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		host,
		httpClient,
	}
}
