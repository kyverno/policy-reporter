package listener_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

type failingInitialStore struct {
	report.PolicyReportStore
	attempts int
	failures int
}

func (s *failingInitialStore) Update(context.Context, openreports.ReportInterface) error {
	s.attempts++
	if s.attempts <= s.failures {
		return errors.New("database unavailable")
	}
	return nil
}

func (s *failingInitialStore) Remove(context.Context, string) error {
	return s.Update(context.Background(), nil)
}

func TestInitialPersistenceAcknowledgement(t *testing.T) {
	for _, event := range []report.Event{report.Added, report.Updated, report.Deleted} {
		for _, failures := range []int{0, 2, 10} {
			t.Run(fmt.Sprintf("%s/failures=%d", event, failures), func(t *testing.T) {
				store := &failingInitialStore{failures: failures}
				initial := report.NewInitialReports("reports")
				initial.Track("reports", "test/report")
				initial.Seal("reports")
				ack := initial.PersistenceResult("test/report")
				acknowledgements := 0
				listener.NewStoreListener(store)(context.Background(), report.LifecycleEvent{Type: event, PolicyReport: fixtures.DefaultPolicyReport, Persisted: func(err error) { acknowledgements++; ack(err) }})
				require.Equal(t, 1, acknowledgements)
				if failures >= 5 {
					require.Equal(t, 5, store.attempts)
					require.False(t, initial.Complete())
				} else {
					require.Equal(t, failures+1, store.attempts)
					require.True(t, initial.Complete())
				}
			})
		}
	}
}
