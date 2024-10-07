package formatting

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func ResourceString(res *corev1.ObjectReference) string {
	var resource string
	if res.Namespace == "" {
		resource = fmt.Sprintf("%s/%s: %s", res.APIVersion, res.Kind, res.Name)
	} else {
		resource = fmt.Sprintf("%s/%s: %s/%s", res.APIVersion, res.Kind, res.Namespace, res.Name)
	}

	return strings.Trim(resource, "/")
}
