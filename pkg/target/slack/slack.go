package slack

import (
	"fmt"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
)

// Options to configure the Slack target
type Options struct {
	target.ClientOptions
	Channel      string
	Webhook      string
	CustomFields map[string]string
	Headers      map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	channel      string
	webhook      string
	client       http.Client
	customFields map[string]string
	headers      map[string]string
}

var colors = map[v1alpha2.PolicySeverity]string{
	v1alpha2.SeverityInfo:     "#68c2ff",
	v1alpha2.SeverityLow:      "#36a64f",
	v1alpha2.SeverityMedium:   "#f2c744",
	v1alpha2.SeverityHigh:     "#b80707",
	v1alpha2.SeverityCritical: "#e20b0b",
}

func (s *client) message(result payload.Payload) *slack.WebhookMessage {
	p := &slack.WebhookMessage{
		Attachments: make([]slack.Attachment, 0, 1),
		Channel:     s.channel,
	}
	att := result.ToSlack(s.channel)

	p.Attachments = append(p.Attachments, *att)
	return p
}

func (s *client) batchMessage(polr v1alpha2.ReportInterface, results []payload.Payload) *slack.WebhookMessage {
	scope := polr.GetScope()
	resource := formatting.ResourceString(scope)

	p := &slack.WebhookMessage{
		Attachments: make([]slack.Attachment, 0, 1),
		Channel:     s.channel,
	}

	if s.channel != "" {
		p.Channel = s.channel
	}

	att := slack.Attachment{
		Color: colors[v1alpha2.SeverityInfo],
		Blocks: slack.Blocks{
			BlockSet: make([]slack.Block, 0),
		},
	}

	att.Blocks.BlockSet = append(
		att.Blocks.BlockSet,
		slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, resource+" Policy Report Result", false, false)),
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("Received %d new Policy Report Results", len(results)), false, false), nil, nil),
	)

	cfl := len(s.customFields)
	if cfl > 0 {
		att.Blocks.BlockSet = append(att.Blocks.BlockSet, slack.NewDividerBlock(), slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Custom Fields*", false, false), nil, nil))

		i := 0

		var propBlock *slack.SectionBlock
		for property, value := range s.customFields {
			if i%2 == 0 {
				propBlock = slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0, 2), nil)
				att.Blocks.BlockSet = append(att.Blocks.BlockSet, propBlock)
			}

			propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
			i++
		}
	}

	p.Attachments = append(p.Attachments, att)

	for _, result := range results {
		p.Attachments = append(p.Attachments, *result.ToSlack(s.channel))
	}

	return p
}

func (s *client) Send(result payload.Payload) {
	s.PostMessage(s.message(result))
}

func (s *client) BatchSend(report v1alpha2.ReportInterface, results []payload.Payload) {
	if report.GetScope() == nil {
		for _, result := range results {
			s.Send(result)
		}

		return
	}

	s.PostMessage(s.batchMessage(report, results))
}

func (s *client) PostMessage(message *slack.WebhookMessage) {
	req, err := http.CreateJSONRequest("POST", s.webhook, message)
	if err != nil {
		zap.L().Error(s.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)

	http.ProcessHTTPResponse(s.Name(), resp, err)
}

func (s *client) Type() target.ClientType {
	return target.BatchSend
}

// NewClient creates a new slack.client to send Results to Slack
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Channel,
		options.Webhook,
		options.HTTPClient,
		options.CustomFields,
		options.Headers,
	}
}
