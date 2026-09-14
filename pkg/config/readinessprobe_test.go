package config_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func Test_ReadinessProbe(t *testing.T) {
	t.Parallel()

	t.Run("returns immediately without leader election", func(t *testing.T) {
		t.Parallel()

		for _, restEnabled := range []bool{false, true} {
			rdy := config.NewReadinessProbe(&config.Config{
				REST:           config.REST{Enabled: restEnabled},
				LeaderElection: config.LeaderElection{Enabled: false},
			})

			waitDone := make(chan struct{})
			go func() {
				rdy.Wait()
				close(waitDone)
			}()

			require.Eventually(t, func() bool {
				select {
				case <-waitDone:
					return true
				default:
					return false
				}
			}, time.Second, time.Millisecond)
			assert.False(t, rdy.Running())
		}
	})

	for _, tc := range []struct {
		name        string
		restEnabled bool
	}{
		{name: "with REST disabled", restEnabled: false},
		{name: "with REST enabled", restEnabled: true},
	} {
		t.Run("waits for leader election "+tc.name, func(t *testing.T) {
			t.Parallel()

			rdy := config.NewReadinessProbe(&config.Config{
				REST:           config.REST{Enabled: tc.restEnabled},
				LeaderElection: config.LeaderElection{Enabled: true},
			})

			const waiterCount = 4
			waitStarted := make(chan struct{}, waiterCount)
			waitDone := make(chan struct{}, waiterCount)
			for range waiterCount {
				go func() {
					waitStarted <- struct{}{}
					rdy.Wait()
					waitDone <- struct{}{}
				}()
			}

			for range waiterCount {
				<-waitStarted
			}
			runtime.Gosched()

			assert.Never(t, func() bool {
				return len(waitDone) > 0
			}, 50*time.Millisecond, time.Millisecond)
			assert.False(t, rdy.Running())

			rdy.Ready()

			require.Eventually(t, func() bool {
				return len(waitDone) == waiterCount
			}, time.Second, time.Millisecond)
			assert.True(t, rdy.Running())
		})
	}

	t.Run("allows concurrent Ready calls", func(t *testing.T) {
		t.Parallel()

		rdy := config.NewReadinessProbe(&config.Config{
			LeaderElection: config.LeaderElection{Enabled: true},
		})

		const callerCount = 20
		start := make(chan struct{})
		var callers sync.WaitGroup
		callers.Add(callerCount)
		for range callerCount {
			go func() {
				defer callers.Done()
				<-start
				rdy.Ready()
			}()
		}

		close(start)
		callers.Wait()
		rdy.Ready()
		rdy.Wait()

		assert.True(t, rdy.Running())
	})
}
