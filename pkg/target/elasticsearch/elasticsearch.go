package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
)

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
		log.Printf("[ERROR] : %v\n", err.Error())
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
		log.Printf("[ERROR] : %v\n", err.Error())
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := e.client.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		log.Printf("[ERROR] PUSH failed: %s\n", err.Error())
	} else if resp.StatusCode > 400 {
		fmt.Printf("StatusCode: %d\n", resp.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		log.Printf("[ERROR] PUSH failed [%d]: %s\n", resp.StatusCode, buf.String())
	} else {
		log.Println("[INFO] PUSH OK")
	}
}

func (e *client) SkipExistingOnStartup() bool {
	return e.skipExistingOnStartup
}

// NewClient creates a new loki.client to send Results to Loki
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
