package violations_test

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_GenerateDataWithSingleSource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.PassClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)

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

func Test_GenerateDataWithMultipleSource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.PassPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.PassClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)

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

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, email.NewFilter(validate.RuleSets{}, validate.RuleSets{Include: []string{"test"}}), true)

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

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(validate.RuleSets{}, validate.RuleSets{Include: []string{"Kyverno"}}), true)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}

func Test_FilterSourcesByNamespace(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), true)
	source := data[0]
	if source.Name != "Kyverno" {
		source = data[1]
	}

	if _, ok := source.NamespaceResults["kyverno"]; ok {
		t.Fatal("expected namespace kyverno to be excluded")
	}
}

func Test_RemoveEmptySource(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)

	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data = violations.FilterSources(data, email.NewFilter(validate.RuleSets{Exclude: []string{"kyverno"}}, validate.RuleSets{}), false)
	if len(data) != 1 {
		t.Fatalf("expected one source left, got: %d", len(data))
	}
}
