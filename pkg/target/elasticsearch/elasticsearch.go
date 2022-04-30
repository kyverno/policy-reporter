package elasticsearch

import (
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

// Rotation Enum
type Rotation = string

// Elasticsearch Index Rotation
const (
	None     Rotation = "none"
	Dayli    Rotation = "dayli"
	Monthly  Rotation = "monthly"
	Annually Rotation = "annually"
)

type client struct {
	target.BaseClient
	host     string
	index    string
	rotation Rotation
	client   http.Client
}

func (e *client) Send(result *report.Result) {
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

	req, err := http.CreateJSONRequest(e.Name(), "POST", host, result)
	if err != nil {
		return
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(name, host, index, rotation string, skipExistingOnStartup bool, filter *target.Filter, httpClient http.Client) target.Client {
	return &client{
		target.NewBaseClient(name, skipExistingOnStartup, filter),
		host,
		index,
		rotation,
		httpClient,
	}
}
