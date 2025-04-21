package teams

import (
	"context"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
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

func (s *client) Send(result payload.Payload) {
	s.PostMessage(s.newMessage(nil, []payload.Payload{result}))
}

func (s *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

func (s *client) BatchSend(report v1alpha2.ReportInterface, results []payload.Payload) {
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

func (s *client) Type() target.ClientType {
	return target.BatchSend
}

func (s *client) newMessage(resource *corev1.ObjectReference, results []payload.Payload) *adaptivecard.Message {
	header := adaptivecard.NewContainer()

	if len(s.customFields) > 0 {
		header.AddElement(false, MapToColumnSet(s.customFields))
	}

	card := adaptivecard.NewCard()
	card.SetFullWidth()
	card.AddContainer(true, header)

	for _, result := range results {
		cont, err := result.ToTeams()
		if err != nil {
			zap.L().Error(s.Name()+": Error in teams conversion", zap.Error(err))
			continue
		}
		card.AddContainer(false, cont)
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
