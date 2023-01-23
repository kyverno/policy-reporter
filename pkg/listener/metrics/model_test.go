package metrics_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
)

func Test_LabelMappings(t *testing.T) {
	results := map[string]string{}
	res := result1.GetResource()

	metrics.LabelGeneratorMapping["namespace"](results, preport, result1)
	if val, ok := results["namespace"]; !ok && val != preport.Namespace {
		t.Errorf("expected result for namespace label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["report"](results, preport, result1)
	if val, ok := results["report"]; !ok && val != preport.Name {
		t.Errorf("expected result for report label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["policy"](results, preport, result1)
	if val, ok := results["report"]; !ok && val != result1.Policy {
		t.Errorf("expected result for policy label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["rule"](results, preport, result1)
	if val, ok := results["rule"]; !ok && val != result1.Rule {
		t.Errorf("expected result for rule label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["kind"](results, preport, result1)
	if val, ok := results["kind"]; !ok && val != res.Kind {
		t.Errorf("expected result for kind label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["name"](results, preport, result1)
	if val, ok := results["name"]; !ok && val != res.Name {
		t.Errorf("expected result for name label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["severity"](results, preport, result1)
	if val, ok := results["severity"]; !ok && val != string(result1.Severity) {
		t.Errorf("expected result for severity label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["category"](results, preport, result1)
	if val, ok := results["category"]; !ok && val != result1.Category {
		t.Errorf("expected result for category label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["source"](results, preport, result1)
	if val, ok := results["source"]; !ok && val != result1.Source {
		t.Errorf("expected result for source label not found: %s", val)
	}
	metrics.LabelGeneratorMapping["status"](results, preport, result1)
	if val, ok := results["status"]; !ok && val != string(result1.Result) {
		t.Errorf("expected result for status label not found: %s", val)
	}
}
