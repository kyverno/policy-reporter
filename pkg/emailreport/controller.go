package emailreport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchclient "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/tools/cache"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	crds "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned"
	informerfactory "github.com/kyverno/policy-reporter/pkg/crd/client/informers/externalversions"
)

const (
	Finalizer = "policyreporter.kyverno.io/email-report-cleanup"

	SourceNamespaceAnnotation = "policyreporter.kyverno.io/email-report-namespace"
	SourceNameAnnotation      = "policyreporter.kyverno.io/email-report-name"
	SourceUIDAnnotation       = "policyreporter.kyverno.io/email-report-uid"
)

type Controller struct {
	reports     crds.Interface
	cronJobs    batchclient.CronJobInterface
	namespace   string
	jobTemplate batchv1.JobTemplateSpec
	factory     informerfactory.SharedInformerFactory
	informer    cache.SharedIndexInformer
}

func NewController(reports crds.Interface, kubeClient kubernetes.Interface, namespace string, jobTemplate batchv1.JobTemplateSpec) *Controller {
	factory := informerfactory.NewSharedInformerFactory(reports, time.Minute)
	return &Controller{
		reports:     reports,
		cronJobs:    kubeClient.BatchV1().CronJobs(namespace),
		namespace:   namespace,
		jobTemplate: jobTemplate,
		factory:     factory,
		informer:    factory.Policyreporter().V1alpha1().EmailReports().Informer(),
	}
}

func (c *Controller) Run(ctx context.Context) error {
	_, err := c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			c.reconcileEvent(ctx, obj)
		},
		UpdateFunc: func(_, obj any) {
			c.reconcileEvent(ctx, obj)
		},
		DeleteFunc: func(obj any) {
			c.reconcileEvent(ctx, obj)
		},
	})
	if err != nil {
		return err
	}

	c.factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return fmt.Errorf("failed to sync EmailReport informer")
	}

	<-ctx.Done()
	c.factory.Shutdown()
	return nil
}

func (c *Controller) reconcileEvent(ctx context.Context, obj any) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		zap.L().Error("failed to identify EmailReport", zap.Error(err))
		return
	}
	if err := c.Reconcile(ctx, key); err != nil {
		zap.L().Error("failed to reconcile EmailReport", zap.String("key", key), zap.Error(err))
	}
}

func (c *Controller) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil || namespace == "" {
		return fmt.Errorf("invalid EmailReport key %q", key)
	}

	report, err := c.reports.PolicyreporterV1alpha1().EmailReports(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return c.deleteCronJob(ctx, namespace, name)
	}
	if err != nil {
		return err
	}

	if report.DeletionTimestamp != nil {
		if err := c.deleteCronJob(ctx, namespace, name); err != nil {
			return err
		}
		if contains(report.Finalizers, Finalizer) {
			updated := report.DeepCopy()
			updated.Finalizers = remove(updated.Finalizers, Finalizer)
			_, err = c.reports.PolicyreporterV1alpha1().EmailReports(namespace).Update(ctx, updated, metav1.UpdateOptions{})
			return err
		}
		return nil
	}

	if err := Validate(report.Spec); err != nil {
		if deleteErr := c.deleteCronJob(ctx, namespace, name); deleteErr != nil {
			return fmt.Errorf("invalid EmailReport: %w; failed to remove previous CronJob: %v", err, deleteErr)
		}
		return err
	}

	if !contains(report.Finalizers, Finalizer) {
		updated := report.DeepCopy()
		updated.Finalizers = append(updated.Finalizers, Finalizer)
		_, err = c.reports.PolicyreporterV1alpha1().EmailReports(namespace).Update(ctx, updated, metav1.UpdateOptions{})
		return err
	}

	desired, err := c.desiredCronJob(report)
	if err != nil {
		return err
	}
	existing, err := c.cronJobs.Get(ctx, desired.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = c.cronJobs.Create(ctx, desired, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	if err != nil {
		return err
	}
	if !matchesSource(existing, namespace, name) {
		return fmt.Errorf("CronJob %s/%s is not managed for EmailReport %s", existing.Namespace, existing.Name, key)
	}
	if apiequality.Semantic.DeepEqual(existing.Spec, desired.Spec) &&
		apiequality.Semantic.DeepEqual(existing.Labels, desired.Labels) &&
		apiequality.Semantic.DeepEqual(existing.Annotations, desired.Annotations) {
		return nil
	}

	desired.ResourceVersion = existing.ResourceVersion
	_, err = c.cronJobs.Update(ctx, desired, metav1.UpdateOptions{})
	return err
}

func (c *Controller) desiredCronJob(report *policyreporterv1alpha1.EmailReport) (*batchv1.CronJob, error) {
	template := c.jobTemplate.DeepCopy()
	container := -1
	for i := range template.Spec.Template.Spec.Containers {
		if template.Spec.Template.Spec.Containers[i].Name == "policy-reporter" {
			container = i
			break
		}
	}
	if container < 0 {
		return nil, fmt.Errorf("EmailReport job template has no policy-reporter container")
	}

	template.Spec.Template.Spec.Containers[container].Command = []string{"/app/policyreporter", "send", "email-report"}
	template.Spec.Template.Spec.Containers[container].Args = append(
		template.Spec.Template.Spec.Containers[container].Args,
		"--namespace="+report.Namespace,
		"--name="+report.Name,
		"--uid="+string(report.UID),
	)

	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      CronJobName(report.Namespace, report.Name),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "policy-reporter",
			},
			Annotations: map[string]string{
				SourceNamespaceAnnotation: report.Namespace,
				SourceNameAnnotation:      report.Name,
				SourceUIDAnnotation:       string(report.UID),
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          report.Spec.Schedule,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			JobTemplate:       *template,
		},
	}, nil
}

func (c *Controller) deleteCronJob(ctx context.Context, namespace, name string) error {
	cronJobName := CronJobName(namespace, name)
	existing, err := c.cronJobs.Get(ctx, cronJobName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !matchesSource(existing, namespace, name) {
		return fmt.Errorf("refusing to delete CronJob %s/%s with different source annotations", existing.Namespace, existing.Name)
	}
	propagation := metav1.DeletePropagationBackground
	return c.cronJobs.Delete(ctx, cronJobName, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

func CronJobName(namespace, name string) string {
	sum := sha256.Sum256([]byte(namespace + "\x00" + name))
	return "email-report-" + hex.EncodeToString(sum[:])[:20]
}

func matchesSource(cronJob *batchv1.CronJob, namespace, name string) bool {
	return cronJob.Annotations[SourceNamespaceAnnotation] == namespace && cronJob.Annotations[SourceNameAnnotation] == name
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func remove(values []string, value string) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}
