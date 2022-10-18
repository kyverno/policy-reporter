package elasticsearch

import (
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure elasticsearch target
type Options struct {
	target.ClientOptions
	Host         string
	Username     string
	Password     string
	Index        string
	Rotation     string
	CustomFields map[string]string
	HTTPClient   http.Client
}

// Rotation Enum
type Rotation = string

// Elasticsearch Index Rotation
const (
	None     Rotation = "none"
	Daily    Rotation = "daily"
	Monthly  Rotation = "monthly"
	Annually Rotation = "annually"
)

type client struct {
	target.BaseClient
	host         string
	index        string
	username     string
	password     string
	rotation     Rotation
	customFields map[string]string
	client       http.Client
}

func (e *client) Send(result report.Result) {
	var host string
	switch e.rotation {
	case None:
		host = e.host + "/" + e.index + "/event"
	case Annually:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006") + "/event"
	case Monthly:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006.01") + "/event"
	default:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006.01.02") + "/event"
	}

	if len(e.customFields) > 0 {
		props := make(map[string]string, 0)

		for property, value := range e.customFields {
			props[property] = value
		}

		for property, value := range result.Properties {
			props[property] = value
		}

		result.Properties = props
	}

	req, err := http.CreateJSONRequest(e.Name(), "POST", host, result)
	if err != nil {
		return
	}

	if e.username != "" {
		req.SetBasicAuth(e.username, e.password)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

// NewClient creates a new elasticsearch.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Index,
		options.Username,
		options.Password,
		options.Rotation,
		options.CustomFields,
		options.HTTPClient,
	}
}
