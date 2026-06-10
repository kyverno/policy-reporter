package config

import (
	"go.uber.org/zap"
)

type ReadinessProbe struct {
	config *Config

	ready   chan struct{}
	running bool
}

func (r *ReadinessProbe) required() bool {
	if !r.config.REST.Enabled {
		return false
	}

	return r.config.LeaderElection.Enabled
}

func (r *ReadinessProbe) Ready() {
	if r.required() && !r.running {
		go func() {
			zap.L().Debug("readiness probe ready")
			close(r.ready)
		}()
	}
}

func (r *ReadinessProbe) Wait() {
	if r.required() && !r.running {
		zap.L().Debug("readiness probe waiting")
		<-r.ready
		r.running = true
		zap.L().Debug("readiness probe finished")
		return
	}
}

func (r *ReadinessProbe) Running() bool {
	return r.running
}

func NewReadinessProbe(config *Config) *ReadinessProbe {
	return &ReadinessProbe{
		config:  config,
		ready:   make(chan struct{}),
		running: false,
	}
}
