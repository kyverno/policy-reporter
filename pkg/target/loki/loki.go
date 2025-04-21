package loki

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"go.uber.org/zap"
)

// Options to configure the Loki target
type Options struct {
	target.ClientOptions
	Host         string
	CustomFields map[string]string
	Headers      map[string]string
	HTTPClient   http.Client
	Username     string
	Password     string
}

type Payload struct {
	Streams []payload.Stream `json:"streams"`
}

type client struct {
	target.BaseClient
	host         string
	client       http.Client
	customFields map[string]string
	headers      map[string]string
	username     string
	password     string
}

func (l *client) Send(result payload.Payload) {
	if len(l.customFields) > 0 {
		if err := result.AddCustomFields(l.customFields); err != nil {
			zap.L().Error(l.Name()+": Error adding custom fields", zap.Error(err))
			return
		}
	}
	s, err := result.ToLoki()
	if err != nil {
		zap.L().Error(l.Name()+": Error converting to loki stream", zap.Error(err))
		return
	}
	l.send(Payload{
		Streams: []payload.Stream{
			s,
		},
	})
}

func (l *client) BatchSend(_ v1alpha2.ReportInterface, results []payload.Payload) {
	lokiResults := []payload.Stream{}

	for _, r := range results {
		lokiRes, err := r.ToLoki()
		if err != nil {
			zap.L().Error(l.Name()+"Error converting to loki stream", zap.Error(err))
			continue
		}

		if len(l.customFields) > 0 {
			if err := r.AddCustomFields(l.customFields); err != nil {
				zap.L().Error(l.Name()+": Error adding custom fields", zap.Error(err))
				continue
			}
		}
		lokiResults = append(lokiResults, lokiRes)
	}

	l.send(Payload{Streams: lokiResults})
}

func (l *client) send(payload Payload) {
	req, err := http.CreateJSONRequest("POST", l.host, payload)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range l.headers {
		req.Header.Set(k, v)
	}

	if l.username != "" {
		req.SetBasicAuth(l.username, l.password)
	}

	resp, err := l.client.Do(req)
	http.ProcessHTTPResponse(l.Name(), resp, err)
}

func (l *client) Type() target.ClientType {
	return target.BatchSend
}

// NewClient creates a new loki.client to send Results to Loki
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.HTTPClient,
		options.CustomFields,
		options.Headers,
		options.Username,
		options.Password,
	}
}
