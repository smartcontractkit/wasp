package main

import (
	"testing"
	"time"

	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
)

func TestProfile(t *testing.T) {
	// start mock servers
	srv := wasp.NewHTTPMockServer(50 * time.Millisecond)
	srv.Run()

	p, err := wasp.NewRPSProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileGunPart{
			{
				Name:     "first API",
				Gun:      NewExampleHTTPGun(srv.URL()),
				Schedule: wasp.Plain(5, 30*time.Second),
			},
			{
				Name:     "second API",
				Gun:      NewExampleHTTPGun(srv.URL()),
				Schedule: wasp.Plain(10, 30*time.Second),
			},
			{
				Name:     "third API",
				Gun:      NewExampleHTTPGun(srv.URL()),
				Schedule: wasp.Plain(20, 30*time.Second),
			},
		})
	if err != nil {
		panic(err)
	}
	p.Run(true)

	// we are checking all active alerts for dashboard with UUID = "wasp" which have label "requirement_name" = "baseline"
	// if any alerts were raised we fail the test
	// lower the latency in NewHTTPMockServer to 30ms make alert disappear
	err = wasp.NewAlertChecker(t, "requirement_name").AnyAlerts("wasp", "baseline")
	require.NoError(t, err)
}
