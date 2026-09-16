package violations

import (
	"bytes"
	"encoding/csv"
	"slices"
	"sort"

	"github.com/kyverno/policy-reporter/pkg/email"
)

// EmailReport selects attachment mode per delivery, leaving Report and the HTML
// API independent of email configuration.
func (o *Reporter) EmailReport(sources []Source, format, attachmentFormat string) (email.Report, error) {
	if err := email.ValidateAttachmentFormat(attachmentFormat); err != nil {
		return email.Report{}, err
	}
	if attachmentFormat == "" {
		return o.Report(sources, format)
	}
	data, err := o.csv(sources)
	if err != nil {
		return email.Report{}, err
	}
	report := o.newReport("The violation report data is attached as violations.csv.", format)
	report.Attachments = []email.Attachment{{Filename: "violations.csv", ContentType: "text/csv; charset=utf-8", Data: data}}
	return report, nil
}

func (o *Reporter) csv(sources []Source) ([]byte, error) {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	if err := w.Write([]string{"Cluster", "Source", "Scope", "Namespace", "Policy", "Rule", "Kind", "Name", "Status", "Severity", "Message"}); err != nil {
		return nil, err
	}
	sources = slices.Clone(sources)
	sort.Slice(sources, func(i, j int) bool { return sources[i].Name < sources[j].Name })
	for _, source := range sources {
		write := func(scope, ns string, results map[string][]Result) error {
			// Match the statuses and their order in the HTML template.
			for _, status := range []string{"warn", "fail", "error"} {
				sorted := slices.Clone(results[status])
				fields := func(r Result) []string {
					return []string{r.Policy, r.Rule, r.Kind, r.Name, r.Status, r.Severity, r.Message}
				}
				sort.Slice(sorted, func(i, j int) bool { return slices.Compare(fields(sorted[i]), fields(sorted[j])) < 0 })
				for _, result := range sorted {
					row := append([]string{o.clusterName, source.Name, scope, ns}, fields(result)...)
					for i, cell := range row {
						row[i] = email.CSVText(cell)
					}
					if err := w.Write(row); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if source.ClusterReports {
			if err := write("cluster", "", source.ClusterResults); err != nil {
				return nil, err
			}
		}
		namespaces := make([]string, 0, len(source.NamespaceResults))
		for ns := range source.NamespaceResults {
			namespaces = append(namespaces, ns)
		}
		sort.Strings(namespaces)
		for _, ns := range namespaces {
			if err := write("namespace", ns, source.NamespaceResults[ns]); err != nil {
				return nil, err
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
