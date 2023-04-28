package main

import (
	"net/http/httptest"
	"time"

	"github.com/smartcontractkit/wasp"
)

func main() {
	// start mock http server
	s := httptest.NewServer(wasp.MockWSServer{
		Sleep: 50 * time.Millisecond,
	})
	defer s.Close()

	// define labels for differentiate one run from another
	labels := map[string]string{
		// check variables in dashboard/dashboard.go
		"go_test_name": "generator_healthcheck",
		"gen_name":     "generator_healthcheck",
		"branch":       "generator_healthcheck",
		"commit":       "generator_healthcheck",
	}

	// create generator
	gen, err := wasp.NewGenerator(&wasp.Config{
		LoadType: wasp.VU,
		// just use plain line profile - 5 VUs for 10s
		Schedule:   wasp.Plain(5, 10*time.Second),
		VU:         NewExampleWSVirtualUser(s.URL),
		Labels:     labels,
		LokiConfig: wasp.NewEnvLokiConfig(),
	})
	if err != nil {
		panic(err)
	}
	// run the generator and wait until it finish
	gen.Run(true)
}
