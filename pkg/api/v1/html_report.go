package v1

import (
	"net/http"
	"slices"

	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"go.uber.org/zap"
)

type HTMLHandler struct {
	reporter *violations.Reporter
	finder   PolicyReportFinder
}

func (h *HTMLHandler) HTMLReport() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		sources := make([]violations.Source, 0)

		namespaced, err := h.finder.FetchNamespacedSources(req.Context())
		if err != nil {
			zap.L().Error("failed to load data", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cluster, err := h.finder.FetchClusterSources(req.Context())
		if err != nil {
			zap.L().Error("failed to load data", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		list := append(namespaced, cluster...)
		slices.Sort(list)
		list = slices.Compact(list)

		for _, source := range list {
			cPass, err := h.finder.CountClusterResults(req.Context(), Filter{
				Sources: []string{source},
				Status:  []string{"pass"},
			})
			if err != nil {
				continue
			}

			statusCounts, err := h.finder.FetchNamespacedStatusCounts(req.Context(), Filter{
				Sources: []string{source},
				Status:  []string{"pass"},
			})
			if err != nil {
				continue
			}

			nsPass := make(map[string]int, len(statusCounts))
			for _, s := range statusCounts[0].Items {
				nsPass[s.Namespace] = s.Count
			}

			clusterResults, err := h.finder.FetchClusterResults(req.Context(), Filter{
				Sources: []string{source},
				Status:  []string{"warn", "fail", "error"},
			}, Pagination{SortBy: defaultOrder})
			if err != nil {
				continue
			}

			cResults := make(map[string][]violations.Result)
			for _, r := range clusterResults {
				if _, ok := cResults[r.Status]; !ok {
					cResults[r.Status] = make([]violations.Result, 0)
				}

				cResults[r.Status] = append(cResults[r.Status], violations.Result{
					Kind:   r.Kind,
					Name:   r.Name,
					Policy: r.Policy,
					Rule:   r.Rule,
					Status: r.Status,
				})
			}

			namespaces, err := h.finder.FetchNamespaces(req.Context(), Filter{
				Sources: []string{source},
			})
			if err != nil {
				continue
			}

			nsResults := make(map[string]map[string][]violations.Result)
			for _, ns := range namespaces {
				results, err := h.finder.FetchNamespacedResults(req.Context(), Filter{
					Sources:    []string{source},
					Status:     []string{"warn", "fail", "error"},
					Namespaces: []string{ns},
				}, Pagination{SortBy: defaultOrder})
				if err != nil {
					continue
				}

				mapping := make(map[string][]violations.Result)
				mapping["warn"] = make([]violations.Result, 0)
				mapping["fail"] = make([]violations.Result, 0)
				mapping["error"] = make([]violations.Result, 0)

				for _, r := range results {
					mapping[r.Status] = append(mapping[r.Status], violations.Result{
						Kind:   r.Kind,
						Name:   r.Name,
						Policy: r.Policy,
						Rule:   r.Rule,
						Status: r.Status,
					})
				}

				nsResults[ns] = mapping
			}

			sources = append(sources, violations.Source{
				Name:             source,
				ClusterReports:   len(cluster) > 0,
				ClusterPassed:    cPass,
				ClusterResults:   cResults,
				NamespacePassed:  nsPass,
				NamespaceResults: nsResults,
			})
		}

		data, err := h.reporter.Report(sources, "HTML")
		if err != nil {
			zap.L().Error("failed to load data", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(data.Message))
	}
}

func NewHTMLHandler(finder PolicyReportFinder, reporter *violations.Reporter) *HTMLHandler {
	return &HTMLHandler{
		finder:   finder,
		reporter: reporter,
	}
}
