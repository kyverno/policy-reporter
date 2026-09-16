package kubernetes_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	or "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	orfake "github.com/openreports/reports-api/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metafake "k8s.io/client-go/metadata/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyverno/policy-reporter/pkg/api"
	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	prfake "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
	orclient "github.com/kyverno/policy-reporter/pkg/kubernetes/openreports"
	wgclient "github.com/kyverno/policy-reporter/pkg/kubernetes/wgpolicy"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/report/result"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type initialClient struct {
	client   report.PolicyReportClient
	queue    workqueue.TypedRateLimitingInterface[string]
	typed    *ktesting.Fake
	meta     *metafake.FakeMetadataClient
	tracker  ktesting.ObjectTracker
	resource schema.GroupVersionResource
}

func newInitialClient(t *testing.T, family string, publisher report.EventPublisher, strict, periodic, seed bool, filter *report.SourceFilter, metaFilters ...*report.MetaFilter) initialClient {
	t.Helper()
	scheme := metafake.NewTestScheme()
	metav1.AddMetaToScheme(scheme)
	meta := metafake.NewSimpleMetadataClient(scheme)
	work := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())
	h := initialClient{queue: work, meta: meta}
	metaFilter := report.NewMetaFilter(false, validate.RuleSets{})
	if len(metaFilters) > 0 {
		metaFilter = metaFilters[0]
	}
	m := metav1.ObjectMeta{Name: "report", Namespace: "test"}
	cm := metav1.ObjectMeta{Name: "cluster-report"}
	var clusterResource schema.GroupVersionResource
	var reports []runtime.Object
	switch family {
	case "wgpolicy":
		typed := prfake.NewSimpleClientset()
		h.typed = &typed.Fake
		h.tracker = typed.Tracker()
		h.resource = wgclient.PolrResource
		clusterResource = wgclient.CpolrResource
		reports = []runtime.Object{&pr.PolicyReport{ObjectMeta: m, Results: []pr.PolicyReportResult{{Source: "test", Policy: "example", Result: pr.StatusFail}}}, &pr.ClusterPolicyReport{ObjectMeta: cm, Results: []pr.PolicyReportResult{{Source: "test", Policy: "example", Result: pr.StatusFail}}}}
		q := wgclient.NewWGPolicyQueue(kubernetes.NewDebouncer(10*time.Millisecond, publisher), work, typed.Wgpolicyk8sV1alpha2(), filter, result.NewReconditioner(nil))
		h.client = wgclient.NewPolicyReportClient(meta, metaFilter, q, periodic, 30*time.Millisecond, nil, strict)
	default:
		typed := orfake.NewSimpleClientset()
		h.typed = &typed.Fake
		h.tracker = typed.Tracker()
		h.resource = orclient.OpenreportsReport
		clusterResource = orclient.OpenreportsCReport
		reports = []runtime.Object{&or.Report{ObjectMeta: m, Results: []or.ReportResult{{Source: "test", Policy: "example", Result: "fail"}}}, &or.ClusterReport{ObjectMeta: cm, Results: []or.ReportResult{{Source: "test", Policy: "example", Result: "fail"}}}}
		q := orclient.NewORQueue(kubernetes.NewDebouncer(10*time.Millisecond, publisher), work, typed.OpenreportsV1alpha1(), filter, result.NewReconditioner(nil))
		h.client = orclient.NewOpenreportsClient(meta, metaFilter, q, periodic, 30*time.Millisecond, nil, strict)
	}
	if seed {
		for _, obj := range reports {
			require.NoError(t, h.tracker.Add(obj))
		}
		_, err := meta.Resource(h.resource).Namespace("test").(metafake.MetadataClient).CreateFake(&metav1.PartialObjectMetadata{ObjectMeta: m}, metav1.CreateOptions{})
		require.NoError(t, err)
		_, err = meta.Resource(clusterResource).(metafake.MetadataClient).CreateFake(&metav1.PartialObjectMetadata{ObjectMeta: cm}, metav1.CreateOptions{})
		require.NoError(t, err)
	}
	t.Cleanup(work.ShutDown)
	return h
}

func startInitialClient(t *testing.T, h initialClient) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	stop := make(chan struct{})
	go func() { done <- h.client.Run(2, stop) }()
	return done
}

func TestInitialPersistenceReadiness(t *testing.T) {
	for _, family := range []string{"wgpolicy", "openreports"} {
		for _, strict := range []bool{false, true} {
			for _, restart := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/strict=%v/restart=%v", family, strict, restart), func(t *testing.T) {
					ctx := context.Background()
					db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
					require.NoError(t, err)
					defer db.Close()
					store, err := database.NewStore(db, "test")
					require.NoError(t, err)
					require.NoError(t, store.PrepareDatabase(ctx))
					entered := make(chan struct{})
					release := make(chan struct{})
					var enterOnce, releaseOnce sync.Once
					defer releaseOnce.Do(func() { close(release) })
					publisher := report.NewEventPublisher()
					var published atomic.Int32
					publisher.RegisterPostListener("completed", func(context.Context, report.LifecycleEvent) { published.Add(1) })
					persist := listener.NewStoreListener(store)
					publisher.RegisterListener(listener.Store, func(ctx context.Context, event report.LifecycleEvent) {
						enterOnce.Do(func() { close(entered) })
						<-release
						persist(ctx, event)
					})
					h := newInitialClient(t, family, publisher, strict, restart, true, report.NewSourceFilter(nil, nil, nil, nil, nil))
					done := startInitialClient(t, h)
					select {
					case <-entered:
					case <-time.After(5 * time.Second):
						t.Fatal("store listener not reached")
					}
					if restart {
						select {
						case err := <-done:
							require.NoError(t, err)
						case <-time.After(5 * time.Second):
							t.Fatal("periodic restart did not occur")
						}
						done = startInitialClient(t, h)
						select {
						case err := <-done:
							require.NoError(t, err)
						case <-time.After(5 * time.Second):
							t.Fatal("second periodic restart did not occur")
						}
						done = nil
					}
					defer h.client.Stop()
					check := func() error {
						if !h.client.HasProcessedInitialReports() {
							return errors.New("initial reports pending")
						}
						return nil
					}
					health := func() error {
						if !h.client.HasSynced() {
							return errors.New("not synced")
						}
						return nil
					}
					options := []api.ServerOption{api.WithHealthChecks([]api.HealthCheck{health}, check)}
					if strict {
						options = append(options, api.WithRESTReadiness(check))
					}
					options = append(options, v1.WithAPI(store, target.NewCollection(), nil), v2.WithAPI(store, nil, target.Targets{}))
					server := api.NewServer(gin.New(), options...)
					request := func(path string, expected int) []byte {
						t.Helper()
						w := httptest.NewRecorder()
						server.Serve(w, httptest.NewRequest("GET", path, nil))
						require.Equal(t, expected, w.Code, path)
						return w.Body.Bytes()
					}
					request("/healthz", 200)
					expected := 200
					if strict {
						expected = 503
					}
					request("/ready", expected)
					request("/v1/policy-reports", expected)
					request("/v1/namespaces", expected)
					request("/v2/sources", expected)
					releaseOnce.Do(func() { close(release) })
					require.Eventually(t, func() bool { return h.client.HasProcessedInitialReports() && h.queue.Len() == 0 }, 5*time.Second, time.Millisecond)
					// Relists also enqueue ordinary updates; wait for those writes before
					// asserting stable counts in the restart/default-behavior scenarios.
					expectedPublications := int32(2)
					if restart {
						expectedPublications = 4
					}
					require.Eventually(t, func() bool { return published.Load() >= expectedPublications }, 5*time.Second, time.Millisecond)
					// For the disabled gate, wait for the writes via the public REST endpoint.
					require.Eventually(t, func() bool {
						w := httptest.NewRecorder()
						server.Serve(w, httptest.NewRequest("GET", "/v1/policy-reports", nil))
						var response struct {
							Count int `json:"count"`
						}
						_ = json.Unmarshal(w.Body.Bytes(), &response)
						return response.Count == 1
					}, 5*time.Second, time.Millisecond)
					for _, path := range []string{"/v1/policy-reports", "/v1/cluster-policy-reports"} {
						var response struct {
							Count int `json:"count"`
						}
						require.NoError(t, json.Unmarshal(request(path, 200), &response))
						require.Equal(t, 1, response.Count, path)
					}
					require.JSONEq(t, `["test"]`, string(request("/v1/namespaces", 200)))
					request("/ready", 200)
					request("/v2/sources", 200)
					if restart {
						done = startInitialClient(t, h)
						require.True(t, h.client.HasProcessedInitialReports())
						select {
						case err := <-done:
							require.NoError(t, err)
						case <-time.After(5 * time.Second):
							t.Fatal("post-initialization restart did not finish")
						}
						done = nil
						require.True(t, h.client.HasProcessedInitialReports())
					}
					h.client.Stop()
					if done != nil {
						select {
						case err := <-done:
							require.NoError(t, err)
						case <-time.After(5 * time.Second):
							t.Fatal("client did not stop")
						}
					}
				})
			}
		}
	}
}

func TestInitialReportsEmptyFilteredDeletedAndFetchErrors(t *testing.T) {
	for _, family := range []string{"wgpolicy", "openreports"} {
		for _, scenario := range []string{"empty", "filtered", "deleted", "fetch error", "disable cluster", "excluded namespace"} {
			t.Run(family+"/"+scenario, func(t *testing.T) {
				publisher := report.NewEventPublisher()
				publisher.RegisterListener(listener.Store, listener.NewStoreListener(report.NewPolicyReportStore()))
				var rules []report.SourceValidation
				if scenario == "filtered" {
					rules = []report.SourceValidation{{Sources: validate.RuleSets{Exclude: []string{"test"}}}}
				}
				metaFilter := report.NewMetaFilter(scenario == "disable cluster", validate.RuleSets{})
				if scenario == "excluded namespace" {
					metaFilter = report.NewMetaFilter(false, validate.RuleSets{Exclude: []string{"test"}})
				}
				h := newInitialClient(t, family, publisher, true, false, scenario != "empty", report.NewSourceFilter(nil, nil, nil, nil, rules), metaFilter)
				if scenario == "deleted" {
					require.NoError(t, h.tracker.Delete(h.resource, "test", "report"))
				}
				if scenario == "fetch error" {
					h.typed.PrependReactor("get", h.resource.Resource, func(ktesting.Action) (bool, runtime.Object, error) { return true, nil, errors.New("API unavailable") })
				}
				done := startInitialClient(t, h)
				defer h.client.Stop()
				if scenario == "fetch error" {
					require.Eventually(t, func() bool { return len(h.typed.Actions()) >= 7 }, 5*time.Second, time.Millisecond)
					require.False(t, h.client.HasProcessedInitialReports())
				} else {
					require.Eventually(t, h.client.HasProcessedInitialReports, 5*time.Second, time.Millisecond)
				}
				h.client.Stop()
				require.NoError(t, <-done)
			})
		}
	}
}

func TestInitialAcknowledgementSurvivesDebounce(t *testing.T) {
	for _, replacement := range []report.Event{report.Updated, report.Deleted} {
		t.Run(replacement.String(), func(t *testing.T) {
			initial := report.NewInitialReports("reports")
			initial.Track("reports", "test/report")
			initial.Seal("reports")
			publisher := report.NewEventPublisher()
			publisher.RegisterListener(listener.Store, listener.NewStoreListener(report.NewPolicyReportStore()))
			debouncer := kubernetes.NewDebouncer(20*time.Millisecond, publisher)
			empty := &openreports.ReportAdapter{Report: &or.Report{ObjectMeta: metav1.ObjectMeta{Name: "report", Namespace: "test"}}}
			full := &openreports.ReportAdapter{Report: &or.Report{ObjectMeta: metav1.ObjectMeta{Name: "report", Namespace: "test"}, Results: []or.ReportResult{{Source: "test", Policy: "example", Result: "fail"}}}}
			debouncer.Add(report.LifecycleEvent{Type: report.Updated, PolicyReport: empty, Persisted: initial.PersistenceResult("test/report")})
			require.False(t, initial.Complete())
			debouncer.Add(report.LifecycleEvent{Type: replacement, PolicyReport: full, Persisted: initial.PersistenceResult("test/report")})
			require.True(t, initial.Complete())
		})
	}
}

// Failed initial writes remain pending even if a relist no longer contains the
// report. The retained key must be fetched and resolved through deletion.
func TestInitialPendingDeletionAfterRestart(t *testing.T) {
	for _, family := range []string{"wgpolicy", "openreports"} {
		t.Run(family, func(t *testing.T) {
			store := &initialUnavailableStore{PolicyReportStore: report.NewPolicyReportStore()}
			publisher := report.NewEventPublisher()
			publisher.RegisterListener(listener.Store, listener.NewStoreListener(store))
			h := newInitialClient(t, family, publisher, true, false, true, report.NewSourceFilter(nil, nil, nil, nil, nil))
			done := startInitialClient(t, h)
			require.Eventually(t, func() bool { return store.attempts.Load() == 5 }, 5*time.Second, time.Millisecond)
			require.False(t, h.client.HasProcessedInitialReports())
			h.client.Stop()
			require.NoError(t, <-done)
			require.NoError(t, h.tracker.Delete(h.resource, "test", "report"))
			h.meta.PrependReactor("list", h.resource.Resource, func(ktesting.Action) (bool, runtime.Object, error) {
				return true, &metav1.List{}, nil
			})
			done = startInitialClient(t, h)
			defer h.client.Stop()

			require.Eventually(t, h.client.HasProcessedInitialReports, 5*time.Second, time.Millisecond)
			h.client.Stop()
			// Cancellation may interrupt the replacement informer's cache-sync wait
			// after the retained workers have already resolved the pending deletion.
			if err := <-done; err != nil {
				require.Contains(t, err.Error(), "failed to sync")
			}
		})
	}
}

type initialUnavailableStore struct {
	report.PolicyReportStore
	attempts atomic.Int32
}

func (s *initialUnavailableStore) Update(ctx context.Context, rep openreports.ReportInterface) error {
	if rep.GetNamespace() != "" {
		s.attempts.Add(1)
		return errors.New("database unavailable")
	}
	return s.PolicyReportStore.Update(ctx, rep)
}
