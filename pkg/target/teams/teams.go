package teams

import (
	"context"
	"fmt"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
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

func (s *client) Send(result openreports.ORResultAdapter) {
	s.PostMessage(s.newMessage(result.GetResource(), []*openreports.ORResultAdapter{&result}))
}

func (s *client) CleanUp(_ context.Context, _ openreports.ReportInterface) {}

func (s *client) BatchSend(report openreports.ReportInterface, results []*openreports.ORResultAdapter) {
	if report.GetScope() == nil {
		for idx := range results {
			s.Send(*results[idx])
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

func (s *client) Type() target.ClientType {
	return target.BatchSend
}

func (s *client) newMessage(resource *corev1.ObjectReference, results []*openreports.ORResultAdapter) *adaptivecard.Message {
	header := adaptivecard.NewContainer()

	if resource != nil {
		if err := header.AddElement(false, adaptivecard.NewTitleTextBlock(formatting.ResourceString(resource), true)); err != nil {
			zap.L().Error(s.Name()+": error adding resource title to header", zap.Error(err))
		}
	} else {
		if err := header.AddElement(false, adaptivecard.NewTitleTextBlock("New PolicyReport Results", true)); err != nil {
			zap.L().Error(s.Name()+": error adding title to header", zap.Error(err))
		}
	}

	if err := header.AddElement(false, adaptivecard.NewTextBlock(fmt.Sprintf("Received %d new Policy Report Results", len(results)), true)); err != nil {
		zap.L().Error(s.Name()+": error adding text block to header", zap.Error(err))
	}

	if len(s.customFields) > 0 {
		if err := header.AddElement(false, MapToColumnSet(s.customFields)); err != nil {
			zap.L().Error(s.Name()+": error adding custom fields to header", zap.Error(err))
		}
	}

	card := adaptivecard.NewCard()
	card.SetFullWidth()
	if err := card.AddContainer(true, header); err != nil {
		zap.L().Error(s.Name()+": error adding header to card", zap.Error(err))
	}

	for idx := range results {
		stats := newFactSet()
		stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Status", Value: string(results[idx].Result)})

		if results[idx].Severity != "" {
			stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Severity", Value: string(results[idx].Severity)})
		}

		policy := "Policy: " + results[idx].Policy

		if results[idx].Rule != "" {
			policy = fmt.Sprintf("%s/%s", policy, results[idx].Rule)
		}

		r := adaptivecard.NewContainer()
		r.Separator = true
		r.Spacing = adaptivecard.SpacingLarge
		if err := r.AddElement(false, newSubTitle(policy)); err != nil {
			zap.L().Error(s.Name()+": error adding policy as subtitle to card", zap.Error(err))
		}

		if err := r.AddElement(false, adaptivecard.NewTextBlock(results[idx].Category, true)); err != nil {
			zap.L().Error(s.Name()+": error adding category to card", zap.Error(err))
		}

		if err := r.AddElement(false, stats); err != nil {
			zap.L().Error(s.Name()+": error adding facts to card", zap.Error(err))
		}

		if err := r.AddElement(false, adaptivecard.NewTextBlock(results[idx].Description, true)); err != nil {
			zap.L().Error(s.Name()+": error adding message to card", zap.Error(err))
		}

		if len(results[idx].Properties) > 0 {
			if err := r.AddElement(false, MapToColumnSet(results[idx].Properties)); err != nil {
				zap.L().Error(s.Name()+": error adding properties to card", zap.Error(err))
			}
		}

		if err := card.AddContainer(false, r); err != nil {
			zap.L().Error(s.Name()+": error adding container element to card",
				zap.Error(err))
		}
	}

	msg := adaptivecard.NewMessage()
	if err := msg.Attach(card); err != nil {
		zap.L().Error(s.Name()+": error attaching card", zap.Error(err))
	}

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
