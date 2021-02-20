package report

import "fmt"

type Status = string
type Severity = string
type Priority = string

const (
	Fail  Status = "fail"
	Warn  Status = "warn"
	Error Status = "error"
	Pass  Status = "pass"
	Skip  Status = "skip"

	Low    Severity = "low"
	Medium Severity = "medium"
	Heigh  Severity = "heigh"

	Alert       Priority = "error"
	Warning     Priority = "warning"
	Information Priority = "info"
	Debug       Priority = "debug"
)

type Resource struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	UID        string
}

type Result struct {
	Message   string
	Policy    string
	Rule      string
	Priority  Priority
	Status    Status
	Severity  Severity
	Category  string
	Scored    bool
	Resources []Resource
}

func (r Result) GetIdentifier() string {
	res := Resource{}
	if len(r.Resources) > 0 {
		res = r.Resources[0]
	}

	return fmt.Sprintf("%s__%s__%s__%s", r.Policy, r.Rule, r.Status, res.UID)
}

type Report interface {
	GetIdentifier() string
}

type Summary struct {
	Pass  int
	Skip  int
	Warn  int
	Error int
	Fail  int
}

type PolicyReport struct {
	Name      string
	Namespace string
	Results   map[string]Result
	Summary   Summary
}

func (pr PolicyReport) GetIdentifier() string {
	return fmt.Sprintf("%s__%s", pr.Namespace, pr.Name)
}

func (pr PolicyReport) GetNewValidation(or PolicyReport) []Result {
	diff := make([]Result, 0)

	for _, r := range pr.Results {
		if _, ok := or.Results[r.GetIdentifier()]; ok {
			continue
		}

		diff = append(diff, r)
	}

	return diff
}

type ClusterPolicyReport struct {
	Name    string
	Results map[string]Result
	Summary Summary
}

func (cr ClusterPolicyReport) GetIdentifier() string {
	return cr.Name
}

func (cr ClusterPolicyReport) GetNewValidation(cor ClusterPolicyReport) []Result {
	diff := make([]Result, 0)

	for _, r := range cr.Results {
		if _, ok := cor.Results[r.GetIdentifier()]; ok {
			continue
		}

		diff = append(diff, r)
	}

	return diff
}
