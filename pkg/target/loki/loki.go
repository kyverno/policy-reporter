package loki

import (
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure the Loko target
type Options struct {
	target.ClientOptions
	Host         string
	CustomLabels map[string]string
	Headers      map[string]string
	HTTPClient   http.Client
	Username     string
	Password     string
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

func newLokiPayload(result v1alpha2.PolicyReportResult, customLabels map[string]string) payload {
	timestamp := time.Now()
	if result.Timestamp.Seconds != 0 {
		timestamp = time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	}

	le := entry{Ts: timestamp.Format(time.RFC3339), Line: "[" + strings.ToUpper(result.Priority.String()) + "] " + result.Message}
	ls := stream{Entries: []entry{le}}

	labels := []string{
		"status=\"" + string(result.Result) + "\"",
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
		labels = append(labels, "severity=\""+string(result.Severity)+"\"")
	}
	if result.Source != "" {
		labels = append(labels, "producer=\""+result.Source+"\"")
	}
	if result.HasResource() {
		res := result.GetResource()
		if res.APIVersion != "" {
			labels = append(labels, "apiVersion=\""+res.APIVersion+"\"")
		}
		labels = append(labels, "kind=\""+res.Kind+"\"")
		labels = append(labels, "name=\""+res.Name+"\"")
		if res.UID != "" {
			labels = append(labels, "uid=\""+string(res.UID)+"\"")
		}
		if res.Namespace != "" {
			labels = append(labels, "namespace=\""+res.Namespace+"\"")
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
	headers      map[string]string
	username     string
	password     string
}

func (l *client) Send(result v1alpha2.PolicyReportResult) {
	req, err := http.CreateJSONRequest(l.Name(), "POST", l.host, newLokiPayload(result, l.customLabels))
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

// NewClient creates a new loki.client to send Results to Loki
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.HTTPClient,
		options.CustomLabels,
		options.Headers,
		options.Username,
		options.Password,
	}
}
