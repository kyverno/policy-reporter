package api_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/api"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_NewServer(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().Unix())).Float64()
	if rnd < 0.3 {
		rnd += 0.4
	}

	port := int(rnd * 10000)

	server := api.NewServer(make([]target.Client, 0), port, make(map[string]string))

	server.RegisterMetricsHandler()
	server.RegisterV1Handler(nil)

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
		t.Errorf("Unexpected Error: %s", err)
		return
	}

	res, err := client.Do(req)

	server.Shutdown(context.Background())

	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected Error Code: %d", res.StatusCode)
	}

	<-serviceDone
}
