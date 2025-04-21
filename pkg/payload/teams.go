package payload

import (
	"fmt"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func (p *PolicyReportResultPayload) ToTeams() (adaptivecard.Container, error) {
	stats := newFactSet()
	stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Status", Value: string(p.Result.Severity)})

	if p.Result.Severity != "" {
		stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Severity", Value: string(p.Result.Severity)})
	}

	policy := fmt.Sprintf("Policy: %s", p.Result.Policy)

	if p.Result.Rule != "" {
		policy = fmt.Sprintf("%s/%s", policy, p.Result.Rule)
	}

	r := adaptivecard.NewContainer()
	r.Separator = true
	r.Spacing = adaptivecard.SpacingLarge
	r.AddElement(false, newSubTitle(policy))
	r.AddElement(false, adaptivecard.NewTextBlock(p.Result.Category, true))
	r.AddElement(false, stats)
	r.AddElement(false, adaptivecard.NewTextBlock(p.Result.Message, true))

	if len(p.Result.Properties) > 0 {
		r.AddElement(false, MapToColumnSet(p.Result.Properties))
	}

	return r, nil
}

func newFactSet() adaptivecard.Element {
	factSet := adaptivecard.Element{
		Type: adaptivecard.TypeElementFactSet,
	}

	return factSet
}

func newFactSetPointer() *adaptivecard.Element {
	factSet := newFactSet()

	return &factSet
}

func newSubTitle(title string) adaptivecard.Element {
	text := adaptivecard.NewTextBlock(title, true)
	text.Weight = adaptivecard.WeightBolder
	text.IsSubtle = true

	return text
}

func MapToColumnSet(list map[string]string) adaptivecard.Element {
	i := 0

	first := adaptivecard.NewColumn()
	first.Items = append(first.Items, newFactSetPointer())

	second := adaptivecard.NewColumn()
	second.Items = append(second.Items, newFactSetPointer())

	propBlock := adaptivecard.NewColumnSet()
	propBlock.Columns = []adaptivecard.Column{first, second}

	for property, value := range list {
		index := i % 2

		propBlock.Columns[index].Items[0].Facts = append(propBlock.Columns[index].Items[0].Facts, adaptivecard.Fact{
			Title: helper.Title(property),
			Value: value,
		})

		i++
	}

	return propBlock
}
