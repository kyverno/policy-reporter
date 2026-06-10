package violations

import (
	"context"
	"sync"

	"github.com/openreports/reports-api/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type Generator struct {
	openreportsClient v1alpha1.OpenreportsV1alpha1Interface
	wgpolicyClient    v1alpha2.Wgpolicyk8sV1alpha2Interface
	filter            email.Filter
	clusterReports    bool
}

func (o *Generator) GenerateData(ctx context.Context) ([]Source, error) {
	mx := &sync.Mutex{}

	sources := make(map[string]*Source)
	wg := &sync.WaitGroup{}

	if o.clusterReports {
		clusterReports := []openreports.ReportInterface{}

		if o.openreportsClient != nil {
			crs, err := o.openreportsClient.ClusterReports().List(ctx, v1.ListOptions{})
			if err != nil {
				return make([]Source, 0), err
			}
			for _, cr := range crs.Items {
				clusterReports = append(clusterReports, &openreports.ClusterReportAdapter{ClusterReport: &cr})
			}
		}

		if o.wgpolicyClient != nil {
			crs, err := o.wgpolicyClient.ClusterPolicyReports().List(ctx, v1.ListOptions{})
			if err != nil {
				return make([]Source, 0), err
			}
			for _, cr := range crs.Items {
				clusterReports = append(clusterReports, &openreports.ClusterReportAdapter{ClusterReport: cr.ToOpenReports()})
			}
		}

		wg.Add(len(clusterReports))

		for _, rep := range clusterReports {
			go func(report openreports.ReportInterface) {
				defer wg.Done()

				if len(report.GetResults()) == 0 {
					return
				}

				rs := report.GetSource()

				if !o.filter.ValidateSource(rs) {
					return
				}

				mx.Lock()
				s, ok := sources[rs]
				if !ok {
					s = NewSource(rs, o.clusterReports)
					sources[rs] = s
				}
				mx.Unlock()

				s.AddClusterPassed(report.GetSummary().Pass)

				zap.L().Info("Processed Report", zap.String("name", report.GetName()))

				length := len(report.GetResults())
				if length == 0 || length == report.GetSummary().Pass+report.GetSummary().Skip {
					return
				}

				for _, result := range report.GetResults() {
					if result.Result == openreports.StatusPass || result.Result == openreports.StatusSkip {
						continue
					}

					s.AddClusterResults(mapResult(report, result))
				}
			}(rep)
		}
	}

	reports := []openreports.ReportInterface{}

	if o.openreportsClient != nil {
		rs, err := o.openreportsClient.Reports(v1.NamespaceAll).List(ctx, v1.ListOptions{})
		if err != nil {
			return make([]Source, 0), err
		}
		for _, r := range rs.Items {
			reports = append(reports, &openreports.ReportAdapter{Report: &r})
		}
	}

	if o.wgpolicyClient != nil {
		crs, err := o.wgpolicyClient.PolicyReports(v1.NamespaceAll).List(ctx, v1.ListOptions{})
		if err != nil {
			return make([]Source, 0), err
		}
		for _, r := range crs.Items {
			reports = append(reports, &openreports.ReportAdapter{Report: r.ToOpenReports()})
		}
	}

	wg.Add(len(reports))

	for _, rep := range reports {
		go func(report openreports.ReportInterface) {
			defer wg.Done()

			if len(report.GetResults()) == 0 {
				return
			}

			rs := report.GetSource()

			if !o.filter.ValidateSource(rs) || !o.filter.ValidateNamespace(report.GetNamespace()) {
				return
			}

			mx.Lock()
			s, ok := sources[rs]
			if !ok {
				s = NewSource(rs, o.clusterReports)
				sources[rs] = s
			}
			mx.Unlock()

			s.AddNamespacedPassed(report.GetNamespace(), report.GetSummary().Pass)

			defer zap.L().Info("Processed Report", zap.String("name", report.GetName()))

			length := len(report.GetResults())
			if length == 0 || length == report.GetSummary().Pass+report.GetSummary().Skip {
				s.InitResults(report.GetNamespace())
				return
			}

			for _, result := range report.GetResults() {
				if result.Result == openreports.StatusPass || result.Result == openreports.StatusSkip {
					continue
				}
				s.AddNamespacedResults(report.GetNamespace(), mapResult(report, result))
			}
		}(rep)
	}

	wg.Wait()

	list := make([]Source, 0, len(sources))
	for _, s := range sources {
		list = append(list, *s)
	}

	return list, nil
}

func NewGenerator(orclient v1alpha1.OpenreportsV1alpha1Interface, wgpolicyclient v1alpha2.Wgpolicyk8sV1alpha2Interface, filter email.Filter, clusterReports bool) *Generator {
	return &Generator{orclient, wgpolicyclient, filter, clusterReports}
}

func FilterSources(sources []Source, filter email.Filter, clusterReports bool) []Source {
	newSources := make([]Source, 0)

	mx := sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(sources))

	for _, s := range sources {
		go func(source Source) {
			defer wg.Done()

			if !filter.ValidateSource(source.Name) {
				return
			}

			newSource := NewSource(source.Name, clusterReports)

			if clusterReports {
				newSource.ClusterPassed = source.ClusterPassed
				newSource.ClusterResults = source.ClusterResults
			}

			for ns, passed := range source.NamespacePassed {
				if !filter.ValidateNamespace(ns) {
					continue
				}

				newSource.AddNamespacedPassed(ns, passed)
			}

			for ns, results := range source.NamespaceResults {
				if !filter.ValidateNamespace(ns) {
					continue
				}

				newSource.NamespaceResults[ns] = results
			}

			if !clusterReports && len(newSource.NamespaceResults) == 0 {
				return
			}

			mx.Lock()
			newSources = append(newSources, *newSource)
			mx.Unlock()
		}(s)
	}

	wg.Wait()

	return newSources
}
