package targetconfig

import "github.com/kyverno/policy-reporter/pkg/filters"

type Config[T any] struct {
	Config          *T                `mapstructure:"config" json:"config"`
	Name            string            `mapstructure:"name" json:"name"`
	MinimumSeverity string            `mapstructure:"minimumSeverity" json:"minimumSeverity"`
	Filter          filters.Filter    `mapstructure:"filter" json:"filter"`
	SecretRef       string            `mapstructure:"secretRef" json:"secretRef"`
	MountedSecret   string            `mapstructure:"mountedSecret" json:"mountedSecret"`
	Sources         []string          `mapstructure:"sources" json:"sources"`
	CustomFields    map[string]string `mapstructure:"customFields" json:"customFields"`
	SkipExisting    bool              `mapstructure:"skipExistingOnStartup" json:"skipExistingOnStartup"`
	Channels        []*Config[T]      `mapstructure:"channels" json:"channels"`
	Valid           bool              `mapstructure:"-" json:"-"`
}

func (config *Config[T]) MapBaseParent(parent *Config[T]) {
	if config.MinimumSeverity == "" {
		config.MinimumSeverity = parent.MinimumSeverity
	}

	if !config.SkipExisting {
		config.SkipExisting = parent.SkipExisting
	}
}

func (config *Config[T]) Secret() string {
	return config.SecretRef
}
