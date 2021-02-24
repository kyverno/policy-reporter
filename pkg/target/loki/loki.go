package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
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

func newLokiPayload(result report.Result) payload {
	le := entry{Ts: time.Now().Format(time.RFC3339), Line: "[" + strings.ToUpper(result.Priority.String()) + "] " + result.Message}
	ls := stream{Entries: []entry{le}}

	res := report.Resource{}

	if len(result.Resources) > 0 {
		res = result.Resources[0]
	}

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
	if res.Kind != "" {
		labels = append(labels, "kind=\""+res.Kind+"\"")
		labels = append(labels, "name=\""+res.Name+"\"")
		labels = append(labels, "uid=\""+res.UID+"\"")
		labels = append(labels, "namespace=\""+res.Namespace+"\"")
	}

	ls.Labels = "{" + strings.Join(labels, ",") + "}"

	return payload{Streams: []stream{ls}}
}

type client struct {
	host            string
	minimumPriority string
	client          httpClient
}

func (l *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(l.minimumPriority) {
		return
	}

	payload := newLokiPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] : %v\n", err.Error())
	}

	req, err := http.NewRequest("POST", l.host, body)
	if err != nil {
		log.Printf("[ERROR] : %v\n", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Policy-API")

	resp, err := l.client.Do(req)
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

// NewClient creates a new loki.client to send Results to Loki
func NewClient(host, minimumPriority string, httpClient httpClient) target.Client {
	return &client{
		host + "/api/prom/push",
		minimumPriority,
		httpClient,
	}
}
