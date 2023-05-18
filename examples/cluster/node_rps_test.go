package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNodeRPS(t *testing.T) {
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	labels := map[string]string{
		"branch": "generator_healthcheck",
		"commit": "generator_healthcheck",
	}

	_, err := wasp.NewProfile().
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.RPS,
			GenName:    "Alpha",
			Schedule:   wasp.Line(1, 20, 30*time.Second),
			Gun:        NewExampleHTTPGun(srv.URL()),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.RPS,
			GenName:    "Beta",
			Schedule:   wasp.Line(1, 40, 30*time.Second),
			Gun:        NewExampleHTTPGun(srv.URL()),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Run(true)
	require.NoError(t, err)
}
