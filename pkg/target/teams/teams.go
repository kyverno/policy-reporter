package teams

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
	"github.com/fjogeleit/policy-reporter/pkg/target/helper"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type section struct {
	Title    string `json:"activityTitle"`
	SubTitle string `json:"activitySubtitle"`
	Text     string `json:"text"`
	Facts    []fact `json:"facts,omitempty"`
}

type payload struct {
	Type       string    `json:"@type"`
	Context    string    `json:"@context"`
	Summary    string    `json:"summary,omitempty"`
	ThemeColor string    `json:"themeColor,omitempty"`
	Sections   []section `json:"sections"`
}

func colorFromPriority(p report.Priority) string {
	if p == report.CriticalPriority {
		return "b80707"
	}
	if p == report.ErrorPriority {
		return "e20b0b"
	}
	if p == report.WarningPriority {
		return "f2c744"
	}
	if p == report.InfoPriority {
		return "36a64f"
	}

	return "68c2ff"
}

func newPayload(result report.Result) payload {
	facts := make([]fact, 0)

	facts = append(facts, fact{"Policy", result.Policy})

	if result.Rule != "" {
		facts = append(facts, fact{"Rule", result.Rule})
	}

	facts = append(facts, fact{"Priority", result.Priority.String()})

	if result.Category != "" {
		facts = append(facts, fact{"Category", result.Category})
	}
	if result.Severity != "" {
		facts = append(facts, fact{"Severity", result.Severity})
	}
	res := report.Resource{}
	if result.Resource.UID != "" {
		res = result.Resource
	}

	if res.UID != "" {
		facts = append(facts, fact{"Kind", res.Kind})
		facts = append(facts, fact{"Name", res.Name})
		facts = append(facts, fact{"UID", res.UID})
		if res.Namespace != "" {
			facts = append(facts, fact{"Namespace", res.Namespace})
		}
		facts = append(facts, fact{"API Version", res.APIVersion})
	}

	for property, value := range result.Properties {
		facts = append(facts, fact{strings.Title(property), value})
	}

	timestamp := time.Now()
	if !result.Timestamp.IsZero() {
		timestamp = result.Timestamp
	}

	sections := make([]section, 0, 1)
	sections = append(sections, section{
		Title:    "New Policy Report Result",
		SubTitle: timestamp.Format(time.RFC3339),
		Text:     result.Message,
		Facts:    facts,
	})

	return payload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		Summary:    result.Message,
		ThemeColor: colorFromPriority(result.Priority),
		Sections:   sections,
	}
}

type client struct {
	webhook               string
	minimumPriority       string
	skipExistingOnStartup bool
	client                httpClient
}

func (s *client) Send(result report.Result) {
	if result.Priority < report.NewPriority(s.minimumPriority) {
		return
	}

	payload := newPayload(result)
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] TEAMS : %v\n", err.Error())
		return
	}

	req, err := http.NewRequest("POST", s.webhook, body)
	if err != nil {
		log.Printf("[ERROR] TEAMS : %v\n", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("User-Agent", "Policy-Reporter")

	resp, err := s.client.Do(req)
	helper.HandleHTTPResponse("TEAMS", resp, err)
}

func (s *client) SkipExistingOnStartup() bool {
	return s.skipExistingOnStartup
}

func (s *client) Name() string {
	return "Teams"
}

func (s *client) MinimumPriority() string {
	return s.minimumPriority
}

// NewClient creates a new teams.client to send Results to MS Teams
func NewClient(host, minimumPriority string, skipExistingOnStartup bool, httpClient httpClient) target.Client {
	return &client{
		host,
		minimumPriority,
		skipExistingOnStartup,
		httpClient,
	}
}
