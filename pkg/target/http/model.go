package http

import (
	"net/http"
	"time"
)

// Client Interface definition for HTTP based targets
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Resource JSON structure for HTTP Requests
type Resource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	UID        string `json:"uid"`
}

// Result JSON structure for HTTP Requests
type Result struct {
	Message           string            `json:"message"`
	Policy            string            `json:"policy"`
	Rule              string            `json:"rule"`
	Priority          string            `json:"priority"`
	Status            string            `json:"status"`
	Severity          string            `json:"severity,omitempty"`
	Category          string            `json:"category,omitempty"`
	Scored            bool              `json:"scored"`
	Properties        map[string]string `json:"properties,omitempty"`
	Resource          Resource          `json:"resource"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}
