package main

import (
	"time"

	"github.com/smartcontractkit/wasp"
)

func main() {
	// start mock http server
	srv := wasp.NewHTTPMockServer(50 * time.Millisecond)
	srv.Run()

	// define labels for differentiate one run from another
	labels := map[string]string{
		// check variables in dashboard/dashboard.go
		"go_test_name": "simple_rps",
		"branch":       "generator_healthcheck",
		"commit":       "generator_healthcheck",
	}

	// create generator
	gen, err := wasp.NewGenerator(&wasp.Config{
		LoadType: wasp.RPSScheduleType,
		// just use plain line profile - 5 RPS for 10s
		Schedule:   wasp.Plain(5, 10*time.Second),
		Gun:        NewExampleHTTPGun(srv.URL()),
		Labels:     labels,
		LokiConfig: wasp.NewEnvLokiConfig(),
	})
	if err != nil {
		panic(err)
	}
	// run the generator and wait until it finish
	gen.Run(true)
}
