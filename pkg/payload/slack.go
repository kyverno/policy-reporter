package payload

import (
	"github.com/slack-go/slack"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

var slackColors = map[v1alpha2.PolicySeverity]string{
	v1alpha2.SeverityInfo:     "#68c2ff",
	v1alpha2.SeverityLow:      "#36a64f",
	v1alpha2.SeverityMedium:   "#f2c744",
	v1alpha2.SeverityHigh:     "#b80707",
	v1alpha2.SeverityCritical: "#e20b0b",
}

func (s *PolicyReportResultPayload) ToSlack(channel string) *slack.Attachment {
	att := slack.Attachment{
		Color: slackColors[s.Result.Severity],
		Blocks: slack.Blocks{
			BlockSet: make([]slack.Block, 0),
		},
	}

	policyBlock := slack.NewSectionBlock(nil, []*slack.TextBlockObject{slack.NewTextBlockObject(slack.MarkdownType, "*Policy*\n"+s.Result.Policy, false, false)}, nil)

	if s.Result.Rule != "" {
		policyBlock.Fields = append(policyBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Rule*\n"+s.Result.Rule, false, false))
	}

	att.Blocks.BlockSet = append(
		att.Blocks.BlockSet,
		slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, "New Policy Report Result", false, false)),
		policyBlock,
	)

	att.Blocks.BlockSet = append(
		att.Blocks.BlockSet,
		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Message*\n"+s.Result.Message, false, false), nil, nil),
		slack.NewSectionBlock(nil, []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, "*Status*\n"+string(s.Result.Result), false, false),
		}, nil),
	)

	b := slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0, 2), nil)

	if s.Result.Category != "" {
		b.Fields = append(b.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Category*\n"+s.Result.Category, false, false))
	}
	if s.Result.Severity != "" {
		b.Fields = append(b.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*Severity*\n"+string(s.Result.Severity), false, false))
	}

	if len(b.Fields) > 0 {
		att.Blocks.BlockSet = append(att.Blocks.BlockSet, b)
	}

	if s.Result.HasResource() {
		res := s.Result.GetResource()

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

	// if len(s.Result.Properties) > 0 || len(s.customFields) > 0 {
	// 	att.Blocks.BlockSet = append(
	// 		att.Blocks.BlockSet,
	// 		slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*Properties*", false, false), nil, nil),
	// 	)

	// 	propBlock := slack.NewSectionBlock(nil, make([]*slack.TextBlockObject, 0), nil)

	// 	for property, value := range s.Result.Properties {
	// 		propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
	// 	}
	// 	for property, value := range s.customFields {
	// 		propBlock.Fields = append(propBlock.Fields, slack.NewTextBlockObject(slack.MarkdownType, "*"+helper.Title(property)+"*\n"+value, false, false))
	// 	}

	// 	att.Blocks.BlockSet = append(att.Blocks.BlockSet, propBlock)
	// }

	return &att
}
