package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestScenario(t *testing.T) {
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	_, err := wasp.NewProfile().
		Add(wasp.NewGenerator(&wasp.Config{
			T: t,
			Labels: map[string]string{
				"branch": "generator_healthcheck",
				"commit": "generator_healthcheck",
			},
			LoadType: wasp.VU,
			VU:       NewExampleScenario(srv.URL()),
			Schedule: wasp.Combine(
				wasp.Plain(5, 30*time.Second),
				wasp.Plain(10, 30*time.Second),
			),
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).Run(true)
	require.NoError(t, err)
}
