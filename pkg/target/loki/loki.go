package loki

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

var keyReplacer = strings.NewReplacer(".", "_", "]", "", "[", "")
var labelReplacer = strings.NewReplacer("/", "")

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

type payload struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Stream map[string]string `json:"stream"`
	Values []value           `json:"values"`
}

type value = []string

func newLokiStream(result v1alpha2.PolicyReportResult, customFields map[string]string) stream {
	timestamp := time.Now()
	if result.Timestamp.Seconds != 0 {
		timestamp = time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))
	}

	labels := map[string]string{
		"status": string(result.Result),
		"policy": result.Policy,
		"source": "policy-reporter",
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
		labels["producer"] = result.Source
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

	return stream{
		Values: []value{[]string{fmt.Sprintf("%v", timestamp.UnixNano()), "[" + strings.ToUpper(string(result.Severity)) + "] " + result.Message}},
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

func (l *client) Send(result v1alpha2.PolicyReportResult) {
	l.send(payload{
		Streams: []stream{
			newLokiStream(result, l.customFields),
		},
	})
}

func (l *client) BatchSend(_ v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
	l.send(payload{Streams: helper.Map(results, func(result v1alpha2.PolicyReportResult) stream {
		return newLokiStream(result, l.customFields)
	})})
}

func (l *client) send(payload payload) {
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

func (l *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {}

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
