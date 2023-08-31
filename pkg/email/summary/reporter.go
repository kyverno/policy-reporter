package summary

import (
	"html/template"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/email"
)

type Reporter struct {
	templateDir string
	clusterName string
	titlePrefix string
}

func (o *Reporter) Report(sources []Source, format string) (email.Report, error) {
	b := new(strings.Builder)

	templ, err := template.ParseFiles(o.templateDir + "/summary.html")
	if err != nil {
		return email.Report{}, err
	}

	err = templ.Execute(b, struct {
		Sources     []Source
		ClusterName string
		TitlePrefix string
	}{
		Sources:     sources,
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
		Title:       o.titlePrefix + " (summary)" + titleCluster + "from " + time.Now().Format("2006-01-02"),
		Message:     b.String(),
		Format:      format,
	}, nil
}

func NewReporter(templateDir, clusterName string, titlePrefix string) *Reporter {
	return &Reporter{templateDir, clusterName, titlePrefix}
}
