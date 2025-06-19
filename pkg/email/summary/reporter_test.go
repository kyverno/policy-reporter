package summary_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/email/summary"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
)

func Test_CreateReport(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeClient()

	_, _ = pClient.Create(ctx, fixtures.DefaultPolicyReport.Report, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, fixtures.EmptyPolicyReport.Report, v1.CreateOptions{})
	_, _ = client.Reports("kyverno").Create(ctx, fixtures.KyvernoPolicyReport, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, fixtures.ClusterPolicyReport.ClusterReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.EmptyClusterPolicyReport, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, fixtures.KyvernoClusterPolicyReport, v1.CreateOptions{})

	generator := summary.NewGenerator(client, nil, filter, true, true)
	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	path, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(path)

	reporter := summary.NewReporter("../../../templates", "Cluster", "Report")
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
	expected := "Report (summary) on Cluster from " + time.Now().Format("2006-01-02")
	if report.Title != expected {
		t.Fatalf("expected title to be '%s', got %s", expected, report.Title)
	}
	if report.Format != "html" {
		t.Fatal("expected format to be set")
	}
}
