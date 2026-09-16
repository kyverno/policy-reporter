package summary

import (
	"bytes"
	"encoding/csv"
	"slices"
	"sort"
	"strconv"

	"github.com/kyverno/policy-reporter/pkg/email"
)

// EmailReport selects the delivery representation per call. Report remains the
// HTML renderer, independent of email attachment configuration.
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
	report := o.newReport("The summary report data is attached as summary.csv.", format)
	report.Attachments = []email.Attachment{{Filename: "summary.csv", ContentType: "text/csv; charset=utf-8", Data: data}}
	return report, nil
}

func (o *Reporter) csv(sources []Source) ([]byte, error) {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	if err := w.Write([]string{"Cluster", "Source", "Scope", "Namespace", "Pass", "Fail", "Warn", "Error", "Skip"}); err != nil {
		return nil, err
	}
	sources = slices.Clone(sources)
	sort.Slice(sources, func(i, j int) bool { return sources[i].Name < sources[j].Name })
	for _, source := range sources {
		write := func(scope, ns string, s *Summary) error {
			return w.Write([]string{email.CSVText(o.clusterName), email.CSVText(source.Name), scope, email.CSVText(ns), strconv.Itoa(s.Pass), strconv.Itoa(s.Fail), strconv.Itoa(s.Warn), strconv.Itoa(s.Error), strconv.Itoa(s.Skip)})
		}
		if source.ClusterReports && source.ClusterScopeSummary != nil {
			if err := write("cluster", "", source.ClusterScopeSummary); err != nil {
				return nil, err
			}
		}
		namespaces := make([]string, 0, len(source.NamespaceScopeSummary))
		for ns := range source.NamespaceScopeSummary {
			namespaces = append(namespaces, ns)
		}
		sort.Strings(namespaces)
		for _, ns := range namespaces {
			if err := write("namespace", ns, source.NamespaceScopeSummary[ns]); err != nil {
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
