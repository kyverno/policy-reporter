package config_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/config"
)

func Test_ReadinessProbe(t *testing.T) {
	t.Run("immediate return without REST enabled", func(t *testing.T) {
		rdy := config.NewReadinessProbe(
			&config.Config{
				REST:           config.REST{Enabled: false},
				LeaderElection: config.LeaderElection{Enabled: false},
			},
		)

		rdy.Wait()
	})

	t.Run("immediate return without LeaderElection enabled", func(t *testing.T) {
		rdy := config.NewReadinessProbe(
			&config.Config{
				REST:           config.REST{Enabled: true},
				LeaderElection: config.LeaderElection{Enabled: false},
			},
		)

		rdy.Wait()
	})

	t.Run("wait for ready state", func(t *testing.T) {
		rdy := config.NewReadinessProbe(
			&config.Config{
				REST:           config.REST{Enabled: true},
				LeaderElection: config.LeaderElection{Enabled: true},
			},
		)

		if rdy.Running() {
			t.Error("should not be running until ready was called")
		}

		go func() {
			rdy.Wait()
			if !rdy.Running() {
				t.Error("should be running after ready was called")
			}
		}()

		rdy.Ready()
	})
}
