package database_test

import (
	"context"
	"database/sql"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
)

var pagination = v1.Pagination{Page: 1, Offset: 20, Direction: "ASC", SortBy: []string{"resource_name"}}

var polrPagination = v1.Pagination{Page: 1, Offset: 20, Direction: "ASC", SortBy: []string{"namespace"}}

var preport = &v1alpha2.PolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		Labels:            map[string]string{"app": "policy-reporter", "scope": "namespaced"},
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var dreport = &v1alpha2.PolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		Labels:            map[string]string{"app": "policy-reporter", "scope": "namespaced"},
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailResult, fixtures.FailPodResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var ureport = &v1alpha2.PolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		Labels:            map[string]string{"app": "policy-reporter", "owner": "team-a", "scope": "namespaced"},
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.PassPodResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1, Pass: 1},
}

var creport = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "cpolr",
		Labels:            map[string]string{"app": "policy-reporter", "scope": "cluster"},
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.PassNamespaceResult, fixtures.FailNamespaceResult},
	Summary: v1alpha2.PolicyReportSummary{},
}

var scopeReport = &v1alpha2.PolicyReport{
	ObjectMeta: metav1.ObjectMeta{
		Name:              "polr-scope-test",
		Namespace:         "test",
		CreationTimestamp: metav1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{fixtures.ScopeResult},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1, Pass: 0},
	Scope: &corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

func Test_PolicyReportStore(t *testing.T) {
	db, err := database.NewSQLiteDB("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	store, _ := database.NewStore(db, "develop")
	store.PrepareDatabase(ctx)

	t.Run("Add/Get/Update PolicyReport", func(t *testing.T) {
		_, err := store.Get(ctx, preport.GetID())
		if err != sql.ErrNoRows {
			t.Fatalf("Should not be found in empty Store")
		}

		err = store.Add(ctx, preport)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		polr, err := store.Get(ctx, preport.GetID())
		if err != nil {
			t.Fatalf("Should found policy reporter after adding: %v", err)
		}

		if len(polr.GetResults()) == 0 {
			t.Fatalf("Failed to load PolicyReportResults: %v", err)
		}

		err = store.Update(ctx, ureport)
		if err != nil {
			t.Fatalf("Failed to update policy report: %v", err)
		}

		r2, _ := store.Get(ctx, ureport.GetID())
		if r2.GetSummary().Pass != 1 {
			t.Errorf("Expected 1 Passed Results in GetSummary() after Update")
		}

		if r2.GetLabels()["owner"] != "team-a" {
			t.Errorf("Expected Labels are updated")
		}
	})

	t.Run("Add/Get PolicyReport with ScopeResource", func(t *testing.T) {
		_, err := store.Get(ctx, scopeReport.GetID())
		if err != sql.ErrNoRows {
			t.Fatalf("Should not be found in empty Store")
		}

		err = store.Add(ctx, scopeReport)
		if err != nil {
			t.Fatalf("Unexpected add error: %s", err)
		}

		rep, err := store.Get(ctx, scopeReport.GetID())
		if err != nil {
			t.Error("Should be found in Store after adding report to the store")
		}
		if len(rep.GetResults()) == 0 {
			t.Fatal("Exptected at least one result on the report")
		}
		res := rep.GetResults()[0]
		if !res.HasResource() {
			t.Error("Expected scope resource set as result resource")
		}

		store.Remove(ctx, rep.GetID())
	})

	t.Run("Add/Get ClusterPolicyReport", func(t *testing.T) {
		_, err := store.Get(ctx, creport.GetID())
		if err != sql.ErrNoRows {
			t.Fatalf("Should not be found in empty Store")
		}

		err = store.Add(ctx, creport)
		if err != nil {
			t.Fatalf("Failed to persist ClusterPolicyReport: %v", err)
		}

		_, err = store.Get(ctx, creport.GetID())
		if err != nil {
			t.Fatalf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("FetchPolicyReports", func(t *testing.T) {
		items, err := store.FetchPolicyReports(ctx, v1.Filter{Namespaces: []string{"test"}, ReportLabel: map[string]string{"scope": "namespaced"}}, polrPagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return one policy report, got %d", len(items))
		}
	})

	t.Run("CountPolicyReports", func(t *testing.T) {
		count, err := store.CountPolicyReports(ctx, v1.Filter{Namespaces: []string{"test"}, ReportLabel: map[string]string{"scope": "namespaced"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 1 {
			t.Fatalf("Should return one policy report, got %d", count)
		}
	})

	t.Run("NamespacedGetLabels()", func(t *testing.T) {
		items, err := store.FetchNamespacedReportLabels(ctx, v1.Filter{Sources: []string{"Kyverno"}, Namespaces: []string{"test"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 3 {
			t.Fatalf("Should return 3 GetLabels() results")
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
	t.Run("FetchClusterReports", func(t *testing.T) {
		items, err := store.FetchClusterPolicyReports(ctx, v1.Filter{ReportLabel: map[string]string{"scope": "cluster"}}, polrPagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return one policy report, got %d", len(items))
		}
	})

	t.Run("CountClusterReports", func(t *testing.T) {
		items, err := store.CountClusterPolicyReports(ctx, v1.Filter{ReportLabel: map[string]string{"scope": "cluster"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if items != 1 {
			t.Fatalf("Should return one policy report, got %d", items)
		}
	})

	t.Run("ClusterGetLabels()", func(t *testing.T) {
		items, err := store.FetchClusterReportLabels(ctx, v1.Filter{Sources: []string{"Kyverno"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 GetLabels() results")
		}

		if len(items["scope"]) != 1 && items["scope"][0] != "cluster" {
			t.Fatalf("Should return cluster as scope value")
		}

		if len(items["app"]) != 1 && items["app"][0] != "policy-reporter" {
			t.Fatalf("Should return policy-reporter as app value")
		}
	})

	t.Run("FetchClusterPolicies", func(t *testing.T) {
		items, err := store.FetchClusterPolicies(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should Find 1 cluster scoped policy, found %d", len(items))
		}
		if items[0] != "require-ns-GetLabels()" {
			t.Fatalf("Should return 'require-ns-GetLabels()' policy")
		}
	})

	t.Run("FetchClusterRules", func(t *testing.T) {
		items, err := store.FetchClusterRules(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should Find 1 cluster scoped rule, found %d", len(items))
		}
		if items[0] != "check-for-GetLabels()-on-namespace" {
			t.Fatalf("Should return 'check-for-GetLabels()-on-namespace' rule")
		}
	})

	t.Run("FetchNamespacedPolicies", func(t *testing.T) {
		items, err := store.FetchNamespacedPolicies(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
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
		items, err := store.FetchNamespacedRules(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should find 1 namespace scoped policy, found %d", len(items))
		}
		if items[0] != "autogen-check-for-requests-and-limits" {
			t.Fatalf("Should return 'require-requests-and-limits-required' policy")
		}
	})

	t.Run("FetchNamespacedResources", func(t *testing.T) {
		items, err := store.FetchNamespacedResources(ctx, v1.Filter{Sources: []string{"Kyverno"}, Kinds: []string{"Pod"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should find 1 distinct resource with namespace Scope, got %d", len(items))
		}
		if items[0].Name != "nginx" {
			t.Errorf("Should return 'nginx' as first result, got %s", items[0].Name)
		}
	})

	t.Run("FetchClusterResources", func(t *testing.T) {
		items, err := store.FetchClusterResources(ctx, v1.Filter{Sources: []string{"Kyverno"}, Kinds: []string{"Namespace"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
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

	t.Run("FetchClusterSources", func(t *testing.T) {
		items, err := store.FetchClusterSources(ctx)
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
		items, err := store.FetchNamespacedSources(ctx)
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

	t.Run("FetchNamespaces", func(t *testing.T) {
		items, err := store.FetchNamespaces(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatal("Should find 1 Namespace")
		}
		if items[0] != "test" {
			t.Errorf("Should return test namespace")
		}
	})

	t.Run("FetchNamespacedStatusCounts", func(t *testing.T) {
		items, err := store.FetchNamespacedStatusCounts(ctx, v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 5 {
			t.Fatalf("Should include 1 item per possible status")
		}

		var passed v1.NamespacedStatusCount
		var failed v1.NamespacedStatusCount
		for _, item := range items {
			if item.Status == v1alpha2.StatusPass {
				passed = item
			}
			if item.Status == v1alpha2.StatusFail {
				failed = item
			}
		}

		if passed.Status != v1alpha2.StatusPass {
			t.Errorf("Expected Pass Counts as first item")
		}
		if passed.Items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}

		if failed.Status != v1alpha2.StatusFail {
			t.Errorf("Expected Pass Counts as first item")
		}
		if failed.Items[0].Count != 1 {
			t.Errorf("Expected count to be one for fail")
		}
	})

	t.Run("FetchNamespacedStatusCounts with StatusFilter", func(t *testing.T) {
		items, err := store.FetchNamespacedStatusCounts(ctx, v1.Filter{Status: []string{v1alpha2.StatusPass}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should have only 1 item for pass counts")
		}
		if items[0].Status != v1alpha2.StatusPass {
			t.Errorf("Expected Pass Counts")
		}
		if items[0].Items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
	})

	t.Run("FetchClusterStatusCounts", func(t *testing.T) {
		items, err := store.FetchClusterStatusCounts(ctx, v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		var passed v1.StatusCount
		var failed v1.StatusCount
		for _, item := range items {
			if item.Status == v1alpha2.StatusPass {
				passed = item
			}
			if item.Status == v1alpha2.StatusFail {
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

	t.Run("FetchClusterStatusCounts with StatusFilter", func(t *testing.T) {
		items, err := store.FetchClusterStatusCounts(ctx, v1.Filter{Status: []string{v1alpha2.StatusPass}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Fatalf("Should have only 1 item for pass counts")
		}
		if items[0].Status != v1alpha2.StatusPass {
			t.Errorf("Expected Pass Counts")
		}
		if items[0].Count != 1 {
			t.Errorf("Expected count to be one for pass")
		}
	})

	t.Run("FetchRuleStatusCounts", func(t *testing.T) {
		items, err := store.FetchRuleStatusCounts(ctx, "require-requests-and-limits-required", "autogen-check-for-requests-and-limits")
		var passed v1.StatusCount
		var failed v1.StatusCount
		for _, item := range items {
			if item.Status == v1alpha2.StatusPass {
				passed = item
			}
			if item.Status == v1alpha2.StatusFail {
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

	t.Run("FetchNamespacedResults", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(ctx, v1.Filter{Namespaces: []string{"test"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 namespaced results")
		}
	})

	t.Run("FetchNamespacedResults with SeverityFilter", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(ctx, v1.Filter{Severities: []string{v1alpha2.SeverityHigh}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != v1alpha2.SeverityHigh {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("CountNamespacedResults", func(t *testing.T) {
		count, err := store.CountNamespacedResults(ctx, v1.Filter{ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 2 {
			t.Fatalf("Should return 2 namespaced result")
		}
	})

	t.Run("CountNamespacedResults with SeverityFilter", func(t *testing.T) {
		count, err := store.CountNamespacedResults(ctx, v1.Filter{Severities: []string{v1alpha2.SeverityHigh}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
	})

	t.Run("FetchNamespacedResults with SearchFilter::Severity", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(ctx, v1.Filter{Search: v1alpha2.SeverityHigh}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != v1alpha2.SeverityHigh {
			t.Fatalf("result with severity high expected")
		}
	})

	t.Run("FetchNamespacedResults with SearchFilter::Kind", func(t *testing.T) {
		items, err := store.FetchNamespacedResults(ctx, v1.Filter{Search: "deployment"}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result, got %d", len(items))
		}
		if items[0].Kind != "Deployment" {
			t.Fatalf("result with kind Deployment expected")
		}
	})

	t.Run("FetchClusterResults", func(t *testing.T) {
		items, err := store.FetchClusterResults(ctx, v1.Filter{Status: []string{v1alpha2.StatusPass, v1alpha2.StatusFail}, ReportLabel: map[string]string{"app": "policy-reporter"}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 2 {
			t.Fatalf("Should return 2 cluster results")
		}
	})

	t.Run("CountClusterResults", func(t *testing.T) {
		count, err := store.CountClusterResults(ctx, v1.Filter{Status: []string{v1alpha2.StatusPass, v1alpha2.StatusFail}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if count != 2 {
			t.Fatalf("Should return 2 cluster results")
		}
	})

	t.Run("FetchClusterResults with SeverityFilter", func(t *testing.T) {
		items, err := store.FetchClusterResults(ctx, v1.Filter{Severities: []string{v1alpha2.SeverityHigh}}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != v1alpha2.SeverityHigh {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("FetchClusterResults with SearchFilter", func(t *testing.T) {
		items, err := store.FetchClusterResults(ctx, v1.Filter{Search: v1alpha2.SeverityHigh}, pagination)
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}

		if len(items) != 1 {
			t.Fatalf("Should return 1 namespaced result")
		}
		if items[0].Severity != v1alpha2.SeverityHigh {
			t.Fatalf("result with severity high")
		}
	})

	t.Run("FetchNamespacedKinds", func(t *testing.T) {
		items, err := store.FetchNamespacedKinds(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
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
		items, err := store.FetchClusterKinds(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
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

	t.Run("FetchNamespacedCategories", func(t *testing.T) {
		items, err := store.FetchNamespacedCategories(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 2 {
			t.Errorf("Should find 2 categories, got %d", len(items))
		}
		if items[0] != "Best Practices" {
			t.Errorf("Should return 'Best Practices' as first category")
		}
	})

	t.Run("FetchClusterCategories", func(t *testing.T) {
		items, err := store.FetchClusterCategories(ctx, v1.Filter{Sources: []string{"Kyverno"}, ReportLabel: map[string]string{"app": "policy-reporter"}})
		if err != nil {
			t.Fatalf("Unexpected Error: %s", err)
		}
		if len(items) != 1 {
			t.Errorf("Should find 1 category, got %d", len(items))
		}
		if items[0] != "namespaces" {
			t.Errorf("Should return 'Best Practices' as first category, get '%s'", items[0])
		}
	})

	err = store.CleanUp(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup policy reports: %v", err)
	}
}
