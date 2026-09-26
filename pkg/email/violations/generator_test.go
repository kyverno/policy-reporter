package violations_test

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyreportv1alpha2 "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	crdfake "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_GenerateDataWithSingleSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.PassClusterPolicyReport.ClusterReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, filter, true)

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
	if source.ClusterPassed != 1 {
		t.Fatalf("unexpected Summary Mapping: %d", source.ClusterPassed)
	}
	if len(source.NamespaceResults["test"]["fail"]) != 3 {
		t.Fatalf("unexpected Summary Mapping: %d", len(source.NamespaceResults["test"]["fail"]))
	}

	result := source.NamespaceResults["test"]["fail"][0]
	if result.Kind != "Deployment" {
		t.Fatalf("unexpected kind: %s", result.Kind)
	}
	if result.Name != "nginx" {
		t.Fatalf("unexpected name: %s", result.Kind)
	}
	if result.Policy != "required-label" {
		t.Fatalf("unexpected policy: %s", result.Kind)
	}
	if result.Rule != "app-label-required" {
		t.Fatalf("unexpected rule: %s", result.Kind)
	}
	if result.Status != "fail" {
		t.Fatalf("unexpected status: %s", result.Status)
	}

	result = source.NamespaceResults["test"]["fail"][2]
	if result.Rule != "app-label-required" {
		t.Fatalf("unexpected rule: %s", result.Rule)
	}
}

func Test_NamespacedGeneratorRestrictsWGPolicyReports(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := crdfake.NewSimpleClientset(
		&policyreportv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "team-a", Namespace: "team-a"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 1},
			Results: []policyreportv1alpha2.PolicyReportResult{{
				Source: "wg", Policy: "team-a-policy", Result: policyreportv1alpha2.StatusFail,
			}},
		},
		&policyreportv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "team-b", Namespace: "team-b"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 1},
			Results: []policyreportv1alpha2.PolicyReportResult{{
				Source: "wg", Policy: "team-b-policy", Result: policyreportv1alpha2.StatusFail,
			}},
		},
		&policyreportv1alpha2.ClusterPolicyReport{
			ObjectMeta: v1.ObjectMeta{Name: "cluster"},
			Summary:    policyreportv1alpha2.PolicyReportSummary{Fail: 1},
			Results: []policyreportv1alpha2.PolicyReportResult{{
				Source: "wg", Policy: "cluster-policy", Result: policyreportv1alpha2.StatusFail,
			}},
		},
	)

	data, err := violations.NewNamespacedGenerator(nil, client.Wgpolicyk8sV1alpha2(), "team-a").GenerateData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatalf("expected one source, got %d", len(data))
	}
	if _, ok := data[0].NamespaceResults["team-b"]; ok {
		t.Fatal("team-b WGPolicyReport leaked into team-a report")
	}
	for status, results := range data[0].ClusterResults {
		if len(results) != 0 {
			t.Fatalf("cluster WGPolicyReport leaked %d %s results", len(results), status)
		}
	}
}

func Test_GenerateDataWithMultipleSource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.PassPolicyReport, v1.CreateOptions{})
	_, _ = client.OpenreportsV1alpha1().Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.PassClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, filter, true)

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
	_, _ = client.OpenreportsV1alpha1().Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{Include: []string{"test"}}), true)

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
	_, _ = client.OpenreportsV1alpha1().Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{Include: []string{"Kyverno"}}), true)
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
	_, _ = client.OpenreportsV1alpha1().Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(nil, validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), true)
	source := data[0]
	if source.Name != "Kyverno" {
		source = data[1]
	}

	if _, ok := source.NamespaceResults["kyverno"]; ok {
		t.Fatal("expected namespace kyverno to be excluded")
	}
}

func Test_RemoveEmptySource(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.OpenreportsV1alpha1().Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client.OpenreportsV1alpha1(), nil, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(nil, validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), false)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}

func Test_NamespacedGeneratorRestrictsReportsAndExcludesClusterScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	clientset, teamA, cluster := NewFakeClient()
	_, err := teamA.Create(ctx, fixtures.DefaultPolicyReport.Report.DeepCopy(), v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	teamBReport := fixtures.DefaultPolicyReport.Report.DeepCopy()
	teamBReport.Name = "team-b-report"
	teamBReport.Namespace = "team-b"
	teamBReport.Results[0].Policy = "team-b-only-policy"
	_, err = clientset.OpenreportsV1alpha1().Reports("team-b").Create(ctx, teamBReport, v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport.DeepCopy(), v1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := violations.NewNamespacedGenerator(clientset.OpenreportsV1alpha1(), nil, "test").GenerateData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatalf("expected one source, got %d", len(data))
	}
	for status, results := range data[0].ClusterResults {
		if len(results) != 0 {
			t.Fatalf("expected no cluster-scoped %s results, got %d", status, len(results))
		}
	}
	if len(data[0].NamespaceResults) != 1 {
		t.Fatalf("expected one namespace, got %d", len(data[0].NamespaceResults))
	}
	if _, ok := data[0].NamespaceResults["team-b"]; ok {
		t.Fatal("team-b violations leaked into team-a report")
	}
}
