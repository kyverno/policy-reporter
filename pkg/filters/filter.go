package filters

type ValueFilter struct {
	// +optional
	Include []string `mapstructure:"include" json:"include"`
	// +optional
	Exclude []string `mapstructure:"exclude" json:"exclude"`
	// +optional
	Selector map[string]string `mapstructure:"selector" json:"selector"`
}

type Filter struct {
	// +optional
	Namespaces ValueFilter `mapstructure:"namespaces" json:"namespaces"`
	// +optional
	Status ValueFilter `mapstructure:"status" json:"status"`
	// +optional
	Severities ValueFilter `mapstructure:"severities" json:"severities"`
	// +optional
	Policies ValueFilter `mapstructure:"policies" json:"policies"`
	// +optional
	Sources ValueFilter `mapstructure:"sources" json:"sources"`
	// +optional
	ReportLabels ValueFilter `mapstructure:"reportLabels" json:"reportLabels"`
}
