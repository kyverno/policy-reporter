package report_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func TestInitialReports(t *testing.T) {
	initial := report.NewInitialReports("namespaced", "cluster")
	initial.Track("namespaced", "test/report")
	done := initial.PersistenceResult("test/report")
	if done == nil {
		t.Fatal("missing acknowledgement")
	}
	done(nil)
	if initial.Complete() {
		t.Fatal("completed before handlers synced")
	}
	initial.Seal("namespaced")
	initial.Track("cluster", "report")
	initial.Seal("cluster")
	initial.PersistenceResult("report")(errors.New("database unavailable"))
	if initial.Complete() {
		t.Fatal("completed after failed write")
	}
	// A relist neither resets pending work nor expands a sealed startup set.
	initial.Track("namespaced", "new/report")
	initial.Track("cluster", "report")
	if len(initial.Pending()) != 1 {
		t.Fatal(initial.Pending())
	}
	ack := initial.PersistenceResult("report")
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() { defer wg.Done(); ack(nil) }()
	}
	wg.Wait()
	if !initial.Complete() {
		t.Fatal("successful persistence did not complete initialization")
	}
	initial.Track("cluster", "later")
	if !initial.Complete() {
		t.Fatal("completion was not latched")
	}
}

func TestInitialReportsEmptyAndFiltered(t *testing.T) {
	initial := report.NewInitialReports("namespaced")
	initial.Track("namespaced", "excluded")
	initial.Resolve("excluded")
	if initial.Complete() {
		t.Fatal("must wait for handler synchronization even with no pending reports")
	}
	initial.Seal("namespaced")
	if !initial.Complete() {
		t.Fatal("empty set should complete")
	}
	var disabled *report.InitialReports
	if !disabled.Complete() {
		t.Fatal("disabled tracker must not gate readiness")
	}
}
