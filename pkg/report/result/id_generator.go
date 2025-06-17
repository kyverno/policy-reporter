package result

import (
	"strconv"
	"strings"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type FieldMapperFunc = func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64

type IDGenerator interface {
	Generate(polr openreports.ReportInterface, res openreports.ORResultAdapter) string
}

var fieldMapper = map[string]FieldMapperFunc{
	"resource": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		var resource *corev1.ObjectReference

		if res.HasResource() {
			resource = res.GetResource()
		} else if polr.GetScope() != nil {
			resource = polr.GetScope()
		}

		if resource != nil {
			h1 = fnv1a.AddString64(h1, string(resource.UID))
			h1 = fnv1a.AddString64(h1, string(resource.Name))
		}

		return h1
	},
	"namespace": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, polr.GetNamespace())
	},
	"policy": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, res.Policy)
	},
	"rule": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, res.Rule)
	},
	"result": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, string(res.Result))
	},
	"category": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, res.Category)
	},
	"message": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, res.Description)
	},
	"created": func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
		return fnv1a.AddString64(h1, res.Timestamp.String())
	},
}

var (
	propertyResolver = func(field string) FieldMapperFunc {
		name := strings.TrimPrefix(field, "property:")

		return func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
			if prop, ok := res.Properties[name]; ok {
				h1 = fnv1a.AddString64(h1, prop)
			}

			return h1
		}
	}

	labelResolver = func(field string) FieldMapperFunc {
		name := strings.TrimPrefix(field, "label:")

		return func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
			if prop, ok := polr.GetLabels()[name]; ok {
				h1 = fnv1a.AddString64(h1, prop)
			}

			return h1
		}
	}

	annotationResolver = func(field string) FieldMapperFunc {
		name := strings.TrimPrefix(field, "annotation:")

		return func(h1 uint64, polr openreports.ReportInterface, res openreports.ORResultAdapter) uint64 {
			if prop, ok := polr.GetAnnotations()[name]; ok {
				h1 = fnv1a.AddString64(h1, prop)
			}

			return h1
		}
	}
)

type customIDGenerator struct {
	mappings []FieldMapperFunc
}

func (g *customIDGenerator) Generate(polr openreports.ReportInterface, res openreports.ORResultAdapter) string {
	if id, ok := res.Properties[v1alpha2.ResultIDKey]; ok {
		return id
	}

	h1 := fnv1a.Init64
	for _, mapping := range g.mappings {
		h1 = mapping(h1, polr, res)
	}

	return strconv.FormatUint(h1, 10)
}

type defaultIDGenerator struct{}

func (g *defaultIDGenerator) Generate(polr openreports.ReportInterface, res openreports.ORResultAdapter) string {
	if id, ok := res.Properties[v1alpha2.ResultIDKey]; ok {
		return id
	}

	h1 := fnv1a.Init64

	resource := polr.GetScope()
	if resource == nil {
		resource = res.GetResource()
	}

	if resource != nil {
		h1 = fnv1a.AddString64(h1, resource.Name)
		h1 = fnv1a.AddString64(h1, string(resource.UID))
	}

	h1 = fnv1a.AddString64(h1, res.Policy)
	h1 = fnv1a.AddString64(h1, res.Rule)
	h1 = fnv1a.AddString64(h1, string(res.Result))
	h1 = fnv1a.AddString64(h1, res.Category)
	h1 = fnv1a.AddString64(h1, res.Description)

	return strconv.FormatUint(h1, 10)
}

func NewIDGenerator(config []string) IDGenerator {
	if len(config) == 0 {
		return &defaultIDGenerator{}
	}

	mappings := make([]FieldMapperFunc, 0, len(config))
	for _, field := range config {
		if strings.HasPrefix(field, "property:") {
			mappings = append(mappings, propertyResolver(field))
			continue
		}
		if strings.HasPrefix(field, "label:") {
			mappings = append(mappings, labelResolver(field))
			continue
		}
		if strings.HasPrefix(field, "annotation:") {
			mappings = append(mappings, annotationResolver(field))
			continue
		}

		mappings = append(mappings, fieldMapper[field])
	}

	return &customIDGenerator{mappings}
}
