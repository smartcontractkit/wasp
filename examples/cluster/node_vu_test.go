package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeVU(t *testing.T) {
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	p, err := wasp.NewProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileVUPart{
			{
				Name: "Gamma",
				VU:   NewExampleScenario(srv.URL()),
				Schedule: wasp.Combine(
					wasp.Line(1, 20, 1*time.Minute),
				),
			},
			{
				Name: "Delta",
				VU:   NewExampleScenario(srv.URL()),
				Schedule: wasp.Combine(
					wasp.Line(1, 40, 1*time.Minute),
				),
			},
		})
	if err != nil {
		panic(err)
	}
	err = p.Run(true)
	require.NoError(t, err)
}
