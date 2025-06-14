package slack

import (
	"fmt"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
	"github.com/kyverno/policy-reporter/pkg/target/http"
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

func (s *client) message(result v1alpha2.PolicyReportResult) *slack.WebhookMessage {
	p := &slack.WebhookMessage{
		Attachments: make([]slack.Attachment, 0, 1),
		Channel:     s.channel,
	}

	if s.channel != "" {
		p.Channel = s.channel
	}

	att := slack.Attachment{
		Color: colors[result.Severity],
		Blocks: slack.Blocks{
			BlockSet: make([]slack.Block, 0),
		},
	}

	policyBlock := slack.NewSectionBlock(nil, []*slack.TextBlockObject{slack.NewTextBlockObject(slack.MarkdownType, "*Policy*\n"+result.Policy, false, false)}, nil)

	if result.Rule != "" {
		policyBlock.Fields = append(policyBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Rule*\n"+result.Rule, false, false))
	}

	att.Blocks.BlockSet = append(
		att.Blocks.BlockSet,
		slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, "New Policy Report Result", false, false)),
		policyBlock,
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Message*\n"+result.Message, false, false), nil, nil),
		slack.NewSectionBlock(nil, []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Status*\n"+string(result.Result), false, false),
		}, nil),
	)

	b := slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0, 2), nil)

	if result.Category != "" {
		b.Fields = append(b.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Category*\n"+result.Category, false, false))
	}
	if result.Severity != "" {
		b.Fields = append(b.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Severity*\n"+string(result.Severity), false, false))
	}

	if len(b.Fields) > 0 {
		att.Blocks.BlockSet = append(att.Blocks.BlockSet, b)
	}

	if result.HasResource() {
		res := result.GetResource()

		att.Blocks.BlockSet = append(
			att.Blocks.BlockSet,
			slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Resource*", false, false), nil, nil),
		)

		if res.APIVersion != "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					slack.NewTextBlockObject(slack.MarkdownType, "*Kind*\n"+res.Kind, false, false),
					slack.NewTextBlockObject(slack.MarkdownType, "*API Version*\n"+res.APIVersion, false, false),
				}, nil),
			)
		} else if res.APIVersion == "" && res.UID != "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					slack.NewTextBlockObject(slack.MarkdownType, "*Kind*\n"+res.Kind, false, false),
				}, nil),
			)
		}

		if res.UID != "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					slack.NewTextBlockObject(slack.MarkdownType, "*Name*\n"+res.Name, false, false),
					slack.NewTextBlockObject(slack.MarkdownType, "*UID*\n"+string(res.UID), false, false),
				}, nil),
			)
		} else if res.UID == "" && res.APIVersion != "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{slack.NewTextBlockObject(slack.MarkdownType, "*Name*\n"+res.Name, false, false)}, nil),
			)
		}

		if res.APIVersion == "" && res.UID == "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{
					slack.NewTextBlockObject(slack.MarkdownType, "*Kind*\n"+res.Kind, false, false),
					slack.NewTextBlockObject(slack.MarkdownType, "*Name*\n"+res.Name, false, false),
				}, nil),
			)
		}

		if res.Namespace != "" {
			att.Blocks.BlockSet = append(
				att.Blocks.BlockSet,
				slack.NewSectionBlock(nil, []*slack.TextBlockObject{slack.NewTextBlockObject(slack.MarkdownType, "*Namespace*\n"+res.Namespace, false, false)}, nil),
			)
		}
	}

	if len(result.Properties) > 0 || len(s.customFields) > 0 {
		att.Blocks.BlockSet = append(
			att.Blocks.BlockSet,
			slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Properties*", false, false), nil, nil),
		)

		propBlock := slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0), nil)

		for property, value := range result.Properties {
			propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
		}
		for property, value := range s.customFields {
			propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
		}

		att.Blocks.BlockSet = append(att.Blocks.BlockSet, propBlock)
	}

	p.Attachments = append(p.Attachments, att)

	return p
}

func (s *client) batchMessage(polr v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) *slack.WebhookMessage {
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

	for idx := range results {
		resultAttachment := slack.Attachment{
			Color: colors[results[idx].Severity],
			Blocks: slack.Blocks{
				BlockSet: make([]slack.Block, 0),
			},
		}

		policy := fmt.Sprintf("Policy: %s", results[idx].Policy)

		if results[idx].Rule != "" {
			policy = fmt.Sprintf("%s/%s", policy, results[idx].Rule)
		}

		resultAttachment.Blocks.BlockSet = append(
			resultAttachment.Blocks.BlockSet,
			slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, policy, false, false)),
		)

		if results[idx].Category != "" {
			resultAttachment.Blocks.BlockSet = append(
				resultAttachment.Blocks.BlockSet,
				slack.NewContextBlock("", slack.NewTextBlockObject(slack.MarkdownType, "*"+results[idx].Category+"*", false, false)),
			)
		}

		b := slack.NewSectionBlock(nil, []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Status*\n"+string(results[idx].Result), false, false),
		}, nil)

		if results[idx].Severity != "" {
			b.Fields = append(b.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Severity*\n"+string(results[idx].Severity), false, false))
		}

		resultAttachment.Blocks.BlockSet = append(
			resultAttachment.Blocks.BlockSet,
			b,
			slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Message*\n"+results[idx].Message, false, false), nil, nil),
		)

		if len(results[idx].Properties) > 0 {
			resultAttachment.Blocks.BlockSet = append(resultAttachment.Blocks.BlockSet, slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Properties*", false, false), nil, nil))

			propBlock := slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0), nil)

			for property, value := range results[idx].Properties {
				propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
			}

			resultAttachment.Blocks.BlockSet = append(resultAttachment.Blocks.BlockSet, propBlock)
		}

		p.Attachments = append(p.Attachments, resultAttachment)
	}

	return p
}

func (s *client) Send(result v1alpha2.PolicyReportResult) {
	s.PostMessage(s.message(result))
}

func (s *client) BatchSend(report v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
	if report.GetScope() == nil {
		for idx := range results {
			s.Send(results[idx])
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
