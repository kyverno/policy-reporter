package summary

import (
	"context"
	"log"
	"sync"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

				s.AddClusterSummary(report.Summary)

				log.Printf("[INFO] Processed ClusterPolicyReport '%s'\n", report.Name)
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

			if len(report.Results) == 0 || !o.filter.ValidateNamespace(report.Namespace) {
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

			s.AddNamespacedSummary(report.Namespace, report.Summary)

			log.Printf("[INFO] Processed PolicyReport '%s'\n", report.Name)
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
				newSource.ClusterScopeSummary = source.ClusterScopeSummary
			}

			for ns, results := range source.NamespaceScopeSummary {
				if !filter.ValidateNamespace(ns) {
					continue
				}

				newSource.NamespaceScopeSummary[ns] = results
			}

			if !clusterReports && len(newSource.NamespaceScopeSummary) == 0 {
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
