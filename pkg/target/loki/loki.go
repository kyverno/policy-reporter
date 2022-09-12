package loki

import (
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Loko target
type Options struct {
	target.ClientOptions
	Host         string
	CustomLabels map[string]string
	HTTPClient   http.Client
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

func newLokiPayload(result report.Result, customLabels map[string]string) payload {
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
	if result.Source != "" {
		labels = append(labels, "producer=\""+result.Source+"\"")
	}
	if result.HasResource() {
		if result.Resource.APIVersion != "" {
			labels = append(labels, "apiVersion=\""+result.Resource.APIVersion+"\"")
		}
		labels = append(labels, "kind=\""+result.Resource.Kind+"\"")
		labels = append(labels, "name=\""+result.Resource.Name+"\"")
		if result.Resource.UID != "" {
			labels = append(labels, "uid=\""+result.Resource.UID+"\"")
		}
		if result.Resource.Namespace != "" {
			labels = append(labels, "namespace=\""+result.Resource.Namespace+"\"")
		}
	}

	for property, value := range result.Properties {
		labels = append(labels, strings.ReplaceAll(property, ".", "_")+"=\""+strings.ReplaceAll(value, "\"", "")+"\"")
	}

	for label, value := range customLabels {
		labels = append(labels, strings.ReplaceAll(label, ".", "_")+"=\""+strings.ReplaceAll(value, "\"", "")+"\"")
	}

	ls.Labels = "{" + strings.Join(labels, ",") + "}"

	return payload{Streams: []stream{ls}}
}

type client struct {
	target.BaseClient
	host         string
	client       http.Client
	customLabels map[string]string
}

func (l *client) Send(result report.Result) {
	req, err := http.CreateJSONRequest(l.Name(), "POST", l.host, newLokiPayload(result, l.customLabels))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	http.ProcessHTTPResponse(l.Name(), resp, err)
}

// NewClient creates a new loki.client to send Results to Loki
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.HTTPClient,
		options.CustomLabels,
	}
}
