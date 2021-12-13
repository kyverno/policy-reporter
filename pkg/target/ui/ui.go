package ui

import (
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	target.BaseClient
	host   string
	client httpClient
}

type resource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	UID        string `json:"uid"`
}

type result struct {
	Message           string    `json:"message"`
	Policy            string    `json:"policy"`
	Rule              string    `json:"rule"`
	Priority          string    `json:"priority"`
	Status            string    `json:"status"`
	Severity          string    `json:"severity,omitempty"`
	Category          string    `json:"category,omitempty"`
	Scored            bool      `json:"scored"`
	Resource          resource  `json:"resource"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

func newPayload(r *report.Result) result {
	return result{
		Message:  r.Message,
		Policy:   r.Policy,
		Rule:     r.Rule,
		Priority: r.Priority.String(),
		Status:   r.Status,
		Severity: r.Severity,
		Category: r.Category,
		Scored:   r.Scored,
		Resource: resource{
			Namespace:  r.Resource.Namespace,
			APIVersion: r.Resource.APIVersion,
			Kind:       r.Resource.Kind,
			Name:       r.Resource.Name,
			UID:        r.Resource.UID,
		},
		CreationTimestamp: r.Timestamp,
	}
}

func (e *client) Send(result *report.Result) {
	req, err := helper.CreateJSONRequest(e.Name(), "POST", e.host, newPayload(result))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := e.client.Do(req)
	helper.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Name() string {
	return "UI"
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(host, minimumPriority string, sources []string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		host + "/api/push",
		httpClient,
	}
}
