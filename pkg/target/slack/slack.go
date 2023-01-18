package slack

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Options to configure the Slack target
type Options struct {
	target.ClientOptions
	Webhook      string
	CustomFields map[string]string
	HTTPClient   http.Client
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
	webhook      string
	client       http.Client
	customFields map[string]string
}

var colors = map[report.Priority]string{
	report.DebugPriority:    "#68c2ff",
	report.InfoPriority:     "#36a64f",
	report.WarningPriority:  "#f2c744",
	report.CriticalPriority: "#b80707",
	report.ErrorPriority:    "#e20b0b",
}

func (s *client) newPayload(result report.Result) payload {
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

	if result.HasResource() {
		res := result.Resource

		att.Blocks = append(att.Blocks, block{Type: "section", Text: &text{Type: "mrkdwn", Text: "*Resource*"}})

		if res.APIVersion != "" {
			att.Blocks = append(att.Blocks, block{
				Type: "section",
				Fields: []field{
					{Type: "mrkdwn", Text: "*Kind*\n" + res.Kind},
					{Type: "mrkdwn", Text: "*API Version*\n" + res.APIVersion},
				},
			})
		} else if res.APIVersion == "" && res.UID != "" {
			att.Blocks = append(att.Blocks, block{
				Type: "section",
				Text: &text{Type: "mrkdwn", Text: "*Kind*\n" + res.Kind},
			})
		}

		if res.UID != "" {
			att.Blocks = append(att.Blocks, block{
				Type: "section",
				Fields: []field{
					{Type: "mrkdwn", Text: "*Name*\n" + res.Name},
					{Type: "mrkdwn", Text: "*UID*\n" + res.UID},
				},
			})
		} else if res.UID == "" && res.APIVersion != "" {
			att.Blocks = append(att.Blocks, block{
				Type: "section",
				Text: &text{Type: "mrkdwn", Text: "*Name*\n" + res.Name},
			})
		}

		if res.APIVersion == "" && res.UID == "" {
			att.Blocks = append(att.Blocks, block{
				Type: "section",
				Fields: []field{
					{Type: "mrkdwn", Text: "*Kind*\n" + res.Kind},
					{Type: "mrkdwn", Text: "*Name*\n" + res.Name},
				},
			})
		}

		if res.Namespace != "" {
			att.Blocks = append(att.Blocks, block{Type: "section", Fields: []field{{Type: "mrkdwn", Text: "*Namespace*\n" + res.Namespace}}})
		}
	}

	if len(result.Properties) > 0 || len(s.customFields) > 0 {
		att.Blocks = append(
			att.Blocks,
			block{Type: "section", Text: &text{Type: "mrkdwn", Text: "*Properties*"}},
		)

		propBlock := block{
			Type:   "section",
			Fields: []field{},
		}

		for property, value := range result.Properties {
			propBlock.Fields = append(propBlock.Fields, field{Type: "mrkdwn", Text: "*" + cases.Title(language.English, cases.NoLower).String(property) + "*\n" + value})
		}
		for property, value := range s.customFields {
			propBlock.Fields = append(propBlock.Fields, field{Type: "mrkdwn", Text: "*" + cases.Title(language.English, cases.NoLower).String(property) + "*\n" + value})
		}

		att.Blocks = append(att.Blocks, propBlock)
	}

	p.Attachments = append(p.Attachments, att)

	return p
}

func (s *client) Send(result report.Result) {
	req, err := http.CreateJSONRequest(s.Name(), "POST", s.webhook, s.newPayload(result))
	if err != nil {
		return
	}

	resp, err := s.client.Do(req)
	http.ProcessHTTPResponse(s.Name(), resp, err)
}

// NewClient creates a new slack.client to send Results to Slack
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.HTTPClient,
		options.CustomFields,
	}
}
