package violations_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/email/violations"
)

func Test_Source(t *testing.T) {
	source := violations.NewSource("kyverno", true)
	t.Run("Source.ClusterReports", func(t *testing.T) {
		if !source.ClusterReports {
			t.Errorf("Expected Surce.ClusterReports to be true")
		}
	})
	t.Run("Source.AddClusterPassed", func(t *testing.T) {
		source.AddClusterPassed(3)

		if source.ClusterPassed != 3 {
			t.Errorf("Unexpected Summary: %d", source.ClusterPassed)
		}
	})
	t.Run("Source.AddClusterResults", func(t *testing.T) {
		source.AddClusterResults([]violations.Result{{
			Name:   "policy-reporter",
			Kind:   "Namespace",
			Policy: "require-label",
			Rule:   "require-label",
			Status: "fail",
		}, {
			Name:   "develop",
			Kind:   "Namespace",
			Policy: "require-label",
			Rule:   "require-label",
			Status: "fail",
		}})

		if len(source.ClusterResults["fail"]) != 2 {
			t.Errorf("Unexpected amount of failing Cluster Results: %d", len(source.ClusterResults["fail"]))
		}
	})
	t.Run("Source.AddNamespacedPassed", func(t *testing.T) {
		source.AddNamespacedPassed("test", 2)

		if source.NamespacePassed["test"] != 2 {
			t.Errorf("Unexpected amount of passed Results in Namespace: %d", source.NamespacePassed["test"])
		}

		source.AddNamespacedPassed("test", 3)

		if source.NamespacePassed["test"] != 5 {
			t.Errorf("Unexpected amount of passed Results in Namespace: %d", source.NamespacePassed["test"])
		}
	})

	t.Run("Source.AddNamespacedResults", func(t *testing.T) {
		source.AddNamespacedResults("test", []violations.Result{{
			Name:   "policy-reporter",
			Kind:   "Deployment",
			Policy: "require-label",
			Rule:   "require-label",
			Status: "fail",
		}})

		if len(source.NamespaceResults["test"]["fail"]) != 1 {
			t.Errorf("Unexpected amount of failing Results in namespace after init: %d", len(source.ClusterResults["fail"]))
		}
		if source.NamespaceResults["test"]["warn"] == nil {
			t.Errorf("Expected warn map is initialized")
		}
		if source.NamespaceResults["test"]["error"] == nil {
			t.Errorf("Expected warn map is initialized")
		}

		source.AddNamespacedResults("test", []violations.Result{{
			Name:   "policy-reporter-ui",
			Kind:   "Deployment",
			Policy: "require-label",
			Rule:   "require-label",
			Status: "fail",
		}})

		if len(source.NamespaceResults["test"]["fail"]) != 2 {
			t.Errorf("Unexpected amount of failing Results in namespace after add: %d", len(source.ClusterResults["fail"]))
		}
	})

	t.Run("Source.InitResults", func(t *testing.T) {
		source.InitResults("test")

		if source.NamespaceResults["test"]["fail"] == nil {
			t.Errorf("Expected fail map is initialized")
		}
		if source.NamespaceResults["test"]["warn"] == nil {
			t.Errorf("Expected warn map is initialized")
		}
		if source.NamespaceResults["test"]["error"] == nil {
			t.Errorf("Expected warn map is initialized")
		}
	})
}
