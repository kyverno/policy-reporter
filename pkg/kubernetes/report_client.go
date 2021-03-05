package kubernetes

import (
	"sync"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"golang.org/x/sync/errgroup"
)

type resultClient struct {
	policyClient        report.PolicyClient
	clusterPolicyClient report.ClusterPolicyClient
}

func (c *resultClient) FetchPolicyResults() ([]report.Result, error) {
	g := new(errgroup.Group)
	mx := new(sync.Mutex)

	var results []report.Result

	g.Go(func() error {
		rs, err := c.policyClient.FetchPolicyResults()
		if err != nil {
			return err
		}

		mx.Lock()
		results = append(results, rs...)
		mx.Unlock()

		return nil
	})

	g.Go(func() error {
		rs, err := c.clusterPolicyClient.FetchPolicyResults()
		if err != nil {
			return err
		}

		mx.Lock()
		results = append(results, rs...)
		mx.Unlock()

		return nil
	})

	return results, g.Wait()
}

func (c *resultClient) RegisterPolicyResultWatcher(skipExisting bool) {
	c.policyClient.RegisterPolicyResultWatcher(skipExisting)
	c.clusterPolicyClient.RegisterPolicyResultWatcher(skipExisting)
}

func (c *resultClient) RegisterPolicyResultCallback(cb report.PolicyResultCallback) {
	c.policyClient.RegisterPolicyResultCallback(cb)
	c.clusterPolicyClient.RegisterPolicyResultCallback(cb)
}

// NewPolicyResultClient creates a new ReportClient based on the kubernetes go-client
func NewPolicyResultClient(policyClient report.PolicyClient, clusterPolicyClient report.ClusterPolicyClient) report.ResultClient {
	return &resultClient{
		policyClient,
		clusterPolicyClient,
	}
}
