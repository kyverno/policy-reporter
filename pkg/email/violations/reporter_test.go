package violations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/email/violations"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_CreateReport(t *testing.T) {
	ctx := context.Background()

	client, pClient, cClient := NewFakeCilent()

	_, _ = pClient.Create(ctx, PolicyReportCRD, v1.CreateOptions{})
	_, _ = pClient.Create(ctx, EmptyPolicyReportCRD, v1.CreateOptions{})
	_, _ = client.PolicyReports("kyverno").Create(ctx, KyvernoPolicyReportCRD, v1.CreateOptions{})

	_, _ = cClient.Create(ctx, ClusterPolicyReportCRD, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, EmptyClusterPolicyReportCRD, v1.CreateOptions{})
	_, _ = cClient.Create(ctx, KyvernoClusterPolicyReportCRD, v1.CreateOptions{})

	generator := violations.NewGenerator(client, filter, true)
	data, err := generator.GenerateData(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	path, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(path)

	reporter := violations.NewReporter("../../../templates", "Cluster")
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
