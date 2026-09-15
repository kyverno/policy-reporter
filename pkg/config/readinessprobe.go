package config

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type ReadinessProbe struct {
	config *Config

	ready     chan struct{}
	readyOnce sync.Once
	running   atomic.Bool
}

func (r *ReadinessProbe) required() bool {
	return r.config.LeaderElection.Enabled
}

func (r *ReadinessProbe) Ready() {
	if !r.required() {
		return
	}

	r.readyOnce.Do(func() {
		zap.L().Debug("readiness probe ready")
		close(r.ready)
	})
}

func (r *ReadinessProbe) Wait() {
	if !r.required() {
		return
	}

	zap.L().Debug("readiness probe waiting")
	<-r.ready
	r.running.Store(true)
	zap.L().Debug("readiness probe finished")
}

func (r *ReadinessProbe) Running() bool {
	return r.running.Load()
}

func NewReadinessProbe(config *Config) *ReadinessProbe {
	return &ReadinessProbe{
		config: config,
		ready:  make(chan struct{}),
	}
}
