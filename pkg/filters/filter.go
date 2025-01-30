package filters

type ValueFilter struct {
	Include  []string          `mapstructure:"include" json:"include"`
	Exclude  []string          `mapstructure:"exclude" json:"exclude"`
	Selector map[string]string `mapstructure:"selector" json:"selector"`
}

type Filter struct {
	Namespaces   ValueFilter `mapstructure:"namespaces" json:"namespaces"`
	Status       ValueFilter `mapstructure:"status" json:"status"`
	Severities   ValueFilter `mapstructure:"severities" json:"severities"`
	Policies     ValueFilter `mapstructure:"policies" json:"policies"`
	Sources      ValueFilter `mapstructure:"sources" json:"sources"`
	ReportLabels ValueFilter `mapstructure:"reportLabels" json:"reportLabels"`
}
