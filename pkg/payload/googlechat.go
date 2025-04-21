package payload

import (
	"bytes"
	"text/template"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

const (
	messageTempl  string = `[{{ .Result.Severity }}] {{ or .Result.Policy .Result.Rule }}`
	resourceTempl string = `{{ if .Namespace }}[{{ .Namespace }}] {{ end }} {{ .APIVersion }}/{{ .Kind }} {{ .Name }}`
)

type header struct {
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}

type decoratedText struct {
	TopLabel string `json:"topLabel"`
	Text     string `json:"text"`
}

type column struct {
	Widgets []widget `json:"widgets"`
}

type columns struct {
	ColumnItems []column `json:"columnItems"`
}

type textParagraph struct {
	Text string `json:"text"`
}

type widget struct {
	DecoratedText *decoratedText `json:"decoratedText,omitempty"`
	TextParagraph *textParagraph `json:"textParagraph,omitempty"`
	Columns       *columns       `json:"columns,omitempty"`
}

type section struct {
	Header      string   `json:"header,omitempty"`
	Collapsible bool     `json:"collapsible,omitempty"`
	Widgets     []widget `json:"widgets,omitempty"`
}

type card struct {
	Header   *header   `json:"header,omitempty"`
	Sections []section `json:"sections,omitempty"`
}

type cardsV2 struct {
	CardID string `json:"cardId,omitempty"`
	Card   card   `json:"card,omitempty"`
}

type GCPayload struct {
	CardsV2 []cardsV2 `json:"cardsV2,omitempty"`
}

func (p *PolicyReportResultPayload) ToGoogleChat() (*GCPayload, error) {
	widgets := []widget{{TextParagraph: &textParagraph{Text: p.Result.Message}}}

	ttmpl, err := template.New("googlechat").Parse(messageTempl)
	if err != nil {
		return nil, err
	}

	prio := p.Result.Severity
	if prio == "" {
		prio = v1alpha2.SeverityInfo
	}

	var textBuffer bytes.Buffer
	err = ttmpl.Execute(&textBuffer, values{Result: p.Result, Resource: p.Result.GetResource()})
	if err != nil {
		return nil, err
	}

	subtitle := ""

	if p.Result.HasResource() {
		res := p.Result.GetResource()

		widgets = append(widgets, widget{
			Columns: &columns{
				ColumnItems: []column{
					{
						Widgets: []widget{
							{DecoratedText: &decoratedText{TopLabel: "Kind", Text: res.Kind}},
							{DecoratedText: &decoratedText{TopLabel: "Namespace", Text: res.Namespace}},
							{DecoratedText: &decoratedText{"Status", string(p.Result.Result)}},
						},
					},
					{
						Widgets: []widget{
							{DecoratedText: &decoratedText{TopLabel: "APIVersion", Text: res.APIVersion}},
							{DecoratedText: &decoratedText{TopLabel: "Name", Text: res.Name}},
							{DecoratedText: &decoratedText{"Source", p.Result.Source}},
						},
					},
				},
			},
		})

		stmpl, err := template.New("googlechat:resource").Parse(resourceTempl)
		if err != nil {
			return nil, err
		}

		var subTitleBuffer bytes.Buffer
		err = stmpl.Execute(&subTitleBuffer, res)
		if err != nil {
			return nil, err
		}

		subtitle = subTitleBuffer.String()
	}

	header := header{
		Title:    textBuffer.String(),
		SubTitle: subtitle,
	}

	if p.Result.Policy != "" {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{"Rule", p.Result.Rule}})
	}
	if p.Result.Category != "" {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{"Category", p.Result.Category}})
	}

	for key, value := range p.Result.Properties {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{TopLabel: key, Text: value}})
	}

	widgets = append(widgets, widget{DecoratedText: &decoratedText{"time", time.Now().Format("02 Jan 06 15:04 MST")}})

	return &GCPayload{
		CardsV2: []cardsV2{
			{
				CardID: p.Result.ID,
				Card: card{
					Header: &header,
					Sections: []section{
						{
							Header:      "Details",
							Collapsible: true,
							Widgets:     widgets,
						},
					},
				},
			},
		},
	}, nil
}
