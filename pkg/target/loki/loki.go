package loki

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
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
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values []Value           `json:"values"`
}

type Value = []string

func newLokiStream(result openreports.ORResultAdapter, customFields map[string]string) Stream {
	timestamp := time.Now()
	if result.Timestamp.Seconds != 0 {
		timestamp = time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	}

	labels := map[string]string{
		"status":    string(result.Result),
		"policy":    result.Policy,
		"createdBy": "policy-reporter",
	}

	if result.Rule != "" {
		labels["rule"] = result.Rule
	}
	if result.Category != "" {
		labels["category"] = result.Category
	}
	if result.Severity != "" {
		labels["severity"] = string(result.Severity)
	}
	if result.Source != "" {
		labels["source"] = result.Source
	}
	if result.HasResource() {
		res := result.GetResource()
		if res.APIVersion != "" {
			labels["apiVersion"] = res.APIVersion
			labels["kind"] = res.Kind
			labels["name"] = res.Name
		}
		if res.UID != "" {
			labels["uid"] = string(res.UID)
		}
		if res.Namespace != "" {
			labels["namespace"] = res.Namespace
		}
	}

	for property, value := range result.Properties {
		labels[keyReplacer.Replace(property)] = labelReplacer.Replace(value)
	}

	for label, value := range customFields {
		labels[keyReplacer.Replace(label)] = labelReplacer.Replace(value)
	}

	return Stream{
		Values: []Value{[]string{fmt.Sprintf("%v", timestamp.UnixNano()), "[" + strings.ToUpper(string(result.Severity)) + "] " + result.Description}},
		Stream: labels,
	}
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

func (l *client) Send(result openreports.ORResultAdapter) {
	l.send(Payload{
		Streams: []Stream{
			newLokiStream(result, l.customFields),
		},
	})
}

func (l *client) BatchSend(_ openreports.ReportInterface, results []openreports.ORResultAdapter) {
	l.send(Payload{Streams: helper.Map(results, func(result openreports.ORResultAdapter) Stream {
		return newLokiStream(result, l.customFields)
	})})
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
