package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestUpdateKeepsPreviousReportWhenInsertFails(t *testing.T) {
	ctx := context.Background()
	db, err := NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	store, err := NewStore(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PrepareDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	rep := result.NewReconditioner(nil).Prepare(fixtures.DefaultPolicyReport)
	if err := store.Add(ctx, rep); err != nil {
		t.Fatal(err)
	}

	count := func(table string) (n int) {
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	reports, results := count("policy_report"), count("policy_report_result")
	if reports == 0 || results == 0 {
		t.Fatalf("setup stored %d reports and %d results", reports, results)
	}

	// Fail late, after the new parent row has been inserted.
	if _, err := db.Exec(`CREATE TRIGGER reject_result BEFORE INSERT ON policy_report_result BEGIN SELECT RAISE(FAIL, 'result insert rejected'); END`); err != nil {
		t.Fatal(err)
	}
	if err := store.Update(ctx, rep); err == nil {
		t.Fatal("expected Update to fail")
	}

	if got := count("policy_report"); got != reports {
		t.Errorf("policy_report rows = %d, want %d", got, reports)
	}
	if got := count("policy_report_result"); got != results {
		t.Errorf("policy_report_result rows = %d, want %d", got, results)
	}
}
