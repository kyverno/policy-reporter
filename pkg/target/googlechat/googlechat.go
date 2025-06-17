package googlechat

import (
	"bytes"
	"text/template"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

const (
	messageTempl  string = `[{{ .Result.Severity }}] {{ or .Result.Policy .Result.Rule }}`
	resourceTempl string = `{{ if .Namespace }}[{{ .Namespace }}] {{ end }} {{ .APIVersion }}/{{ .Kind }} {{ .Name }}`
)

type values struct {
	Result   openreports.ORResultAdapter
	Priority string
	Resource *corev1.ObjectReference
}

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

type Payload struct {
	CardsV2 []cardsV2 `json:"cardsV2,omitempty"`
}

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Webhook      string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	webhook      string
	headers      map[string]string
	customFields map[string]string
	client       http.Client
}

func mapPayload(result openreports.ORResultAdapter) (*Payload, error) {
	widgets := []widget{{TextParagraph: &textParagraph{Text: result.Description}}}

	ttmpl, err := template.New("googlechat").Parse(messageTempl)
	if err != nil {
		return nil, err
	}

	prio := result.Severity
	if prio == "" {
		prio = openreports.SeverityInfo
	}

	var textBuffer bytes.Buffer
	err = ttmpl.Execute(&textBuffer, values{Result: result, Resource: result.GetResource()})
	if err != nil {
		return nil, err
	}

	subtitle := ""

	if result.HasResource() {
		res := result.GetResource()

		widgets = append(widgets, widget{
			Columns: &columns{
				ColumnItems: []column{
					{
						Widgets: []widget{
							{DecoratedText: &decoratedText{TopLabel: "Kind", Text: res.Kind}},
							{DecoratedText: &decoratedText{TopLabel: "Namespace", Text: res.Namespace}},
							{DecoratedText: &decoratedText{"Status", string(result.Result)}},
						},
					},
					{
						Widgets: []widget{
							{DecoratedText: &decoratedText{TopLabel: "APIVersion", Text: res.APIVersion}},
							{DecoratedText: &decoratedText{TopLabel: "Name", Text: res.Name}},
							{DecoratedText: &decoratedText{"Source", result.Source}},
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

	if result.Policy != "" {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{"Rule", result.Rule}})
	}
	if result.Category != "" {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{"Category", result.Category}})
	}

	for key, value := range result.Properties {
		widgets = append(widgets, widget{DecoratedText: &decoratedText{TopLabel: key, Text: value}})
	}

	widgets = append(widgets, widget{DecoratedText: &decoratedText{"time", time.Now().Format("02 Jan 06 15:04 MST")}})

	return &Payload{
		CardsV2: []cardsV2{
			{
				CardID: result.ID,
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

func (e *client) Send(result openreports.ORResultAdapter) {
	if len(e.customFields) > 0 {
		props := make(map[string]string, 0)

		for property, value := range e.customFields {
			props[property] = value
		}

		for property, value := range result.Properties {
			props[property] = value
		}

		result.Properties = props
	}

	payload, err := mapPayload(result)
	if err != nil {
		zap.L().Error(e.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	req, err := http.CreateJSONRequest("POST", e.webhook, payload)
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
