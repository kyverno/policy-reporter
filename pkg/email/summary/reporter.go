package summary

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/templates"
)

type Reporter struct {
	templateDir string
	clusterName string
	titlePrefix string
}

func (o *Reporter) Report(sources []Source, format string) (email.Report, error) {
	b := new(strings.Builder)

	templ, err := parseTemplate(o.templateDir, "summary.html")
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

// parseTemplate loads name preferring the on-disk templateDir when the file
// exists there and otherwise falling back to the embedded copy that ships with
// the binary, keeping the --template-dir override working while making
// rendering resilient to a missing templates directory.
func parseTemplate(templateDir, name string) (*template.Template, error) {
	if templateDir != "" {
		path := filepath.Join(templateDir, name)
		if _, err := os.Stat(path); err == nil {
			return template.ParseFiles(path)
		}
	}

	return template.ParseFS(templates.FS, name)
}

func NewReporter(templateDir, clusterName string, titlePrefix string) *Reporter {
	return &Reporter{templateDir, clusterName, titlePrefix}
}
