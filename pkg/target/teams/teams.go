package teams

import (
	"context"
	"fmt"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Slack target
type Options struct {
	target.ClientOptions
	Webhook      string
	CustomFields map[string]string
	Headers      map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	webhook      string
	customFields map[string]string
	headers      map[string]string
	client       http.Client
}

func (s *client) Send(result v1alpha2.PolicyReportResult) {
	s.PostMessage(s.newMessage(result.GetResource(), []v1alpha2.PolicyReportResult{result}))
}

func (s *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (s *client) BatchSend(report v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
	if report.GetScope() == nil {
		for _, r := range results {
			s.Send(r)
		}
	}

	s.PostMessage(s.newMessage(report.GetScope(), results))
}

func (s *client) PostMessage(message *adaptivecard.Message) {
	if err := message.Validate(); err != nil {
		zap.L().Error(s.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

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

func (s *client) SupportsBatchSend() bool {
	return true
}

func (s *client) newMessage(resource *corev1.ObjectReference, results []v1alpha2.PolicyReportResult) *adaptivecard.Message {
	header := adaptivecard.NewContainer()

	if resource != nil {
		header.AddElement(false, adaptivecard.NewTitleTextBlock(formatting.ResourceString(resource), true))
	} else {
		header.AddElement(false, adaptivecard.NewTitleTextBlock("New PolicyReport Results", true))
	}

	header.AddElement(false, adaptivecard.NewTextBlock(fmt.Sprintf("Received %d new Policy Report Results", len(results)), true))

	if len(s.customFields) > 0 {
		header.AddElement(false, MapToColumnSet(s.customFields))
	}

	card := adaptivecard.NewCard()
	card.SetFullWidth()
	card.AddContainer(true, header)

	for _, result := range results {
		stats := newFactSet()
		stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Status", Value: string(result.Result)})

		if result.Severity != "" {
			stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Severity", Value: string(result.Severity)})
		}

		policy := fmt.Sprintf("Policy: %s", result.Policy)

		if result.Rule != "" {
			policy = fmt.Sprintf("%s/%s", policy, result.Rule)
		}

		r := adaptivecard.NewContainer()
		r.Separator = true
		r.Spacing = adaptivecard.SpacingLarge
		r.AddElement(false, newSubTitle(policy))
		r.AddElement(false, adaptivecard.NewTextBlock(result.Category, true))
		r.AddElement(false, stats)
		r.AddElement(false, adaptivecard.NewTextBlock(result.Message, true))

		if len(result.Properties) > 0 {
			r.AddElement(false, MapToColumnSet(result.Properties))
		}

		card.AddContainer(false, r)
	}

	msg := adaptivecard.NewMessage()
	msg.Attach(card)

	return msg
}

// NewClient creates a new teams.client to send Results to MS Teams
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.CustomFields,
		options.Headers,
		options.HTTPClient,
	}
}
