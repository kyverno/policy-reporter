package emailreport_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	crdfake "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/emailreport"
)

const controllerNamespace = "policy-reporter"

func newEmailReport(namespace, name, schedule string) *policyreporterv1alpha1.EmailReport {
	return &policyreporterv1alpha1.EmailReport{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			UID:       types.UID(namespace + "-uid"),
		},
		Spec: policyreporterv1alpha1.EmailReportSpec{
			Type:     policyreporterv1alpha1.EmailReportTypeSummary,
			Schedule: schedule,
			To:       []string{namespace + "@example.test"},
		},
	}
}

func jobTemplate() batchv1.JobTemplateSpec {
	return batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Containers: []corev1.Container{{
			Name:  "policy-reporter",
			Image: "local/policy-reporter:test",
			Args:  []string{"--config=/app/config.yaml", "--template-dir=/app/templates"},
		}},
	}}}}
}

func TestControllerReconcileLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	report := newEmailReport("team-a", "daily", "*/5 * * * *")
	crdClient := crdfake.NewSimpleClientset(report)
	kubeClient := k8sfake.NewSimpleClientset()
	controller := emailreport.NewController(crdClient, kubeClient, controllerNamespace, jobTemplate())
	key := "team-a/daily"

	require.NoError(t, controller.Reconcile(ctx, key))
	current, err := crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Get(ctx, "daily", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Contains(t, current.Finalizers, emailreport.Finalizer)

	require.NoError(t, controller.Reconcile(ctx, key))
	require.NoError(t, controller.Reconcile(ctx, key))
	cronJobs, err := kubeClient.BatchV1().CronJobs(controllerNamespace).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, cronJobs.Items, 1)
	cronJob := &cronJobs.Items[0]
	assert.Equal(t, "*/5 * * * *", cronJob.Spec.Schedule)
	assert.Equal(t, batchv1.ForbidConcurrent, cronJob.Spec.ConcurrencyPolicy)
	assert.Equal(t, []string{"/app/policyreporter", "send", "email-report"}, cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command)
	assert.ElementsMatch(t, []string{
		"--config=/app/config.yaml",
		"--template-dir=/app/templates",
		"--namespace=team-a",
		"--name=daily",
		"--uid=team-a-uid",
	}, cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args)

	current.Spec.Schedule = "0 9 * * *"
	_, err = crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Update(ctx, current, metav1.UpdateOptions{})
	require.NoError(t, err)
	require.NoError(t, controller.Reconcile(ctx, key))
	cronJob, err = kubeClient.BatchV1().CronJobs(controllerNamespace).Get(ctx, emailreport.CronJobName("team-a", "daily"), metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "0 9 * * *", cronJob.Spec.Schedule)

	current, err = crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Get(ctx, "daily", metav1.GetOptions{})
	require.NoError(t, err)
	current.Spec.Schedule = "invalid"
	_, err = crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Update(ctx, current, metav1.UpdateOptions{})
	require.NoError(t, err)
	require.Error(t, controller.Reconcile(ctx, key))
	_, err = kubeClient.BatchV1().CronJobs(controllerNamespace).Get(ctx, emailreport.CronJobName("team-a", "daily"), metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))
}

func TestControllerUsesCurrentResourceState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	report := newEmailReport("team-a", "daily", "*/5 * * * *")
	crdClient := crdfake.NewSimpleClientset(report)
	kubeClient := k8sfake.NewSimpleClientset()
	controller := emailreport.NewController(crdClient, kubeClient, controllerNamespace, jobTemplate())

	require.NoError(t, controller.Reconcile(ctx, "team-a/daily"))
	current, err := crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Get(ctx, "daily", metav1.GetOptions{})
	require.NoError(t, err)
	current.Spec.Schedule = "0 10 * * *"
	_, err = crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Update(ctx, current, metav1.UpdateOptions{})
	require.NoError(t, err)

	// Reconciliation receives only the key and must fetch this latest value.
	require.NoError(t, controller.Reconcile(ctx, "team-a/daily"))
	cronJob, err := kubeClient.BatchV1().CronJobs(controllerNamespace).Get(ctx, emailreport.CronJobName("team-a", "daily"), metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "0 10 * * *", cronJob.Spec.Schedule)
}

func TestControllerSeparatesSameNameAcrossNamespaces(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	crdClient := crdfake.NewSimpleClientset(
		newEmailReport("team-a", "report", "*/1 * * * *"),
		newEmailReport("team-b", "report", "*/2 * * * *"),
	)
	kubeClient := k8sfake.NewSimpleClientset()
	controller := emailreport.NewController(crdClient, kubeClient, controllerNamespace, jobTemplate())

	for _, key := range []string{"team-a/report", "team-b/report"} {
		require.NoError(t, controller.Reconcile(ctx, key))
		require.NoError(t, controller.Reconcile(ctx, key))
	}
	cronJobs, err := kubeClient.BatchV1().CronJobs(controllerNamespace).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, cronJobs.Items, 2)
	assert.NotEqual(t, emailreport.CronJobName("team-a", "report"), emailreport.CronJobName("team-b", "report"))
}

func TestControllerDeletionRemovesCronJobBeforeFinalizer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	report := newEmailReport("team-a", "daily", "*/5 * * * *")
	report.Finalizers = []string{emailreport.Finalizer}
	deletionTime := metav1.NewTime(time.Now())
	report.DeletionTimestamp = &deletionTime
	crdClient := crdfake.NewSimpleClientset(report)
	kubeClient := k8sfake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: controllerNamespace,
			Name:      emailreport.CronJobName("team-a", "daily"),
			Annotations: map[string]string{
				emailreport.SourceNamespaceAnnotation: "team-a",
				emailreport.SourceNameAnnotation:      "daily",
			},
		},
	})
	controller := emailreport.NewController(crdClient, kubeClient, controllerNamespace, jobTemplate())

	require.NoError(t, controller.Reconcile(ctx, "team-a/daily"))
	_, err := kubeClient.BatchV1().CronJobs(controllerNamespace).Get(ctx, emailreport.CronJobName("team-a", "daily"), metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))
	current, err := crdClient.PolicyreporterV1alpha1().EmailReports("team-a").Get(ctx, "daily", metav1.GetOptions{})
	require.NoError(t, err)
	assert.NotContains(t, current.Finalizers, emailreport.Finalizer)
}
