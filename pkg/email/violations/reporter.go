package violations

import (
	"html/template"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/kyverno/policy-reporter/pkg/email"
)

type Reporter struct {
	templateDir string
	clusterName string
}

func (o *Reporter) Report(sources []Source, format string) (email.Report, error) {
	b := new(strings.Builder)

	vioTempl := template.New("violations.html").Funcs(template.FuncMap{
		"color": email.ColorFromStatus,
		"title": cases.Title(language.English, cases.NoLower),
		"hasViolations": func(results map[string][]Result) bool {
			return (len(results["warn"]) + len(results["fail"]) + len(results["error"])) > 0
		},
		"lenNamespaceResults": func(source Source, ns, status string) int {
			return len(source.NamespaceResults[ns][status])
		},
	})

	templ, err := vioTempl.ParseFiles(o.templateDir + "/violations.html")
	if err != nil {
		return email.Report{}, err
	}

	err = templ.Execute(b, struct {
		Sources     []Source
		Status      []string
		ClusterName string
	}{Sources: sources, Status: []string{"warn", "fail", "error"}, ClusterName: o.clusterName})
	if err != nil {
		return email.Report{}, err
	}

	return email.Report{
		ClusterName: o.clusterName,
		Title:       "Summary Report from " + time.Now().Format("2006-01-02"),
		Message:     b.String(),
		Format:      format,
	}, nil
}

func NewReporter(templateDir string, clusterName string) *Reporter {
	return &Reporter{templateDir, clusterName}
}
