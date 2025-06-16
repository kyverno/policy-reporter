package violations

import (
	"context"
	"sync"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	reportsv1alpha1 "openreports.io/apis/openreports.io/v1alpha1"
	"openreports.io/pkg/client/clientset/versioned/typed/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type Generator struct {
	client         v1alpha1.OpenreportsV1alpha1Interface
	filter         email.Filter
	clusterReports bool
}

func (o *Generator) GenerateData(ctx context.Context) ([]Source, error) {
	mx := &sync.Mutex{}

	sources := make(map[string]*Source)
	wg := &sync.WaitGroup{}

	if o.clusterReports {
		clusterReports, err := o.client.ClusterReports().List(ctx, v1.ListOptions{})
		if err != nil {
			return make([]Source, 0), err
		}

		wg.Add(len(clusterReports.Items))

		for _, rep := range clusterReports.Items {
			go func(report reportsv1alpha1.ClusterReport) {
				defer wg.Done()

				if len(report.Results) == 0 {
					return
				}

				rs := report.Results[0].Source

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

				s.AddClusterPassed(report.Summary.Pass)

				zap.L().Info("Processed PolicyRepor", zap.String("name", report.Name))

				length := len(report.Results)
				if length == 0 || length == report.Summary.Pass+report.Summary.Skip {
					return
				}

				for _, result := range report.Results {
					if result.Result == openreports.StatusPass || result.Result == openreports.StatusSkip {
						continue
					}

					s.AddClusterResults(mapResult(&openreports.ORClusterReportAdapter{ClusterReport: &report}, result))
				}
			}(rep)
		}
	}

	reports, err := o.client.Reports(v1.NamespaceAll).List(ctx, v1.ListOptions{})
	if err != nil {
		return make([]Source, 0), err
	}

	wg.Add(len(reports.Items))

	for _, rep := range reports.Items {
		go func(report reportsv1alpha1.Report) {
			defer wg.Done()

			if len(report.Results) == 0 {
				return
			}

			rs := report.Results[0].Source

			if !o.filter.ValidateSource(rs) || !o.filter.ValidateNamespace(report.Namespace) {
				return
			}

			mx.Lock()
			s, ok := sources[rs]
			if !ok {
				s = NewSource(rs, o.clusterReports)
				sources[rs] = s
			}
			mx.Unlock()

			s.AddNamespacedPassed(report.Namespace, report.Summary.Pass)

			defer zap.L().Info("Processed PolicyRepor", zap.String("name", report.Name))

			length := len(report.Results)
			if length == 0 || length == report.Summary.Pass+report.Summary.Skip {
				s.InitResults(report.Namespace)
				return
			}

			for _, result := range report.Results {
				if result.Result == openreports.StatusPass || result.Result == openreports.StatusSkip {
					continue
				}
				s.AddNamespacedResults(report.Namespace, mapResult(&openreports.ORReportAdapter{Report: &report}, result))
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

func NewGenerator(client v1alpha1.OpenreportsV1alpha1Interface, filter email.Filter, clusterReports bool) *Generator {
	return &Generator{client, filter, clusterReports}
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
