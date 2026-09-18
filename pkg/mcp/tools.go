package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	db "github.com/kyverno/policy-reporter/pkg/database"
)

// ComplianceFilter holds the optional filters shared by the compliance tools.
// Namespaces is required and scopes the result to the given namespace(s).
type ComplianceFilter struct {
	Names      []string `json:"names,omitempty" jsonschema:"Only include results for resources of the given resource names"`
	Namespaces []string `json:"namespaces" jsonschema:"One or more namespaces to check compliance for"`
	Sources    []string `json:"sources,omitempty" jsonschema:"Only include results reported by the given policy sources (e.g. kyverno, trivy)"`
	Categories []string `json:"categories,omitempty" jsonschema:"Only include results from the given policy categories"`
	Kinds      []string `json:"kinds,omitempty" jsonschema:"Only include results for resources of the given Kubernetes kinds"`
	Resources  []string `json:"resources,omitempty" jsonschema:"Only include results for resources of the given API groups/versions"`
	Status     []string `json:"status,omitempty" jsonschema:"Only include resources that have at least one result with the given status (pass, fail, warn, error, skip)"`
	Severities []string `json:"severities,omitempty" jsonschema:"Only include resources that have at least one result with the given severity (info, low, medium, high, critical)"`
	Search     string   `json:"search,omitempty" jsonschema:"Free text search matched against the resource name or kind"`
}

func (f ComplianceFilter) toFilter() db.Filter {
	return db.Filter{
		Namespaces:   f.Namespaces,
		Sources:      f.Sources,
		Categories:   f.Categories,
		Kinds:        f.Kinds,
		ResourceAPIs: f.Resources,
		Status:       f.Status,
		Severities:   f.Severities,
		Search:       f.Search,
	}
}

// ResourceComplianceRequest is the input for the list_resource_compliance tool.
type ResourceComplianceRequest struct {
	ComplianceFilter
}

// ResourceCompliance describes the compliance summary of a single resource.
type ResourceCompliance struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Pass      int    `json:"pass"`
	Warn      int    `json:"warn"`
	Fail      int    `json:"fail"`
	Error     int    `json:"error"`
	Skip      int    `json:"skip"`
	Info      int    `json:"info"`
	Low       int    `json:"low"`
	Medium    int    `json:"medium"`
	High      int    `json:"high"`
	Critical  int    `json:"critical"`
	Unknown   int    `json:"unknown"`
}

// ResourceComplianceResponse is the output of the list_resource_compliance tool.
type ResourceComplianceResponse struct {
	Resources []ResourceCompliance `json:"resources"`
	Total     int                  `json:"total"`
}

// ResourceComplianceResultsRequest is the input for the get_resource_compliance_results tool.
type ResourceComplianceResultsRequest struct {
	Namespace  string   `json:"namespace" jsonschema:"Namespace of the resource"`
	Kind       string   `json:"kind" jsonschema:"Kubernetes kind of the resource (e.g. Pod, Deployment)"`
	Name       string   `json:"name" jsonschema:"Name of the resource"`
	Sources    []string `json:"sources,omitempty" jsonschema:"Only include results reported by the given policy sources (e.g. kyverno, trivy)"`
	Categories []string `json:"categories,omitempty" jsonschema:"Only include results from the given policy categories"`
	Search     string   `json:"search,omitempty" jsonschema:"Free text search across namespace, resource name, policy, rule, result and severity"`
}

// ComplianceResult describes a single policy result reported for a resource.
type ComplianceResult struct {
	Source     string            `json:"source"`
	Category   string            `json:"category"`
	Policy     string            `json:"policy"`
	Rule       string            `json:"rule"`
	Result     string            `json:"result"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	Properties map[string]string `json:"properties,omitempty"`
	Timestamp  int64             `json:"timestamp"`
}

// ResourceComplianceResultsResponse is the output of the get_resource_compliance_results tool.
type ResourceComplianceResultsResponse struct {
	Results []ComplianceResult `json:"results"`
	Total   int                `json:"total"`
}

// NamespaceComplianceRequest is the input for the get_namespace_compliance_summary tool.
type NamespaceComplianceRequest struct {
	ComplianceFilter
}

// NamespaceCompliance describes the aggregated compliance counts of a namespace.
type NamespaceCompliance struct {
	Namespace string `json:"namespace"`
	Pass      int    `json:"pass"`
	Warn      int    `json:"warn"`
	Fail      int    `json:"fail"`
	Error     int    `json:"error"`
	Skip      int    `json:"skip"`
	Info      int    `json:"info"`
	Low       int    `json:"low"`
	Medium    int    `json:"medium"`
	High      int    `json:"high"`
	Critical  int    `json:"critical"`
	Unknown   int    `json:"unknown"`
}

func registerTools(s *server.MCPServer, store *db.Store) {
	resourceComplianceTool := mcp.NewTool("list_resource_compliance",
		mcp.WithDescription("List the policy compliance summary (pass/fail/warn/error/skip and severity counts) for resources within one or more namespaces, with optional filters."),
		mcp.WithInputSchema[ResourceComplianceRequest](),
		mcp.WithOutputSchema[ResourceComplianceResponse](),
	)
	s.AddTool(resourceComplianceTool, mcp.NewStructuredToolHandler(listResourceComplianceHandler(store)))

	namespaceComplianceTool := mcp.NewTool("get_namespace_compliance_summary",
		mcp.WithDescription("Get the aggregated policy compliance summary (pass/fail/warn/error/skip and severity counts) per namespace for one or more namespaces, with optional filters."),
		mcp.WithInputSchema[NamespaceComplianceRequest](),
		mcp.WithOutputSchema[[]NamespaceCompliance](),
	)
	s.AddTool(namespaceComplianceTool, mcp.NewStructuredToolHandler(getNamespaceComplianceSummaryHandler(store)))

	resourceResultsTool := mcp.NewTool("get_resource_compliance_results",
		mcp.WithDescription("Get all individual policy compliance results (one entry per policy/rule evaluation) for a single Kubernetes resource, identified by its namespace, kind and name."),
		mcp.WithInputSchema[ResourceComplianceResultsRequest](),
		mcp.WithOutputSchema[ResourceComplianceResultsResponse](),
	)
	s.AddTool(resourceResultsTool, mcp.NewStructuredToolHandler(getResourceComplianceResultsHandler(store)))
}

func listResourceComplianceHandler(store *db.Store) func(context.Context, mcp.CallToolRequest, ResourceComplianceRequest) (ResourceComplianceResponse, error) {
	return func(ctx context.Context, _ mcp.CallToolRequest, args ResourceComplianceRequest) (ResourceComplianceResponse, error) {
		filter := args.toFilter()

		pagination := db.Pagination{
			SortBy:    []string{"resource_namespace", "resource_name"},
			Direction: "ASC",
		}

		results, err := store.FetchNamespaceResourceResults(ctx, filter, pagination)
		if err != nil {
			return ResourceComplianceResponse{}, err
		}

		total, err := store.CountNamespaceResourceResults(ctx, filter)
		if err != nil {
			return ResourceComplianceResponse{}, err
		}

		resources := make([]ResourceCompliance, 0, len(results))
		for _, r := range results {
			resources = append(resources, ResourceCompliance{
				ID:        r.ID,
				Kind:      r.Resource.Kind,
				Name:      r.Resource.Name,
				Namespace: r.Resource.Namespace,
				Pass:      r.Pass,
				Warn:      r.Warn,
				Fail:      r.Fail,
				Error:     r.Error,
				Skip:      r.Skip,
				Info:      r.Info,
				Low:       r.Low,
				Medium:    r.Medium,
				High:      r.High,
				Critical:  r.Critical,
				Unknown:   r.Unknown,
			})
		}

		return ResourceComplianceResponse{
			Resources: resources,
			Total:     total,
		}, nil
	}
}

// getNamespaceComplianceSummaryHandler aggregates the per-resource compliance
// summaries into a single summary per namespace. It reuses
// FetchNamespaceResourceResults without pagination so the same filters and
// semantics as list_resource_compliance apply.
func getNamespaceComplianceSummaryHandler(store *db.Store) func(context.Context, mcp.CallToolRequest, NamespaceComplianceRequest) ([]NamespaceCompliance, error) {
	return func(ctx context.Context, _ mcp.CallToolRequest, args NamespaceComplianceRequest) ([]NamespaceCompliance, error) {
		var err error

		filter := args.toFilter()

		if len(args.Sources) == 0 {
			args.Sources, err = store.FetchSources(ctx, filter)
			if err != nil {
				return nil, err
			}
		}

		order := make([]string, 0)
		summaries := make(map[string]*NamespaceCompliance)

		for _, source := range args.Sources {
			counts, err := store.FetchNamespaceStatusCounts(ctx, source, filter)
			if err != nil {
				return nil, err
			}

			for _, count := range counts {
				ns := count.Namespace

				s, ok := summaries[ns]
				if !ok {
					s = &NamespaceCompliance{Namespace: ns}
					summaries[ns] = s
					order = append(order, ns)
				}

				switch count.Status {
				case "pass":
					s.Pass += count.Count
				case "warn":
					s.Warn += count.Count
				case "fail":
					s.Fail += count.Count
				case "error":
					s.Error += count.Count
				case "skip":
					s.Skip += count.Count
				}
			}
		}

		for _, source := range args.Sources {
			counts, err := store.FetchNamespaceSeverityCounts(ctx, source, filter)
			if err != nil {
				return nil, err
			}

			for _, count := range counts {
				ns := count.Namespace

				s, ok := summaries[ns]
				if !ok {
					s = &NamespaceCompliance{Namespace: ns}
					summaries[ns] = s
					order = append(order, ns)
				}

				switch count.Severity {
				case "low":
					s.Low += count.Count
				case "medium":
					s.Medium += count.Count
				case "high":
					s.High += count.Count
				case "critical":
					s.Critical += count.Count
				case "unknown":
					s.Unknown += count.Count
				}
			}
		}

		list := make([]NamespaceCompliance, 0, len(order))
		for _, ns := range order {
			list = append(list, *summaries[ns])
		}

		return list, nil
	}
}

func getResourceComplianceResultsHandler(store *db.Store) func(context.Context, mcp.CallToolRequest, ResourceComplianceResultsRequest) (ResourceComplianceResultsResponse, error) {
	return func(ctx context.Context, _ mcp.CallToolRequest, args ResourceComplianceResultsRequest) (ResourceComplianceResultsResponse, error) {
		filter := db.Filter{
			Namespaces: []string{args.Namespace},
			Kinds:      []string{args.Kind},
			Resources:  []string{args.Name},
			Sources:    args.Sources,
			Categories: args.Categories,
			Search:     args.Search,
		}

		pagination := db.Pagination{
			SortBy:    []string{"created"},
			Direction: "DESC",
		}

		results, err := store.FetchResults(ctx, true, filter, pagination)
		if err != nil {
			return ResourceComplianceResultsResponse{}, err
		}

		list := make([]ComplianceResult, 0, len(results))
		for _, r := range results {
			list = append(list, ComplianceResult{
				Source:     r.Source,
				Category:   r.Category,
				Policy:     r.Policy,
				Rule:       r.Rule,
				Result:     r.Result,
				Severity:   r.Severity,
				Message:    r.Message,
				Properties: r.Properties,
				Timestamp:  r.Created,
			})
		}

		return ResourceComplianceResultsResponse{
			Results: list,
			Total:   len(list),
		}, nil
	}
}
