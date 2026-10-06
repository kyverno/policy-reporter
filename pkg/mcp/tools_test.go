package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func subject(kind, name, ns, uid string) []corev1.ObjectReference {
	return []corev1.ObjectReference{{APIVersion: "v1", Kind: kind, Name: name, Namespace: ns, UID: k8stypes.UID(uid)}}
}

func res(source, category, policy, rule string, status v1alpha1.Result, severity v1alpha1.ResultSeverity, subjects []corev1.ObjectReference) v1alpha1.ReportResult {
	return v1alpha1.ReportResult{
		Description: policy + " " + rule,
		Result:      status,
		Scored:      true,
		Policy:      policy,
		Rule:        rule,
		Timestamp:   v1.Timestamp{Seconds: 1614093000},
		Source:      source,
		Category:    category,
		Severity:    severity,
		Subjects:    subjects,
		Properties:  map[string]string{"k": "v"},
	}
}

func newTestStore(t *testing.T) *db.Store {
	t.Helper()

	ctx := context.Background()

	sqlDB, err := db.NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	store, err := db.NewStore(sqlDB, "test")
	require.NoError(t, err)
	require.NoError(t, store.PrepareDatabase(ctx))

	recon := result.NewReconditioner(nil)

	reports := []openreports.ReportInterface{
		&openreports.ReportAdapter{Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{Name: "report-a", Namespace: "team-a"},
			Results: []v1alpha1.ReportResult{
				res("kyverno", "security", "require-labels", "check-label", v1alpha2.StatusFail, v1alpha2.SeverityHigh, subject("Deployment", "nginx", "team-a", "uid-1")),
				res("kyverno", "security", "disallow-latest", "no-latest", v1alpha2.StatusPass, v1alpha2.SeverityMedium, subject("Deployment", "nginx", "team-a", "uid-1")),
				res("trivy", "vulnerability", "CVE-1", "scan", v1alpha2.StatusWarn, v1alpha2.SeverityLow, subject("Pod", "api", "team-a", "uid-2")),
			},
		}},
		&openreports.ReportAdapter{Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{Name: "report-b", Namespace: "team-b"},
			Results: []v1alpha1.ReportResult{
				res("kyverno", "security", "require-labels", "check-label", v1alpha2.StatusPass, v1alpha2.SeverityHigh, subject("Deployment", "web", "team-b", "uid-3")),
			},
		}},
		&openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{
			ObjectMeta: v1.ObjectMeta{Name: "cluster-report"},
			Results: []v1alpha1.ReportResult{
				res("kyverno", "security", "require-labels", "ns-label", v1alpha2.StatusFail, v1alpha2.SeverityCritical, subject("Namespace", "team-a", "", "uid-4")),
			},
		}},
	}

	for _, r := range reports {
		require.NoError(t, store.Add(ctx, recon.Prepare(r)))
	}

	return store
}

func TestListResourceComplianceHandler(t *testing.T) {
	handler := listResourceComplianceHandler(newTestStore(t))

	t.Run("single namespace", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"team-a"}},
		})
		require.NoError(t, err)

		assert.Equal(t, 2, resp.Total)
		require.Len(t, resp.Resources, 2)

		byName := map[string]ResourceCompliance{}
		for _, r := range resp.Resources {
			assert.Equal(t, "team-a", r.Namespace)
			byName[r.Name] = r
		}

		require.Contains(t, byName, "nginx")
		assert.Equal(t, "Deployment", byName["nginx"].Kind)
		assert.Equal(t, 1, byName["nginx"].Fail)
		assert.Equal(t, 1, byName["nginx"].Pass)
		assert.Equal(t, 1, byName["nginx"].High)
		assert.Equal(t, 1, byName["nginx"].Medium)

		require.Contains(t, byName, "api")
		assert.Equal(t, 1, byName["api"].Warn)
	})

	t.Run("multiple namespaces", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"team-a", "team-b"}},
		})
		require.NoError(t, err)
		assert.Equal(t, 3, resp.Total)
	})

	t.Run("source filter", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"team-a"}, Sources: []string{"trivy"}},
		})
		require.NoError(t, err)
		require.Len(t, resp.Resources, 1)
		assert.Equal(t, "api", resp.Resources[0].Name)
	})

	t.Run("unknown namespace", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"missing"}},
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Resources)
		assert.NotNil(t, resp.Resources)
		assert.Zero(t, resp.Total)
	})
}

func TestGetNamespaceComplianceSummaryHandler(t *testing.T) {
	handler := getNamespaceComplianceSummaryHandler(newTestStore(t))

	find := func(list []NamespaceCompliance, source, ns string) (NamespaceCompliance, bool) {
		for _, s := range list {
			if s.Source == source && s.Namespace == ns {
				return s, true
			}
		}
		return NamespaceCompliance{}, false
	}

	t.Run("all sources", func(t *testing.T) {
		list, err := handler(context.Background(), mcp.CallToolRequest{}, NamespaceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"team-a"}},
		})
		require.NoError(t, err)

		kyverno, ok := find(list, "kyverno", "team-a")
		require.True(t, ok)
		assert.Equal(t, 1, kyverno.Pass)
		assert.Equal(t, 1, kyverno.Fail)
		assert.Equal(t, 1, kyverno.High)
		assert.Equal(t, 1, kyverno.Medium)

		trivy, ok := find(list, "trivy", "team-a")
		require.True(t, ok)
		assert.Equal(t, 1, trivy.Warn)
		assert.Equal(t, 1, trivy.Low)
	})

	t.Run("source filter", func(t *testing.T) {
		list, err := handler(context.Background(), mcp.CallToolRequest{}, NamespaceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"team-a", "team-b"}, Sources: []string{"kyverno"}},
		})
		require.NoError(t, err)

		for _, s := range list {
			assert.Equal(t, "kyverno", s.Source)
		}

		b, ok := find(list, "kyverno", "team-b")
		require.True(t, ok)
		assert.Equal(t, 1, b.Pass)
		assert.Zero(t, b.Fail)
	})

	t.Run("no matches", func(t *testing.T) {
		list, err := handler(context.Background(), mcp.CallToolRequest{}, NamespaceComplianceRequest{
			ComplianceFilter{Namespaces: []string{"missing"}},
		})
		require.NoError(t, err)
		assert.Empty(t, list)
		assert.NotNil(t, list)
	})
}

func TestGetResourceComplianceResultsHandler(t *testing.T) {
	handler := getResourceComplianceResultsHandler(newTestStore(t))

	t.Run("resource results", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceResultsRequest{
			Namespace: "team-a", Kind: "Deployment", Name: "nginx",
		})
		require.NoError(t, err)

		assert.Equal(t, 2, resp.Total)
		policies := []string{}
		for _, r := range resp.Results {
			policies = append(policies, r.Policy)
			assert.Equal(t, "kyverno", r.Source)
			assert.Equal(t, "security", r.Category)
			assert.Equal(t, int64(1614093000), r.Timestamp)
			assert.Equal(t, map[string]string{"k": "v"}, r.Properties)
		}
		assert.ElementsMatch(t, []string{"require-labels", "disallow-latest"}, policies)
	})

	t.Run("search", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceResultsRequest{
			Namespace: "team-a", Kind: "Deployment", Name: "nginx", Search: "disallow-latest",
		})
		require.NoError(t, err)
		require.Len(t, resp.Results, 1)
		assert.Equal(t, "pass", resp.Results[0].Result)
		assert.Equal(t, "no-latest", resp.Results[0].Rule)
	})

	t.Run("unknown resource", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, ResourceComplianceResultsRequest{
			Namespace: "team-a", Kind: "Deployment", Name: "missing",
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Results)
		assert.NotNil(t, resp.Results)
		assert.Zero(t, resp.Total)
	})
}

func TestListPolicyResultsHandler(t *testing.T) {
	handler := listPolicyResultsHandler(newTestStore(t))

	t.Run("without namespace includes cluster scoped", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, PolicyResultsRequest{
			Policies: []string{"require-labels"},
		})
		require.NoError(t, err)

		assert.Equal(t, 3, resp.Total)

		var cluster []PolicyComplianceResult
		for _, r := range resp.Results {
			assert.Equal(t, "require-labels", r.Policy)
			if r.Namespace == "" {
				cluster = append(cluster, r)
			}
		}
		require.Len(t, cluster, 1)
		assert.Equal(t, "Namespace", cluster[0].Kind)
		assert.Equal(t, "critical", cluster[0].Severity)
	})

	t.Run("with namespace excludes cluster scoped", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, PolicyResultsRequest{
			Policies:   []string{"require-labels"},
			Namespaces: []string{"team-a"},
		})
		require.NoError(t, err)

		require.Len(t, resp.Results, 1)
		assert.Equal(t, "team-a", resp.Results[0].Namespace)
		assert.Equal(t, "nginx", resp.Results[0].Name)
		assert.Equal(t, "fail", resp.Results[0].Result)
	})

	t.Run("status and source filters", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, PolicyResultsRequest{
			Sources: []string{"kyverno"},
			Status:  []string{"fail"},
		})
		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		for _, r := range resp.Results {
			assert.Equal(t, "fail", r.Result)
		}
	})

	t.Run("pagination spans namespaced and cluster results", func(t *testing.T) {
		var names []string
		for page := 1; page <= 3; page++ {
			resp, err := handler(context.Background(), mcp.CallToolRequest{}, PolicyResultsRequest{
				Policies: []string{"require-labels"}, Page: page, PageSize: 2,
			})
			require.NoError(t, err)
			assert.Equal(t, 3, resp.Total)
			assert.Equal(t, 2, resp.TotalPages)
			assert.Equal(t, page, resp.Page)

			for _, r := range resp.Results {
				names = append(names, r.Kind+"/"+r.Name)
			}
			if page == 1 {
				assert.Len(t, resp.Results, 2)
			} else if page == 2 {
				require.Len(t, resp.Results, 1)
				assert.Equal(t, "Namespace", resp.Results[0].Kind)
			} else {
				assert.Empty(t, resp.Results)
			}
		}
		assert.Equal(t, []string{"Deployment/nginx", "Deployment/web", "Namespace/team-a"}, names)
	})

	t.Run("no matches", func(t *testing.T) {
		resp, err := handler(context.Background(), mcp.CallToolRequest{}, PolicyResultsRequest{
			Policies: []string{"missing"},
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Results)
		assert.NotNil(t, resp.Results)
	})
}

func TestToFilter(t *testing.T) {
	f := ComplianceFilter{
		Names: []string{"n"}, Namespaces: []string{"ns"}, Sources: []string{"s"}, Categories: []string{"c"},
		Kinds: []string{"k"}, Resources: []string{"r"}, Status: []string{"fail"}, Severities: []string{"high"}, Search: "q",
	}.toFilter()

	assert.Equal(t, []string{"ns"}, f.Namespaces)
	assert.Equal(t, []string{"r"}, f.ResourceAPIs)
	assert.Equal(t, "q", f.Search)

	p := PolicyResultsRequest{Names: []string{"n"}, Resources: []string{"r"}, Policies: []string{"p"}}.toFilter()
	assert.Equal(t, []string{"n"}, p.Resources)
	assert.Equal(t, []string{"r"}, p.ResourceAPIs)
	assert.Equal(t, []string{"p"}, p.Policies)
}
