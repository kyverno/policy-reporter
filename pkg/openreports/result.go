package openreports

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
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

func (r *ResultAdapter) ResourceString() string {
	if !r.HasResource() {
		return ""
	}

	return ToResourceString(r.GetResource())
}
