package openreports

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
)

type ResultAdapter struct {
	ID string
	v1alpha1.ReportResult
}

func (r *ResultAdapter) GetResource() *corev1.ObjectReference {
	if len(r.Subjects) == 0 {
		return nil
	}

	return &r.Subjects[0]
}

func (r *ResultAdapter) HasResource() bool {
	return len(r.Subjects) > 0
}

func (r *ResultAdapter) GetKind() string {
	if !r.HasResource() {
		return ""
	}

	return r.GetResource().Kind
}

func (r *ResultAdapter) GetID() string {
	return r.ID
}

func ToResourceString(res *corev1.ObjectReference) string {
	var resource string

	if res.Namespace != "" {
		resource = res.Namespace
	}

	if res.Kind != "" && resource != "" {
		resource = fmt.Sprintf("%s/%s", resource, strings.ToLower(res.Kind))
	} else if res.Kind != "" {
		resource = strings.ToLower(res.Kind)
	}

	if res.Name != "" && resource != "" {
		resource = fmt.Sprintf("%s/%s", resource, res.Name)
	} else if res.Name != "" {
		resource = res.Name
	}

	return resource
}

func ToResourceID(res *corev1.ObjectReference) string {
	if res == nil {
		return ""
	}

	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, res.Namespace)
	h1 = fnv1a.AddString64(h1, res.Name)
	h1 = fnv1a.AddString64(h1, res.Kind)
	h1 = fnv1a.AddString64(h1, string(res.UID))
	h1 = fnv1a.AddString64(h1, res.APIVersion)

	return strconv.FormatUint(h1, 10)
}

func (r *ResultAdapter) ResourceString() string {
	if !r.HasResource() {
		return ""
	}

	return ToResourceString(r.GetResource())
}
