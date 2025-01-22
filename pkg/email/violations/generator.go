package violations

import (
	"context"
	"sync"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	api "github.com/kyverno/policy-reporter/pkg/crd/client/policy-report/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
)

type Generator struct {
	client         api.Wgpolicyk8sV1alpha2Interface
	filter         email.Filter
	clusterReports bool
}

func (o *Generator) GenerateData(ctx context.Context) ([]Source, error) {
	mx := &sync.Mutex{}

	sources := make(map[string]*Source)
	wg := &sync.WaitGroup{}

	if o.clusterReports {
		clusterReports, err := o.client.ClusterPolicyReports().List(ctx, v1.ListOptions{})
		if err != nil {
			return make([]Source, 0, 0), err
		}

		wg.Add(len(clusterReports.Items))

		for _, rep := range clusterReports.Items {
			go func(report v1alpha2.ClusterPolicyReport) {
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
					if result.Result == v1alpha2.StatusPass || result.Result == v1alpha2.StatusSkip {
						continue
					}

					s.AddClusterResults(mapResult(&report, result))
				}
			}(rep)
		}
	}

	reports, err := o.client.PolicyReports(v1.NamespaceAll).List(ctx, v1.ListOptions{})
	if err != nil {
		return make([]Source, 0, 0), err
	}

	wg.Add(len(reports.Items))

	for _, rep := range reports.Items {
		go func(report v1alpha2.PolicyReport) {
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
				if result.Result == v1alpha2.StatusPass || result.Result == v1alpha2.StatusSkip {
					continue
				}
				s.AddNamespacedResults(report.Namespace, mapResult(&report, result))
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

func NewGenerator(client api.Wgpolicyk8sV1alpha2Interface, filter email.Filter, clusterReports bool) *Generator {
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
