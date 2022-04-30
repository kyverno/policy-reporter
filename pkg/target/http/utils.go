package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kyverno/policy-reporter/pkg/report"
)

// CreateJSONRequest for the given configuration
func CreateJSONRequest(target, method, host string, payload interface{}) (*http.Request, error) {
	body := new(bytes.Buffer)

	json.NewEncoder(body).Encode(payload)

	req, err := http.NewRequest(method, host, body)
	if err != nil {
		log.Printf("[ERROR] %s : %v\n", target, err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Policy-Reporter")

	return req, nil
}

// ProcessHTTPResponse Logs Error or Success messages
func ProcessHTTPResponse(target string, resp *http.Response, err error) {
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		log.Printf("[ERROR] %s PUSH failed: %s\n", target, err.Error())
	} else if resp.StatusCode >= 400 {
		fmt.Printf("StatusCode: %d\n", resp.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		log.Printf("[ERROR] %s PUSH failed [%d]: %s\n", target, resp.StatusCode, buf.String())
	} else {
		log.Printf("[INFO] %s PUSH OK\n", target)
	}
}

func NewJSONResult(r *report.Result) Result {
	return Result{
		Message:  r.Message,
		Policy:   r.Policy,
		Rule:     r.Rule,
		Priority: r.Priority.String(),
		Status:   r.Status,
		Severity: r.Severity,
		Category: r.Category,
		Scored:   r.Scored,
		Resource: Resource{
			Namespace:  r.Resource.Namespace,
			APIVersion: r.Resource.APIVersion,
			Kind:       r.Resource.Kind,
			Name:       r.Resource.Name,
			UID:        r.Resource.UID,
		},
		CreationTimestamp: r.Timestamp,
	}
}
