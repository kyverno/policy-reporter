package payload

import (
	"fmt"
	"strings"
	"time"
)

type Value = []string

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values []Value           `json:"values"`
}

func (s *PolicyReportResultPayload) ToLoki() (Stream, error) {
	timestamp := time.Now()
	if s.Result.Timestamp.Seconds != 0 {
		timestamp = time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	}

	labels := map[string]string{
		"status":    string(s.Result.Result),
		"policy":    s.Result.Policy,
		"createdBy": "policy-reporter",
	}

	if s.Result.Rule != "" {
		labels["rule"] = s.Result.Rule
	}
	if s.Result.Category != "" {
		labels["category"] = s.Result.Category
	}
	if s.Result.Severity != "" {
		labels["severity"] = string(s.Result.Severity)
	}
	if s.Result.Source != "" {
		labels["source"] = s.Result.Source
	}
	if s.Result.HasResource() {
		res := s.Result.GetResource()
		if res.APIVersion != "" {
			labels["apiVersion"] = res.APIVersion
			labels["kind"] = res.Kind
			labels["name"] = res.Name
		}
		if res.UID != "" {
			labels["uid"] = string(res.UID)
		}
		if res.Namespace != "" {
			labels["namespace"] = res.Namespace
		}
	}

	for property, value := range s.Result.Properties {
		labels[keyReplacer.Replace(property)] = labelReplacer.Replace(value)
	}

	return Stream{
		Values: []Value{[]string{fmt.Sprintf("%v", timestamp.UnixNano()), "[" + strings.ToUpper(string(s.Result.Severity)) + "] " + s.Result.Message}},
		Stream: labels,
	}, nil
}
