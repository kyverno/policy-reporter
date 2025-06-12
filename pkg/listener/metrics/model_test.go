package metrics_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
)

func Test_LabelMappings(t *testing.T) {
	results := map[string]string{}
	res := fixtures.FailPodResult.GetResource()

	metrics.LabelGeneratorMapping["namespace"](results, preport, fixtures.FailPodResult)
	if val, ok := results["namespace"]; !ok && val != preport.Namespace {
		t.Errorf("expected result for namespace label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["report"](results, preport, fixtures.FailPodResult)
	if val, ok := results["report"]; !ok && val != preport.Name {
		t.Errorf("expected result for report label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["policy"](results, preport, fixtures.FailPodResult)
	if val, ok := results["report"]; !ok && val != fixtures.FailPodResult.Policy {
		t.Errorf("expected result for policy label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["rule"](results, preport, fixtures.FailPodResult)
	if val, ok := results["rule"]; !ok && val != fixtures.FailPodResult.Rule {
		t.Errorf("expected result for rule label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["kind"](results, preport, fixtures.FailPodResult)
	if val, ok := results["kind"]; !ok && val != res.Kind {
		t.Errorf("expected result for kind label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["name"](results, preport, fixtures.FailPodResult)
	if val, ok := results["name"]; !ok && val != res.Name {
		t.Errorf("expected result for name label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["severity"](results, preport, fixtures.FailPodResult)
	if val, ok := results["severity"]; !ok && val != string(fixtures.FailPodResult.Severity) {
		t.Errorf("expected result for severity label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["category"](results, preport, fixtures.FailPodResult)
	if val, ok := results["category"]; !ok && val != fixtures.FailPodResult.Category {
		t.Errorf("expected result for category label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["source"](results, preport, fixtures.FailPodResult)
	if val, ok := results["source"]; !ok && val != fixtures.FailPodResult.Source {
		t.Errorf("expected result for source label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["status"](results, preport, fixtures.FailPodResult)
	if val, ok := results["status"]; !ok && val != string(fixtures.FailPodResult.Result) {
		t.Errorf("expected result for status label not found: %s", val)
	}

	metrics.LabelGeneratorMapping["name"](results, preport, fixtures.TrivyResult)
	if val, ok := results["name"]; !ok && val != "" {
		t.Errorf("expected empty name without resource, got: %s", val)
	}
	metrics.LabelGeneratorMapping["kind"](results, preport, fixtures.TrivyResult)
	if val, ok := results["kind"]; !ok && val != "" {
		t.Errorf("expected empty name without resource, got: %s", val)
	}
	metrics.LabelGeneratorMapping["message"](results, preport, fixtures.FailPodResult)
	if val, ok := results["namespace"]; !ok && val != fixtures.FailPodResult.Description {
		t.Errorf("expected result for message label not found: %s", val)
	}
}
