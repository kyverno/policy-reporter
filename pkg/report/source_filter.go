package report

import (
	"strings"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type PodClient interface {
	Get(res *corev1.ObjectReference) (*corev1.Pod, error)
}

type JobClient interface {
	Get(res *corev1.ObjectReference) (*batchv1.Job, error)
}

type ReportSelector struct {
	Source string
}

type SourceValidation struct {
	Selector              ReportSelector
	Kinds                 validate.RuleSets
	Sources               validate.RuleSets
	Namespaces            validate.RuleSets
	UncontrolledOnly      bool
	DisableClusterReports bool
}

type SourceFilter struct {
	pods        PodClient
	jobs        JobClient
	validations []SourceValidation
}

func (s *SourceFilter) Validate(polr v1alpha2.ReportInterface) bool {
	for _, validation := range s.validations {
		if ok := s.run(polr, validation); !ok {
			return false
		}
	}

	return true
}

func (s *SourceFilter) run(polr v1alpha2.ReportInterface, options SourceValidation) bool {
	logger := zap.L().With(
		zap.String("namespace", polr.GetNamespace()),
		zap.String("report", polr.GetName()),
	)

	if !Match(polr, options.Selector) {
		return true
	}

	if options.DisableClusterReports && polr.GetNamespace() == "" {
		logger.Debug("filter cluster report")
		return false
	}

	if options.Sources.Enabled() && !validate.MatchRuleSet(polr.GetSource(), options.Sources) {
		logger.Debug("filter report source")
		return false
	}

	scope := polr.GetScope()
	if scope == nil {
		return true
	}

	logger = logger.With(zap.Any("scope", scope))

	if options.Kinds.Enabled() && !validate.MatchRuleSet(scope.Kind, options.Kinds) {
		logger.Debug("filter scope resource kind")
		return false
	}

	if options.Namespaces.Enabled() && !validate.MatchRuleSet(scope.Namespace, options.Namespaces) {
		logger.Debug("filter scope resource namespace")
		return false
	}

	if options.UncontrolledOnly && s.pods != nil && scope.Kind == "Pod" {
		pod, err := s.pods.Get(scope)
		if err != nil {
			zap.L().Error("failed to get pod", zap.Error(err), zap.Any("resource", scope))
			return true
		}

		if ok := Uncontrolled(pod.OwnerReferences); ok {
			return true
		}

		logger.Debug("filter controlled pod resource")
		return false
	}

	if options.UncontrolledOnly && s.jobs != nil && scope.Kind == "Job" {
		job, err := s.jobs.Get(scope)
		if err != nil {
			zap.L().Error("failed to get job", zap.Error(err), zap.Any("resource", scope))
			return true
		}

		if ok := Uncontrolled(job.OwnerReferences); ok {
			return true
		}

		logger.Debug("filter controlled job resource")
		return false
	}

	return true
}

func NewSourceFilter(pods PodClient, jobs JobClient, validations []SourceValidation) *SourceFilter {
	return &SourceFilter{pods: pods, jobs: jobs, validations: validations}
}

var controller = []string{"ReplicaSet", "DaemonSet", "CronJob", "Job"}

func Uncontrolled(owner []metav1.OwnerReference) bool {
	if len(owner) == 0 {
		return true
	}

	for _, o := range owner {
		isController := o.Controller
		if isController == nil {
			continue
		}

		if *isController == true && helper.Contains(o.Kind, controller) {
			return false
		}
	}

	return true
}

func Match(polr v1alpha2.ReportInterface, selector ReportSelector) bool {
	return selector.Source == "" || strings.ToLower(selector.Source) == strings.ToLower(polr.GetSource())
}
