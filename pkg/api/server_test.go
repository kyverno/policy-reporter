package api_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var logger = zap.NewNop()

func Test_NewServer(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().Unix())).Float64()
	if rnd < 0.3 {
		rnd += 0.4
	}

	port := int(rnd * 10000)

	server := api.NewServer(
		make([]target.Client, 0),
		port,
		logger,
		nil,
		func() bool { return true },
	)

	server.RegisterMetricsHandler()
	server.RegisterV1Handler(nil)
	server.RegisterProfilingHandler()

	serviceRunning := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		close(serviceRunning)
		err := server.Start()
		if err != nil {
			fmt.Println(err)
		}
		defer close(serviceDone)
	}()

	<-serviceRunning

	client := http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ready", port), nil)
	if err != nil {
		return
	}

	res, err := client.Do(req)

	server.Shutdown(context.Background())
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected Error Code: %d", res.StatusCode)
	}

	<-serviceDone
}

func Test_SetupServerWithAuth(t *testing.T) {
	server := api.NewServer(
		make([]target.Client, 0),
		8080,
		logger,
		&api.BasicAuth{Username: "user", Password: "password"},
		func() bool { return true },
	)

	server.RegisterMetricsHandler()
	server.RegisterV1Handler(nil)
	server.RegisterProfilingHandler()
}
