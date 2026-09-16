package send_test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/cmd"
)

type capturedMail struct {
	Message struct {
		Subject string `json:"subject"`
		Body    struct {
			Content string `json:"content"`
		} `json:"body"`
		To []struct {
			Email struct {
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"toRecipients"`
		Attachments []struct {
			Type        string `json:"@odata.type"`
			Name        string `json:"name"`
			ContentType string `json:"contentType"`
			Content     []byte `json:"contentBytes"`
		} `json:"attachments"`
	} `json:"message"`
}

func reportFixture(name, ns, source string, pass int) map[string]any {
	fixture := map[string]any{
		"metadata": map[string]any{"name": name, "namespace": ns},
		"source":   source,
		"summary":  map[string]int{"pass": pass, "fail": 1, "warn": 0, "error": 0, "skip": 0},
		"results":  []any{map[string]any{"source": source, "policy": "require-label", "rule": "app-label", "result": "fail", "severity": "medium", "message": "Missing label, \"app\"\n雪", "resources": []any{map[string]any{"kind": "Deployment", "name": name, "namespace": ns}}}},
	}
	for range pass {
		fixture["results"] = append(fixture["results"].([]any), map[string]any{"source": source, "policy": "passed", "result": "pass"})
	}
	return fixture
}

func localReportsAPI(t *testing.T, empty bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		items := []any{}
		kind := "ReportList"
		switch r.URL.Path {
		case "/apis/openreports.io/v1alpha1":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"kind": "APIResourceList", "apiVersion": "v1", "groupVersion": "openreports.io/v1alpha1", "resources": []any{}}))
			return
		case "/apis/openreports.io/v1alpha1/reports":
			if !empty {
				items = []any{reportFixture("a1", "team-a", "kyverno", 2), reportFixture("a2", "team-a", "kyverno", 3), reportFixture("b1", "team-b", "trivy", 4)}
			}
		case "/apis/openreports.io/v1alpha1/clusterreports":
			kind = "ClusterReportList"
			if !empty {
				items = []any{reportFixture("cluster", "", "kyverno", 1)}
			}
		default:
			http.NotFound(w, r)
			return
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"apiVersion": "openreports.io/v1alpha1", "kind": kind, "items": items}))
	}))
}

func runReportCommand(t *testing.T, apiURL, reportType, transport, reportConfig string) string {
	t.Helper()
	dir := t.TempDir()
	logFile, err := os.Create(filepath.Join(dir, "stderr.log"))
	require.NoError(t, err)
	previous := os.Stderr
	os.Stderr = logFile
	defer func() { os.Stderr = previous; logFile.Close() }()
	kube := filepath.Join(dir, "kubeconfig")
	require.NoError(t, os.WriteFile(kube, []byte(fmt.Sprintf("apiVersion: v1\nkind: Config\nclusters:\n- name: local\n  cluster:\n    server: %s\ncontexts:\n- name: local\n  context:\n    cluster: local\ncurrent-context: local\n", apiURL)), 0o600))
	cfg := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfg, []byte("emailReports:\n  clusterName: local\n"+transport+"\n  "+reportType+":\n"+reportConfig), 0o600))
	cli := cmd.NewCLI("test")
	cli.SetArgs([]string{"send", reportType, "--config", cfg, "--kubeconfig", kube, "--auto-memory-enabled=false"})
	require.NoError(t, cli.Execute())
	_ = zap.L().Sync()
	logs, err := os.ReadFile(logFile.Name())
	require.NoError(t, err)
	return string(logs)
}

func TestCSVCommandGraph(t *testing.T) {
	for _, reportType := range []string{"summary", "violations"} {
		t.Run(reportType, func(t *testing.T) {
			api := localReportsAPI(t, false)
			defer api.Close()
			var mx sync.Mutex
			var messages []capturedMail
			graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.HasSuffix(r.URL.Path, "/token") {
					fmt.Fprint(w, `{"access_token":"local-token","token_type":"Bearer","expires_in":3600}`)
					return
				}
				require.Equal(t, "Bearer local-token", r.Header.Get("Authorization"))
				var m capturedMail
				require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
				mx.Lock()
				messages = append(messages, m)
				mx.Unlock()
				w.WriteHeader(http.StatusAccepted)
			}))
			defer graph.Close()
			transport := fmt.Sprintf("  graphAPI:\n    enabled: true\n    tenant: test\n    clientID: test\n    clientSecret: test\n    userID: test\n    azureADEndpoint: %s\n    graphEndpoint: %s\n", graph.URL, graph.URL)
			runReportCommand(t, api.URL, reportType, transport, `    to: [reader@example.test]
    attachmentFormat: csv
    channels:
    - to: [a@example.test]
      attachmentFormat: csv
      filter:
        disableClusterReports: true
        namespaces:
          include: [team-a]
    - to: [b@example.test]
      filter:
        disableClusterReports: true
        namespaces:
          include: [team-b]
`)
			require.Len(t, messages, 3)
			var m capturedMail
			for _, candidate := range messages {
				switch candidate.Message.To[0].Email.Address {
				case "reader@example.test":
					m = candidate
				case "a@example.test":
					require.Len(t, candidate.Message.Attachments, 1)
					rows, e := csv.NewReader(strings.NewReader(string(candidate.Message.Attachments[0].Content))).ReadAll()
					require.NoError(t, e)
					for _, row := range rows[1:] {
						require.Equal(t, "namespace", row[2])
						require.Equal(t, "team-a", row[3])
					}
				case "b@example.test":
					require.Empty(t, candidate.Message.Attachments)
					require.Contains(t, candidate.Message.Body.Content, "team-b")
					require.NotContains(t, candidate.Message.Body.Content, "team-a")
				default:
					t.Fatal("unexpected recipient")
				}
			}
			require.Len(t, m.Message.Attachments, 1, "CSV mode must deliver a real attachment instead of only the detailed body")
			a := m.Message.Attachments[0]
			require.Equal(t, "#microsoft.graph.fileAttachment", a.Type)
			require.Equal(t, reportType+".csv", a.Name)
			require.Equal(t, "text/csv; charset=utf-8", a.ContentType)
			rows, err := csv.NewReader(strings.NewReader(string(a.Content))).ReadAll()
			require.NoError(t, err)
			if reportType == "summary" {
				require.Equal(t, [][]string{{"Cluster", "Source", "Scope", "Namespace", "Pass", "Fail", "Warn", "Error", "Skip"}, {"local", "kyverno", "cluster", "", "1", "1", "0", "0", "0"}, {"local", "kyverno", "namespace", "team-a", "5", "2", "0", "0", "0"}, {"local", "trivy", "cluster", "", "0", "0", "0", "0", "0"}, {"local", "trivy", "namespace", "team-b", "4", "1", "0", "0", "0"}}, rows)
			} else {
				require.Len(t, rows, 5)
				require.Equal(t, []string{"Cluster", "Source", "Scope", "Namespace", "Policy", "Rule", "Kind", "Name", "Status", "Severity", "Message"}, rows[0])
				require.Equal(t, []string{"local", "kyverno", "namespace", "team-a", "require-label", "app-label", "Deployment", "a1", "fail", "medium", "Missing label, \"app\"\n雪"}, rows[2])
			}
			require.NotContains(t, m.Message.Body.Content, "require-label")
			require.Contains(t, m.Message.Body.Content, "attached")
		})
	}
}

func TestCSVCommandGraphRejection(t *testing.T) {
	for _, reportType := range []string{"summary", "violations"} {
		t.Run(reportType, func(t *testing.T) {
			api := localReportsAPI(t, false)
			defer api.Close()
			graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.HasSuffix(r.URL.Path, "/token") {
					fmt.Fprint(w, `{"access_token":"local","token_type":"Bearer","expires_in":3600}`)
					return
				}
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				fmt.Fprint(w, `{"error":{"message":"attachment too large"}}`)
			}))
			defer graph.Close()
			transport := fmt.Sprintf("  graphAPI:\n    enabled: true\n    tenant: test\n    clientID: test\n    clientSecret: test\n    userID: test\n    azureADEndpoint: %s\n    graphEndpoint: %s\n", graph.URL, graph.URL)
			logs := runReportCommand(t, api.URL, reportType, transport, "    to: [reader@example.test]\n    attachmentFormat: csv\n")
			require.Contains(t, logs, "failed to send report")
			require.Contains(t, logs, "413")
			require.NotContains(t, logs, "email sent to")
		})
	}
}
