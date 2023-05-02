package config

import (
	"go.uber.org/zap"
)

type ReadinessProbe struct {
	config *Config

	ready   chan bool
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
			r.ready <- true
		}()
	}
}

func (r *ReadinessProbe) Wait() {
	if r.required() && !r.running {
		r.running = <-r.ready
		close(r.ready)
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
		ready:   make(chan bool),
		running: false,
	}
}
