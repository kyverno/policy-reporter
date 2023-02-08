package summary_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_CreateReport(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, filter, true)
	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	path, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(path)

	reporter := summary.NewReporter("../../../templates", "Cluster")
	report, err := reporter.Report(data, "html")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if report.Message == "" {
		t.Fatal("expected validate report message")
	}
	if report.ClusterName != "Cluster" {
		t.Fatal("expected clustername to be set")
	}
	if report.Format != "html" {
		t.Fatal("expected format to be set")
	}
}
