package summary_test

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_GenerateDataWithSingleSource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)

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

func Test_GenerateDataWithMultipleSource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(data) != 2 {
		t.Fatalf("expected two sources, got: %d", len(data))
	}
}

func Test_GenerateDataWithSourceFilter(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, email.NewFilter(nil, validate.RuleSets{}, validate.RuleSets{Include: []string{"test"}}), true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(data) != 1 {
		t.Fatalf("expected one source, got: %d", len(data))
	}
}

func Test_FilterSourcesBySource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)

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
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)

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
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = summary.FilterSources(data, email.NewFilter(nil, validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), false)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}
