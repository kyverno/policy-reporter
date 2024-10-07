package validate

type RuleSets struct {
	Exclude  []string
	Include  []string
	Selector map[string]string
}

func (r RuleSets) Count() int {
	return len(r.Exclude) + len(r.Include)
}

func (r RuleSets) Enabled() bool {
	return r.Count() > 0
}
