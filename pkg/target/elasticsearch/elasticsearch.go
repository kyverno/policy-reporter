package elasticsearch

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/helper"
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
	host                  string
	index                 string
	rotation              Rotation
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func (e *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(e.minimumPriority) {
		return
	}

	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(result); err != nil {
		log.Printf("[ERROR] ELASTICSEARCH : %v\n", err.Error())
		return
	}

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

	req, err := http.NewRequest("POST", host, body)
	if err != nil {
		log.Printf("[ERROR] ELASTICSEARCH : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := e.client.Do(req)
	helper.HandleHTTPResponse("ELASTICSEARCH", resp, err)
}

func (e *client) SkipExistingOnStartup() bool {
	return e.skipExistingOnStartup
}

func (e *client) Name() string {
	return "Elasticsearch"
}

func (e *client) MinimumPriority() string {
	return e.minimumPriority
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(host, index, rotation, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		host,
		index,
		Rotation(rotation),
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
