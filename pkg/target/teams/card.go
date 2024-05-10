package teams

import (
	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

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
