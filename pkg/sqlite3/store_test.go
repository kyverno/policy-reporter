package sqlite3_test

import (
	"testing"
	"time"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/sqlite3"
)

var pagination = v1.Pagination{Page: 1, Offset: 20, Direction: "ASC", SortBy: []string{"resource_name"}}

var result1 = report.Result{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Category: "resources",
	Severity: report.High,
	Scored:   true,
	Source:   "Kyverno",
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

var result2 = report.Result{
	ID:       "124",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Pass,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Kyverno",
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	},
}

var cresult1 = report.Result{
	ID:       "125",
	Message:  "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:   "require-ns-labels",
	Rule:     "check-for-labels-on-namespace",
	Priority: report.ErrorPriority,
	Status:   report.Pass,
	Category: "namespaces",
	Severity: report.Medium,
	Scored:   true,
	Source:   "Kyverno",
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
	},
}

var cresult2 = report.Result{
	ID:       "126",
	Message:  "validation error: The label `test` is required. Rule check-for-labels-on-namespace",
	Policy:   "require-ns-labels",
	Rule:     "check-for-labels-on-namespace",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Category: "namespaces",
	Severity: report.High,
	Scored:   true,
	Source:   "Kyverno",
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "dev",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
	},
}

var preport = report.PolicyReport{
	ID:                report.GeneratePolicyReportID("polr-test", "test"),
	Name:              "polr-test",
	Namespace:         "test",
	Labels:            map[string]string{"app": "policy-reporter", "scope": "namespaced"},
	Results:           []report.Result{result1},
	Summary:           report.Summary{Fail: 1},
	CreationTimestamp: time.Now(),
}

var dreport = report.PolicyReport{
	ID:                report.GeneratePolicyReportID("polr-test", "test"),
	Name:              "polr-test",
	Namespace:         "test",
	Labels:            map[string]string{"app": "policy-reporter", "scope": "namespaced"},
	Results:           []report.Result{result1, result1, result2},
	Summary:           report.Summary{Fail: 1},
	CreationTimestamp: time.Now(),
}

var ureport = report.PolicyReport{
	ID:                report.GeneratePolicyReportID("polr-test", "test"),
	Name:              "polr-test",
	Namespace:         "test",
	Labels:            map[string]string{"app": "policy-reporter", "owner": "team-a", "scope": "namespaced"},
	Results:           []report.Result{result1, result2},
	Summary:           report.Summary{Fail: 1, Pass: 1},
	CreationTimestamp: time.Now(),
}

var creport = report.PolicyReport{
	ID:                report.GeneratePolicyReportID("cpolr", ""),
	Name:              "cpolr",
	Labels:            map[string]string{"app": "policy-reporter", "scope": "cluster"},
	Results:           []report.Result{cresult1, cresult2},
	Summary:           report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_PolicyReportStore(t *testing.T) {
	db, _ := sqlite3.NewDatabase("test.db")
	defer db.Close()
	store, _ := sqlite3.NewPolicyReportStore(db)

	t.Run("Add/Get/Update PolicyReport", func(t *testing.T) {
		_, ok := store.Get(preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		err := store.Add(preport)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		r1, ok := store.Get(preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}
		if r1.Summary.Pass != 0 {
			t.Errorf("Expected 0 Passed Results in Summary")
		}

		if r1.Labels["app"] != "policy-reporter" {
			t.Errorf("Expected labels are persisted")
		}

		store.Update(ureport)
		r2, _ := store.Get(preport.GetIdentifier())
		if r2.Summary.Pass != 1 {
			t.Errorf("Expected 1 Passed Results in Summary after Update")
		}

		if r2.Labels["owner"] != "team-a" {
			t.Errorf("Expected labels are updated")
		}
	})
	t.Run("Add/Get ClusterPolicyReport", func(t *testing.T) {
		_, ok := store.Get(creport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(creport)
		_, ok = store.Get(creport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("FetchNamespacedKinds", func(t *testing.T) {
		items, err := store.FetchNamespacedKinds(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 2 {
			t.Fatalf("Should Find 2 Kinds with Namespace Scope")
		}
		if items[0] != "Deployment" {
			t.Errorf("Should return 'Deployment' as first result")
		}
		if items[1] != "Pod" {
			t.Errorf("Should return 'Pod' as second result")
		}
	})

	t.Run("FetchClusterKinds", func(t *testing.T) {
		items, err := store.FetchClusterKinds(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should find 1 kind with cluster scope")
		}
		if items[0] != "Namespace" {
			t.Errorf("Should return 'Namespace' as first result")
		}
	})

	t.Run("FetchNamespacedResources", func(t *testing.T) {
		items, err := store.FetchNamespacedResources(v1.Filter{Sources: []string{"kyverno"}, Kinds: []string{"pod"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 2 {
			t.Fatalf("Should Find 2 Resources with Namespace Scope")
		}
		if items[0].Name != "nginx" {
			t.Errorf("Should return 'nginx' as first result, got %s", items[0].Name)
		}
		if items[1].Name != "nginx" {
			t.Errorf("Should return 'nginx' as second result get %s", items[1].Name)
		}
	})

	t.Run("FetchClusterResources", func(t *testing.T) {
		items, err := store.FetchClusterResources(v1.Filter{Sources: []string{"kyverno"}, Kinds: []string{"namespace"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 2 {
			t.Fatalf("Should find 2 resources with cluster scope")
		}
		if items[0].Name != "dev" {
			t.Errorf("Should return 'test' as first result")
		}
		if items[1].Name != "test" {
			t.Errorf("Should return 'test' as second result")
		}
	})

	t.Run("FetchNamespacedStatusCounts", func(t *testing.T) {
		items, err := store.FetchNamespacedStatusCounts(v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 5 {
			t.Fatalf("Should include 1 item per possible status")
		}

		var passed v1.NamespacedStatusCount
		var failed v1.NamespacedStatusCount
		for _, item := range items {
			if item.Status == report.Pass {
				passed = item
			}
			if item.Status == report.Fail {
				failed = item
			}
		}

		if passed.Status != report.Pass {
			t.Errorf("Expected Pass Counts as first item")
		}
		if passed.Items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}

		if failed.Status != report.Fail {
			t.Errorf("Expected Pass Counts as first item")
		}
		if failed.Items[0].Count != 1 {
			t.Errorf("Expected count to be one for fail")
		}
	})

	t.Run("FetchNamespacedStatusCounts with StatusFilter", func(t *testing.T) {
		items, err := store.FetchNamespacedStatusCounts(v1.Filter{Status: []string{report.Pass}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should have only 1 item for pass counts")
		}
		if items[0].Status != report.Pass {
			t.Errorf("Expected Pass Counts")
		}
		if items[0].Items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
	})

	t.Run("FetchRuleStatusCounts", func(t *testing.T) {
		items, err := store.FetchRuleStatusCounts("require-requests-and-limits-required", "autogen-check-for-requests-and-limits")
		var passed v1.StatusCount
		var failed v1.StatusCount
		for _, item := range items {
			if item.Status == report.Pass {
				passed = item
			}
			if item.Status == report.Fail {
				failed = item
			}
		}

		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if passed.Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}

		if failed.Count != 1 {
			t.Errorf("Expected count to be one for fail")
		}
	})

	t.Run("FetchStatusCounts", func(t *testing.T) {
		items, err := store.FetchStatusCounts(v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		var passed v1.StatusCount
		var failed v1.StatusCount
		for _, item := range items {
			if item.Status == report.Pass {
				passed = item
			}
			if item.Status == report.Fail {
				failed = item
			}
		}
		if len(items) != 5 {
			t.Fatalf("Should include 1 item per possible status")
		}
		if passed.Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
		if failed.Count != 1 {
			t.Errorf("Expected count to be one for fail")
		}
	})

	t.Run("FetchStatusCounts with StatusFilter", func(t *testing.T) {
		items, err := store.FetchStatusCounts(v1.Filter{Status: []string{report.Pass}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should have only 1 item for pass counts")
		}
		if items[0].Status != report.Pass {
			t.Errorf("Expected Pass Counts")
		}
		if items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
	})

	t.Run("FetchNamespacedResults", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(v1.Filter{Namespaces: []string{"test"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 namespaced results")
		}
	})

	t.Run("FetchNamespacedResults with SeverityFilter", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(v1.Filter{Severities: []string{report.High}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != report.High {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("CountNamespacedResults", func(t *testing.T) {
		count, err := store.CountNamespacedResults(v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 2 {
			t.Fatalf("Should return 2 namespaced result")
		}
	})

	t.Run("CountNamespacedResults with SeverityFilter", func(t *testing.T) {
		count, err := store.CountNamespacedResults(v1.Filter{Severities: []string{report.High}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
	})

	t.Run("FetchNamespacedResults with SearchFilter", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(v1.Filter{Search: report.High}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != report.High {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("FetchClusterResults", func(t *testing.T) {
		items, err := store.FetchClusterResults(v1.Filter{Status: []string{report.Pass, report.Fail}, ReportLabel: map[string]string{"app": "policy-reporter"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 cluster results")
		}
	})

	t.Run("CountClusterResults", func(t *testing.T) {
		count, err := store.CountClusterResults(v1.Filter{Status: []string{report.Pass, report.Fail}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 2 {
			t.Fatalf("Should return 2 cluster results")
		}
	})

	t.Run("FetchClusterResults with SeverityFilter", func(t *testing.T) {
		items, err := store.FetchClusterResults(v1.Filter{Severities: []string{report.High}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != report.High {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("FetchClusterResults with SearchFilter", func(t *testing.T) {
		items, err := store.FetchClusterResults(v1.Filter{Search: report.High}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != report.High {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("FetchStatusCounts with StatusFilter", func(t *testing.T) {
		items, err := store.FetchStatusCounts(v1.Filter{Status: []string{report.Pass}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should have only 1 item for pass counts")
		}
		if items[0].Status != report.Pass {
			t.Errorf("Expected Pass Counts")
		}
		if items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
	})

	t.Run("FetchNamespaces", func(t *testing.T) {
		items, err := store.FetchNamespaces(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 Namespace")
		}
		if items[0] != "test" {
			t.Errorf("Should return test namespace")
		}
	})

	t.Run("FetchCategories", func(t *testing.T) {
		items, err := store.FetchCategories(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 3 {
			t.Errorf("Should Find 2 Categories")
		}
		if items[0] != "Best Practices" {
			t.Errorf("Should return 'Best Practices' as first category")
		}
	})

	t.Run("FetchClusterPolicies", func(t *testing.T) {
		items, err := store.FetchClusterPolicies(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should Find 1 cluster scoped Policy")
		}
		if items[0] != "require-ns-labels" {
			t.Errorf("Should return 'require-ns-labels' policy")
		}
	})

	t.Run("FetchClusterRules", func(t *testing.T) {
		items, err := store.FetchClusterRules(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should Find 1 cluster scoped Policy")
		}
		if items[0] != "check-for-labels-on-namespace" {
			t.Errorf("Should return 'check-for-labels-on-namespace' rule")
		}
	})

	t.Run("FetchNamespacedPolicies", func(t *testing.T) {
		items, err := store.FetchNamespacedPolicies(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 namespace scoped policy")
		}
		if items[0] != "require-requests-and-limits-required" {
			t.Errorf("Should return 'require-requests-and-limits-required' policy")
		}
	})

	t.Run("FetchNamespacedRules", func(t *testing.T) {
		items, err := store.FetchNamespacedRules(v1.Filter{Sources: []string{"kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 namespace scoped policy")
		}
		if items[0] != "autogen-check-for-requests-and-limits" {
			t.Errorf("Should return 'require-requests-and-limits-required' policy")
		}
	})

	t.Run("FetchClusterSources", func(t *testing.T) {
		items, err := store.FetchClusterSources()
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 Source")
		}
		if items[0] != "Kyverno" {
			t.Errorf("Should return Kyverno")
		}
	})

	t.Run("FetchNamespacedSources", func(t *testing.T) {
		items, err := store.FetchNamespacedSources()
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 Source")
		}
		if items[0] != "Kyverno" {
			t.Errorf("Should return Kyverno")
		}
	})

	t.Run("NamespacedResults: ReportLabel Filter", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 namespaced results")
		}

		items, err = store.FetchNamespacedResults(v1.Filter{ReportLabel: map[string]string{"app": "not-exist"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 0 {
			t.Fatalf("Should return 0 namespaced results")
		}
	})

	t.Run("ClusterResults: ReportLabel Filter", func(t *testing.T) {
		items, err := store.FetchClusterResults(v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 namespaced results")
		}

		items, err = store.FetchClusterResults(v1.Filter{ReportLabel: map[string]string{"app": "not-exist"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 0 {
			t.Fatalf("Should return 0 namespaced results")
		}
	})

	t.Run("NamespacedLabels", func(t *testing.T) {
		items, err := store.FetchNamespacedReportLabels(v1.Filter{Sources: []string{"Kyverno"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 3 {
			t.Fatalf("Should return 3 labels results")
		}

		if len(items["scope"]) != 1 && items["scope"][0] != "namespaced" {
			t.Fatalf("Should return cluster as scope value")
		}

		if len(items["app"]) != 1 && items["app"][0] != "policy-reporter" {
			t.Fatalf("Should return policy-reporter as app value")
		}

		if len(items["owner"]) != 1 && items["owner"][0] != "team-a" {
			t.Fatalf("Should return policy-reporter as app value")
		}
	})

	t.Run("ClusterLabels", func(t *testing.T) {
		items, err := store.FetchClusterReportLabels(v1.Filter{Sources: []string{"Kyverno"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 labels results")
		}

		if len(items["scope"]) != 1 && items["scope"][0] != "cluster" {
			t.Fatalf("Should return cluster as scope value")
		}

		if len(items["app"]) != 1 && items["app"][0] != "policy-reporter" {
			t.Fatalf("Should return policy-reporter as app value")
		}
	})

	t.Run("Delete/Get", func(t *testing.T) {
		_, ok := store.Get(preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(preport.GetIdentifier())
		_, ok = store.Get(preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})

	t.Run("CleanUp", func(t *testing.T) {
		store.Add(preport)

		store.CleanUp()
		list, _ := store.FetchNamespacedResults(v1.Filter{}, v1.Pagination{Page: 0})
		if len(list) == 1 {
			t.Fatalf("Should have no results after CleanUp")
		}
	})

	t.Run("Insert duplicates", func(t *testing.T) {
		err := store.Add(dreport)
		if err != nil {
			t.Errorf("Should ignore duplicated ID: %s", err)
		}

		polr, ok := store.Get(preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		if len(polr.Results) != 2 {
			t.Errorf("Should ignore duplicated result")
		}

		store.Remove(dreport.GetIdentifier())
		_, ok = store.Get(dreport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})
}
