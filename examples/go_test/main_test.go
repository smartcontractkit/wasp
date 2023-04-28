package main

import (
	"testing"
	"time"

	"github.com/smartcontractkit/wasp"
)

func TestGenUsageWithTests(t *testing.T) {
	// start mock servers
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	// define labels for differentiate one run from another
	labels := map[string]string{
		// check variables in dashboard/dashboard.go
		"gen_name": "generator_healthcheck",
		"branch":   "generator_healthcheck",
		"commit":   "generator_healthcheck",
	}

	g, err := wasp.NewGenerator(&wasp.Config{
		// T fills "go_test_name" label implicitly
		T:        t,
		LoadType: wasp.RPS,
		// just use plain line profile - 5 RPS for 10s
		Schedule:   wasp.Plain(5, 10*time.Second),
		Gun:        NewExampleHTTPGun(srv.URL()),
		Labels:     labels,
		LokiConfig: wasp.NewEnvLokiConfig(),
	})
	if err != nil {
		panic(err)
	}
	g.Run(true)
}
