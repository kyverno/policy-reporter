package loki

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/helper"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type payload struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Labels  string  `json:"labels"`
	Entries []entry `json:"entries"`
}

type entry struct {
	Ts   string `json:"ts"`
	Line string `json:"line"`
}

func newLokiPayload(result report.Result) payload {
	timestamp := time.Now()
	if !result.Timestamp.IsZero() {
		timestamp = result.Timestamp
	}

	le := entry{Ts: timestamp.Format(time.RFC3339), Line: "[" + strings.ToUpper(result.Priority.String()) + "] " + result.Message}
	ls := stream{Entries: []entry{le}}

	var labels = []string{
		"status=\"" + result.Status + "\"",
		"policy=\"" + result.Policy + "\"",
		"priority=\"" + result.Priority.String() + "\"",
		"source=\"policy-reporter\"",
	}

	if result.Rule != "" {
		labels = append(labels, "rule=\""+result.Rule+"\"")
	}
	if result.Category != "" {
		labels = append(labels, "category=\""+result.Category+"\"")
	}
	if result.Severity != "" {
		labels = append(labels, "severity=\""+result.Severity+"\"")
	}
	if result.HasResource() {
		labels = append(labels, "kind=\""+result.Resource.Kind+"\"")
		labels = append(labels, "name=\""+result.Resource.Name+"\"")
		labels = append(labels, "apiVersion=\""+result.Resource.APIVersion+"\"")
		labels = append(labels, "uid=\""+result.Resource.UID+"\"")
		labels = append(labels, "namespace=\""+result.Resource.Namespace+"\"")
	}

	for property, value := range result.Properties {
		labels = append(labels, strings.ReplaceAll(property, ".", "_")+"=\""+strings.ReplaceAll(value, "\"", "")+"\"")
	}

	ls.Labels = "{" + strings.Join(labels, ",") + "}"

	return payload{Streams: []stream{ls}}
}

type client struct {
	host                  string
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func (l *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(l.minimumPriority) {
		return
	}

	payload := newLokiPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] LOKI : %v\n", err.Error())
		return
	}

	req, err := http.NewRequest("POST", l.host, body)
	if err != nil {
		log.Printf("[ERROR] LOKI : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := l.client.Do(req)
	helper.HandleHTTPResponse("LOKI", resp, err)
}

func (l *client) SkipExistingOnStartup() bool {
	return l.skipExistingOnStartup
}

func (l *client) Name() string {
	return "Loki"
}

func (l *client) MinimumPriority() string {
	return l.minimumPriority
}

// NewClient creates a new loki.client to send Results to Loki
func NewClient(host, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		host + "/api/prom/push",
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
