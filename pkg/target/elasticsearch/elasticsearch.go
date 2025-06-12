package elasticsearch

import (
	"time"

	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Options to configure elasticsearch target
type Options struct {
	target.ClientOptions
	Host         string
	Username     string
	Password     string
	ApiKey       string
	Index        string
	Rotation     string
	CustomFields map[string]string
	Headers      map[string]string
	HTTPClient   http.Client
	// https://www.elastic.co/blog/moving-from-types-to-typeless-apis-in-elasticsearch-7-0
	TypelessApi bool
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
	apiKey       string
	rotation     Rotation
	customFields map[string]string
	headers      map[string]string
	client       http.Client
	// https://www.elastic.co/blog/moving-from-types-to-typeless-apis-in-elasticsearch-7-0
	typelessApi bool
}

func (e *client) Send(result v1alpha1.ReportResult) {
	var host string
	var apiSuffix string
	if e.typelessApi {
		apiSuffix = "_doc"
	} else {
		apiSuffix = "event"
	}

	switch e.rotation {
	case None:
		host = e.host + "/" + e.index + "/" + apiSuffix
	case Annually:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006") + "/" + apiSuffix
	case Monthly:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006.01") + "/" + apiSuffix
	default:
		host = e.host + "/" + e.index + "-" + time.Now().Format("2006.01.02") + "/" + apiSuffix
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

	req, err := http.CreateJSONRequest("POST", host, http.NewJSONResult(result))
	if err != nil {
		return
	}

	for k, v := range e.headers {
		req.Header.Set(k, v)
	}

	if e.username != "" {
		req.SetBasicAuth(e.username, e.password)
	} else if e.apiKey != "" {
		req.Header.Add("Authorization", "ApiKey "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new elasticsearch.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Index,
		options.Username,
		options.Password,
		options.ApiKey,
		options.Rotation,
		options.CustomFields,
		options.Headers,
		options.HTTPClient,
		options.TypelessApi,
	}
}
