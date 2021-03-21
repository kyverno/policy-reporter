package ui

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/helper"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	host                  string
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
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

func newPayload(r report.Result) result {
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
			Namespace:  r.Resources[0].Namespace,
			APIVersion: r.Resources[0].APIVersion,
			Kind:       r.Resources[0].Kind,
			Name:       r.Resources[0].Name,
			UID:        r.Resources[0].UID,
		},
		CreationTimestamp: time.Now(),
	}
}

func (e *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(e.minimumPriority) {
		return
	}

	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(newPayload(result)); err != nil {
		log.Printf("[ERROR] UI : %v\n", err.Error())
		return
	}

	req, err := http.NewRequest("POST", e.host, body)
	if err != nil {
		log.Printf("[ERROR] UI : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := e.client.Do(req)
	helper.HandleHTTPResponse("UI", resp, err)
}

func (e *client) SkipExistingOnStartup() bool {
	return e.skipExistingOnStartup
}

func (e *client) Name() string {
	return "UI"
}

func (e *client) MinimumPriority() string {
	return e.minimumPriority
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(host, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		host + "/api/push",
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
