package report

import (
	"fmt"
	"strings"

	"github.com/go-openapi/inflect"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gocache "zgo.at/zcache/v2"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/jobs"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/pods"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/replicasets"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type ReportSelector struct {
	Source  string
	Sources []string `mapstructure:"sources"`
}

type SourceValidation struct {
	Selector              ReportSelector
	Kinds                 validate.RuleSets
	Resources             validate.RuleSets
	Sources               validate.RuleSets
	Namespaces            validate.RuleSets
	UncontrolledOnly      bool
	DisableClusterReports bool
}

type SourceFilter struct {
	pods        pods.Client
	jobs        jobs.Client
	replicasets replicasets.Client
	validations []SourceValidation
	controlled  *gocache.Cache[types.UID, bool]
}

func (s *SourceFilter) Validate(polr openreports.ReportInterface) bool {
	for _, validation := range s.validations {
		if ok := s.run(polr, validation); !ok {
			return false
		}
	}

	return true
}

func (s *SourceFilter) run(polr openreports.ReportInterface, options SourceValidation) bool {
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

	logger = logger.With(
		zap.String("kind", scope.Kind),
		zap.String("name", scope.Name),
		zap.String("namespace", scope.Namespace),
		zap.String("uid", string(scope.UID)),
	)

	if options.Kinds.Enabled() && !validate.MatchRuleSet(scope.Kind, options.Kinds) {
		logger.Debug("filter scope resource kind")
		return false
	}

	if options.Resources.Enabled() && !validate.MatchRuleSet(ToAPIString(scope), options.Resources) {
		logger.Debug("filter scope resource api")
		return false
	}

	if options.Namespaces.Enabled() && !validate.MatchRuleSet(scope.Namespace, options.Namespaces) {
		logger.Debug("filter scope resource namespace")
		return false
	}

	if !options.UncontrolledOnly {
		return true
	}

	if ok, controlled := s.controlled.Get(scope.UID); ok {
		logger.Debug("resource found in cache", zap.Bool("filter", controlled))
		return controlled
	}

	if s.pods != nil && scope.Kind == "Pod" {
		pod, err := s.pods.Get(scope)
		if err != nil {
			logger.Error("failed to get pod", zap.Error(err), zap.String("name", scope.Name), zap.String("namespace", scope.Namespace))
			return true
		}

		if ok := Uncontrolled(pod.OwnerReferences, podControllers); ok {
			s.cache(pod.UID, true)
			return true
		}

		logger.Debug("filter controlled pod resource")
		s.cache(pod.UID, false)
		return false
	}

	if s.replicasets != nil && scope.Kind == "ReplicaSet" && scope.APIVersion == "apps/v1" {
		rs, err := s.replicasets.Get(scope)
		if err != nil {
			logger.Error("failed to get replicaset", zap.Error(err))
			return true
		}

		if ok := Uncontrolled(rs.OwnerReferences, rsControllers); ok {
			s.cache(rs.UID, true)
			return true
		}

		logger.Debug("filter controlled replicaset resource")
		s.cache(rs.UID, false)
		return false
	}

	if s.jobs != nil && scope.Kind == "Job" {
		job, err := s.jobs.Get(scope)
		if err != nil {
			logger.Error("failed to get job", zap.Error(err))
			return true
		}

		if ok := Uncontrolled(job.OwnerReferences, jobControllers); ok {
			s.cache(job.UID, true)
			return true
		}

		logger.Debug("filter controlled job resource")
		s.cache(job.UID, false)
		return false
	}

	return true
}

func (s *SourceFilter) cache(uid types.UID, controlled bool) {
	if uid == "" || s.controlled == nil {
		return
	}

	s.controlled.Set(uid, controlled)
}

func NewSourceFilter(pods pods.Client, jobs jobs.Client, rs replicasets.Client, cache *gocache.Cache[types.UID, bool], validations []SourceValidation) *SourceFilter {
	return &SourceFilter{pods: pods, jobs: jobs, replicasets: rs, controlled: cache, validations: validations}
}

var podControllers = map[string]bool{
	"apps/v1/ReplicaSet":  true,
	"apps/v1/StatefulSet": true,
	"batch/v1/Job":        true,
}

var jobControllers = map[string]bool{
	"batch/v1/CronJob": true,
}

var rsControllers = map[string]bool{
	"apps/v1/Deployment": true,
}

func Uncontrolled(owner []metav1.OwnerReference, controllers map[string]bool) bool {
	if len(owner) == 0 {
		return true
	}

	for _, o := range owner {
		isController := o.Controller
		if isController == nil {
			continue
		}

		if *isController && controllers[ToResourceString(o)] {
			return false
		}
	}

	return true
}

func Match(polr openreports.ReportInterface, selector ReportSelector) bool {
	if len(selector.Sources) > 0 {
		return helper.Contains(polr.GetSource(), selector.Sources)
	}

	return selector.Source == "" || strings.EqualFold(selector.Source, polr.GetSource())
}

func ToResourceString(o metav1.OwnerReference) string {
	return fmt.Sprintf("%s/%s", o.APIVersion, o.Kind)
}

func ToAPIString(res *corev1.ObjectReference) string {
	resource := inflect.Pluralize(strings.ToLower(res.Kind))

	return fmt.Sprintf("%s.%s", resource, res.APIVersion)
}
