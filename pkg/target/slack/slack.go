package slack

import (
	"context"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Slack target
type Options struct {
	target.ClientOptions
	Webhook      string
	Channel      string
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
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type client struct {
	target.BaseClient
	webhook      string
	channel      string
	client       http.Client
	customFields map[string]string
}

var colors = map[v1alpha2.Priority]string{
	v1alpha2.DebugPriority:    "#68c2ff",
	v1alpha2.InfoPriority:     "#36a64f",
	v1alpha2.WarningPriority:  "#f2c744",
	v1alpha2.CriticalPriority: "#b80707",
	v1alpha2.ErrorPriority:    "#e20b0b",
}

func (s *client) message(result v1alpha2.PolicyReportResult) *slack.WebhookMessage {
	p := &slack.WebhookMessage{
		Attachments: make([]slack.Attachment, 0, 1),
		Channel:     s.channel,
	}

	att := slack.Attachment{
		Color: colors[result.Priority],
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
	)

	att.Blocks.BlockSet = append(
		att.Blocks.BlockSet,
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Message*\n"+result.Message, false, false), nil, nil),
		slack.NewSectionBlock(nil, []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Priority*\n"+result.Priority.String(), false, false),
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

func (s *client) Send(result v1alpha2.PolicyReportResult) {
	if err := slack.PostWebhook(s.webhook, s.message(result)); err != nil {
		zap.L().Error(s.Name()+": PUSH FAILED", zap.Error(err))
	}
}

func (s *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (s *client) BatchSend(_ v1alpha2.ReportInterface, _ []v1alpha2.PolicyReportResult) {}

func (s *client) SupportsBatchSend() bool {
	return false
}

// NewClient creates a new slack.client to send Results to Slack
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.Channel,
		options.HTTPClient,
		options.CustomFields,
	}
}
