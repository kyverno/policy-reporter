package kubernetes

import (
	"context"
	"sync"
	"testing"
	"time"

	or "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_Debouncer(t *testing.T) {
	t.Parallel()
	t.Run("Skip Empty Update", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := NewDebouncer(200*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to skip the empty modify event")
		}
	})

	t.Run("Execute Empty Updates after waiting time", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := NewDebouncer(10*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(5 * time.Millisecond)

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to execute empty update event after wait time")
		}
	})

	t.Run("Skip Empty Update when delete event follows directly", func(t *testing.T) {
		t.Parallel()
		counter := 0
		wg := sync.WaitGroup{}
		wg.Add(2)

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			counter++
			wg.Done()
		})

		debouncer := NewDebouncer(200*time.Millisecond, publisher)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Added,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Updated,
			PolicyReport: &openreports.ReportAdapter{Report: fixtures.MinPolicyReport.Report},
		})

		time.Sleep(10 * time.Millisecond)

		debouncer.Add(report.LifecycleEvent{
			Type:         report.Deleted,
			PolicyReport: fixtures.DefaultPolicyReport,
		})

		wg.Wait()

		if counter != 2 {
			t.Error("Expected to skip the empty modify event")
		}
	})
}

type debounceRecorder struct {
	mutx   sync.Mutex
	events []report.LifecycleEvent
}

func (r *debounceRecorder) record(_ context.Context, event report.LifecycleEvent) {
	r.mutx.Lock()
	r.events = append(r.events, event)
	r.mutx.Unlock()
}

func (r *debounceRecorder) snapshot() []report.LifecycleEvent {
	r.mutx.Lock()
	defer r.mutx.Unlock()
	return append([]report.LifecycleEvent(nil), r.events...)
}

func newDebouncePublisher(recorder *debounceRecorder) report.EventPublisher {
	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", recorder.record)
	return publisher
}

func debounceReport(eventType report.Event, name string, results int) report.LifecycleEvent {
	rep := &or.Report{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test"}}
	for i := 0; i < results; i++ {
		rep.Results = append(rep.Results, or.ReportResult{Source: "test", Policy: "example", Result: "fail"})
	}
	return report.LifecycleEvent{Type: eventType, PolicyReport: &openreports.ReportAdapter{Report: rep}}
}

func debounceEmptyUpdate(marker string) report.LifecycleEvent {
	return report.LifecycleEvent{Type: report.Updated, PolicyReport: &openreports.ReportAdapter{Report: &or.Report{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-report", Namespace: "test", Labels: map[string]string{"marker": marker}},
	}}}
}

// pendingWindow returns the open window of a report without racing the timers.
func pendingWindow(d *debouncer, id string) *window {
	d.mutx.Lock()
	defer d.mutx.Unlock()
	return d.pending[id]
}

// setPendingWindow installs a window directly, so tests can expire windows that
// no live timer owns.
func setPendingWindow(d *debouncer, id string, w *window) {
	d.mutx.Lock()
	d.pending[id] = w
	d.mutx.Unlock()
}

func stoppedWindow(t *testing.T, event report.LifecycleEvent) *window {
	t.Helper()
	w := &window{event: event, timer: time.NewTimer(time.Hour)}
	t.Cleanup(func() { w.timer.Stop() })
	return w
}

func Test_DebouncerOrdering(t *testing.T) {
	t.Run("Coalesce empty updates into the original window", func(t *testing.T) {
		t.Parallel()
		recorder := &debounceRecorder{}
		debouncer := NewDebouncer(time.Hour, newDebouncePublisher(recorder)).(*debouncer)
		first := debounceEmptyUpdate("first")
		id := first.PolicyReport.GetID()

		debouncer.Add(first)
		open := pendingWindow(debouncer, id)
		require.NotNil(t, open)

		debouncer.Add(debounceEmptyUpdate("latest"))
		coalesced := pendingWindow(debouncer, id)
		require.Same(t, open, coalesced, "second empty update must reuse the open window instead of restarting it")
		require.Equal(t, "latest", coalesced.event.PolicyReport.GetLabels()["marker"], "window must retain the latest empty update")
		require.Empty(t, recorder.snapshot(), "empty updates must not publish before the window expires")

		debouncer.expire(id, coalesced)
		events := recorder.snapshot()
		require.Len(t, events, 1, "coalesced window must publish exactly once")
		require.Equal(t, "latest", events[0].PolicyReport.GetLabels()["marker"])
		require.Nil(t, pendingWindow(debouncer, id))

		debouncer.expire(id, coalesced)
		require.Len(t, recorder.snapshot(), 1, "expired window must not publish again")
	})

	for _, replacement := range []report.LifecycleEvent{
		debounceReport(report.Updated, "policy-report", 1),
		debounceReport(report.Added, "policy-report", 1),
		debounceReport(report.Deleted, "policy-report", 0),
	} {
		t.Run(replacement.Type.String()+" supersedes the pending window", func(t *testing.T) {
			t.Parallel()
			recorder := &debounceRecorder{}
			debouncer := NewDebouncer(time.Hour, newDebouncePublisher(recorder)).(*debouncer)
			pending := debounceEmptyUpdate("pending")
			id := pending.PolicyReport.GetID()

			debouncer.Add(pending)
			canceled := pendingWindow(debouncer, id)
			require.NotNil(t, canceled)

			debouncer.Add(replacement)

			events := recorder.snapshot()
			require.Len(t, events, 1, "superseding event must publish immediately")
			require.True(t, events[0].PolicyReport == replacement.PolicyReport, "superseding event must be published, not the pending update")
			require.Nil(t, pendingWindow(debouncer, id), "superseding event must cancel the pending window")

			debouncer.expire(id, canceled)
			require.Len(t, recorder.snapshot(), 1, "canceled window must not publish when its timer still fires")
		})
	}

	t.Run("Stale window cannot publish or delete a newer window", func(t *testing.T) {
		t.Parallel()
		recorder := &debounceRecorder{}
		debouncer := NewDebouncer(time.Hour, newDebouncePublisher(recorder)).(*debouncer)
		current := debounceEmptyUpdate("current")
		id := current.PolicyReport.GetID()

		stale := stoppedWindow(t, debounceEmptyUpdate("stale"))
		newer := stoppedWindow(t, current)
		setPendingWindow(debouncer, id, newer)

		// A timer of a replaced window fires after the newer window took over.
		debouncer.expire(id, stale)
		require.Empty(t, recorder.snapshot(), "stale timer must not publish")
		require.Same(t, newer, pendingWindow(debouncer, id), "stale timer must not delete the newer window")

		debouncer.expire(id, newer)
		events := recorder.snapshot()
		require.Len(t, events, 1, "newer window must still publish")
		require.Equal(t, "current", events[0].PolicyReport.GetLabels()["marker"])
	})

	t.Run("Recreated report publishes its own window", func(t *testing.T) {
		t.Parallel()
		recorder := &debounceRecorder{}
		debouncer := NewDebouncer(10*time.Millisecond, newDebouncePublisher(recorder))

		deleted := debounceReport(report.Deleted, "policy-report", 0)
		added := debounceReport(report.Added, "policy-report", 1)
		debouncer.Add(debounceEmptyUpdate("pending"))
		debouncer.Add(deleted)
		debouncer.Add(added)
		require.Len(t, recorder.snapshot(), 2, "delete and recreate publish immediately")

		debouncer.Add(debounceEmptyUpdate("recreated"))
		require.Eventually(t, func() bool { return len(recorder.snapshot()) == 3 }, 5*time.Second, time.Millisecond)

		events := recorder.snapshot()
		require.True(t, events[0].PolicyReport == deleted.PolicyReport, "delete must be published first")
		require.True(t, events[1].PolicyReport == added.PolicyReport, "recreate must follow the delete")
		require.Equal(t, "recreated", events[2].PolicyReport.GetLabels()["marker"], "recreated window must publish its own event")
	})

	t.Run("Coalesced window keeps the persistence acknowledgement", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name        string
			acknowledge bool
		}{
			{name: "both events acknowledge", acknowledge: true},
			{name: "latest has no callback", acknowledge: false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				initial := report.NewInitialReports("reports")
				initial.Track("reports", "test/report")
				initial.Seal("reports")
				recorder := &debounceRecorder{}
				publisher := report.NewEventPublisher()
				publisher.RegisterListener("test", func(ctx context.Context, event report.LifecycleEvent) {
					recorder.record(ctx, event)
					if event.Persisted != nil {
						event.Persisted(nil)
					}
				})
				debouncer := NewDebouncer(time.Hour, publisher).(*debouncer)

				first := report.LifecycleEvent{Type: report.Updated, PolicyReport: debounceEmptyUpdate("first").PolicyReport, Persisted: initial.PersistenceResult("test/report")}
				require.NotNil(t, first.Persisted)
				latest := report.LifecycleEvent{Type: report.Updated, PolicyReport: debounceEmptyUpdate("latest").PolicyReport}
				if tc.acknowledge {
					latest.Persisted = initial.PersistenceResult("test/report")
					require.NotNil(t, latest.Persisted)
				}

				debouncer.Add(first)
				debouncer.Add(latest)
				require.False(t, initial.Complete(), "coalesced window must not acknowledge before it expires")

				debouncer.expire(first.PolicyReport.GetID(), pendingWindow(debouncer, first.PolicyReport.GetID()))
				events := recorder.snapshot()
				require.Len(t, events, 1)
				require.NotNil(t, events[0].Persisted, "retained event must keep a persistence callback")
				require.True(t, initial.Complete(), "retained event must acknowledge the initial report")
			})
		}
	})

	t.Run("Serialize publications per report", func(t *testing.T) {
		t.Parallel()
		recorder := &debounceRecorder{}
		entered := make(chan struct{})
		release := make(chan struct{})
		var first sync.Once

		publisher := report.NewEventPublisher()
		publisher.RegisterListener("test", func(_ context.Context, event report.LifecycleEvent) {
			blocked := false
			first.Do(func() { blocked = true })
			if blocked {
				close(entered)
				<-release
			}
			recorder.record(context.Background(), event)
		})
		debouncer := NewDebouncer(10*time.Millisecond, publisher)

		update1 := debounceReport(report.Updated, "policy-report", 1)
		update2 := debounceReport(report.Updated, "policy-report", 1)
		other := debounceReport(report.Added, "other-report", 1)

		published := make(chan struct{})
		go func() {
			debouncer.Add(update1)
			close(published)
		}()
		<-entered

		added := make(chan struct{})
		go func() {
			debouncer.Add(update2)
			close(added)
		}()
		select {
		case <-added:
		case <-time.After(5 * time.Second):
			t.Fatal("Add blocked behind an in-flight publication of the same report")
		}

		otherAdded := make(chan struct{})
		go func() {
			debouncer.Add(other)
			close(otherAdded)
		}()
		select {
		case <-otherAdded:
		case <-time.After(5 * time.Second):
			t.Fatal("unrelated report blocked behind an in-flight listener")
		}

		require.Len(t, recorder.snapshot(), 1, "blocked listener must stop further publications of its report")
		require.True(t, recorder.snapshot()[0].PolicyReport == other.PolicyReport)

		close(release)
		select {
		case <-published:
		case <-time.After(5 * time.Second):
			t.Fatal("first publication did not finish")
		}

		events := recorder.snapshot()
		require.Len(t, events, 3)
		require.True(t, events[1].PolicyReport == update1.PolicyReport, "in-flight publication must finish before the queued one")
		require.True(t, events[2].PolicyReport == update2.PolicyReport, "queued publication must keep its order")
	})
}
