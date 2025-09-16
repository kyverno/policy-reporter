package jira_test

import (
	"reflect"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/target/jira"
)

func Test_Text(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"valid json", `{"key": "value"}`, map[string]any{"key": "value"}},
		{"text json", "text", "text"},
		{"slice json", `["text", "text"]`, []string{"text", "text"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jira.ConvertProperty(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
