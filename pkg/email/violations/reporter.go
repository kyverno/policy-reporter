package violations

import (
	"html/template"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type Reporter struct {
	templateDir string
	clusterName string
	titlePrefix string
}

func (o *Reporter) Report(sources []Source, format string) (email.Report, error) {
	b := new(strings.Builder)

	vioTempl := template.New("violations.html").Funcs(template.FuncMap{
		"color": email.ColorFromStatus,
		"title": helper.Title,
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
		TitlePrefix string
	}{
		Sources:     sources,
		Status:      []string{"warn", "fail", "error"},
		ClusterName: o.clusterName,
		TitlePrefix: o.titlePrefix,
	})
	if err != nil {
		return email.Report{}, err
	}

	titleCluster := " "
	if o.clusterName != "" {
		titleCluster = " on " + o.clusterName + " "
	}

	return email.Report{
		ClusterName: o.clusterName,
		Title:       o.titlePrefix + " (violations)" + titleCluster + "from " + time.Now().Format("2006-01-02"),
		Message:     b.String(),
		Format:      format,
	}, nil
}

func NewReporter(templateDir string, clusterName string, titlePrefix string) *Reporter {
	return &Reporter{templateDir, clusterName, titlePrefix}
}
