package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNode(t *testing.T) {
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	p, err := wasp.NewProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileVUPart{
			{
				Name: "Two sequential calls scenario",
				VU:   NewExampleScenario(srv.URL()),
				Schedule: wasp.Combine(
					wasp.Plain(5, 30*time.Second),
					wasp.Plain(10, 30*time.Second),
				),
			},
		})
	if err != nil {
		panic(err)
	}
	err = p.Run(true)
	require.NoError(t, err)
}
