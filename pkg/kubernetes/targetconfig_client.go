package kubernetes

import (
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
)

type TargetConfigClient struct {
	tcClient tcv1alpha1.Interface
}
