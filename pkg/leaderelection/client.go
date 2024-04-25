package leaderelection

import (
	"context"
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	k8sleaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Client struct {
	client          v1.CoordinationV1Interface
	lockName        string
	namespace       string
	identity        string
	leaseDuration   time.Duration
	renewDeadline   time.Duration
	retryPeriod     time.Duration
	releaseOnCancel bool

	onStartedLeading func(c context.Context)
	onStoppedLeading func()
	onNewLeader      func(currentID, lockID string)
}

func (c *Client) RegisterOnStart(callback func(c context.Context)) *Client {
	c.onStartedLeading = callback

	return c
}

func (c *Client) RegisterOnStop(callback func()) *Client {
	c.onStoppedLeading = callback

	return c
}

func (c *Client) RegisterOnNew(callback func(currentID string, lockID string)) *Client {
	c.onNewLeader = callback

	return c
}

func (c *Client) Run(ctx context.Context) error {
	k8sleaderelection.RunOrDie(ctx, k8sleaderelection.LeaderElectionConfig{
		Lock:            c.CreateLock(),
		ReleaseOnCancel: c.releaseOnCancel,
		LeaseDuration:   c.leaseDuration,
		RenewDeadline:   c.renewDeadline,
		RetryPeriod:     c.retryPeriod,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: c.onStartedLeading,
			OnStoppedLeading: c.onStoppedLeading,
			OnNewLeader: func(identity string) {
				c.onNewLeader(identity, c.identity)
			},
		},
	})

	return errors.New("leaderelection stopped")
}

func (c *Client) CreateLock() *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      c.lockName,
			Namespace: c.namespace,
		},
		Client: c.client,
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: c.identity,
		},
	}
}

func New(
	client v1.CoordinationV1Interface,
	lockName string,
	namespace string,
	identity string,
	leaseDuration time.Duration,
	renewDeadline time.Duration,
	retryPeriod time.Duration,
	releaseOnCancel bool,
) *Client {
	return &Client{
		client,
		lockName,
		namespace,
		identity,
		leaseDuration,
		renewDeadline,
		retryPeriod,
		releaseOnCancel,
		func(c context.Context) {},
		func() {},
		func(currentID, lockID string) {},
	}
}
