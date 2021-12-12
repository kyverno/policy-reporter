package elasticsearch

import (
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
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

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	target.BaseClient
	host     string
	index    string
	rotation Rotation
	client   httpClient
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

	req, err := helper.CreateJSONRequest(e.Name(), "POST", host, result)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := e.client.Do(req)
	helper.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Name() string {
	return "Elasticsearch"
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(host, index, rotation, minimumPriority string, sources []string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		target.NewBaseClient(minimumPriority, sources, skipExistingOnStartup),
		host,
		index,
		rotation,
		httpClient,
	}
}
