package summary_test

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyreportv1alpha2 "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	crdfake "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_GenerateDataWithSingleSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(data) != 1 {
		t.Fatalf("expected one source got: %d", len(data))
	}

	source := data[0]
	if source.Name != "test" {
		t.Fatalf("expected source name 'test', got: %s", source.Name)
	}
	if source.ClusterScopeSummary.Fail != 4 {
		t.Fatalf("unexpected Summary Mapping: %d", source.ClusterScopeSummary.Fail)
	}
	if source.NamespaceScopeSummary["test"].Fail != 3 {
		t.Fatalf("unexpected Summary Mapping: %d", source.NamespaceScopeSummary["test"].Fail)
	}
}

func Test_NamespacedGeneratorRestrictsWGPolicyReports(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := crdfake.NewSimpleClientset(
		&policyreportv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "team-a", Namespace: "team-a"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 2},
			Results:    []policyreportv1alpha2.PolicyReportResult{{Source: "wg", Result: policyreportv1alpha2.StatusFail}},
		},
		&policyreportv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "team-b", Namespace: "team-b"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 7},
			Results:    []policyreportv1alpha2.PolicyReportResult{{Source: "wg", Result: policyreportv1alpha2.StatusFail}},
		},
		&policyreportv1alpha2.ClusterPolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "cluster"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 11},
			Results:    []policyreportv1alpha2.PolicyReportResult{{Source: "wg", Result: policyreportv1alpha2.StatusFail}},
		},
	)

	data, err := summary.NewNamespacedGenerator(nil, client.Wgpolicyk8sV1alpha2(), "team-a").GenerateData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 || data[0].NamespaceScopeSummary["team-a"].Fail != 2 {
		t.Fatalf("unexpected team-a summary: %#v", data)
	}
	if _, ok := data[0].NamespaceScopeSummary["team-b"]; ok {
		t.Fatal("team-b WGPolicyReport leaked into team-a report")
	}
	if data[0].ClusterScopeSummary.Fail != 0 {
		t.Fatal("cluster WGPolicyReport leaked into team-a report")
	}
}

func Test_GenerateDataWithMultipleSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(data) != 2 {
		t.Fatalf("expected two sources, got: %d", len(data))
	}
}

func Test_GenerateDataWithSourceFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{Include: []string{"test"}}), true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(data) != 1 {
		t.Fatalf("expected one source, got: %d", len(data))
	}
}

func Test_FilterSourcesBySource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = summary.FilterSources(data, email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{Include: []string{"Kyverno"}}), true)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}

func Test_FilterSourcesByNamespace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = summary.FilterSources(data, email.NewFilter(nil, validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), true)
	source := data[0]
	if source.Name != "Kyverno" {
		source = data[1]
	}

	if _, ok := source.NamespaceScopeSummary["kyverno"]; ok {
		t.Fatal("expected namespace kyverno to be excluded")
	}
}

func Test_RemoveEmptySource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = summary.FilterSources(data, email.NewFilter(nil, validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), false)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}

func Test_NamespacedGeneratorRestrictsReportsAndExcludesClusterScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, teamA, cluster := NewFakeClient()
	_, err := teamA.Create(ctx, fixtures.DefaultPolicyReport.Report.DeepCopy(), v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	teamBReport := fixtures.DefaultPolicyReport.Report.DeepCopy()
	teamBReport.Name = "team-b-report"
	teamBReport.Namespace = "team-b"
	teamBReport.Summary.Fail = 9
	_, err = client.Reports("team-b").Create(ctx, teamBReport, v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport.DeepCopy(), v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := summary.NewNamespacedGenerator(client, nil, "test").GenerateData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatalf("expected one source, got %d", len(data))
	}
	if data[0].ClusterReports {
		t.Fatal("expected cluster reports to be disabled")
	}
	if data[0].ClusterScopeSummary.Fail != 0 {
		t.Fatalf("expected no cluster-scoped failures, got %d", data[0].ClusterScopeSummary.Fail)
	}
	if len(data[0].NamespaceScopeSummary) != 1 {
		t.Fatalf("expected one namespace, got %d", len(data[0].NamespaceScopeSummary))
	}
	if data[0].NamespaceScopeSummary["test"].Fail != 3 {
		t.Fatalf("expected team-a fail count 3, got %d", data[0].NamespaceScopeSummary["test"].Fail)
	}
	if _, ok := data[0].NamespaceScopeSummary["team-b"]; ok {
		t.Fatal("team-b summary leaked into team-a report")
	}
}
