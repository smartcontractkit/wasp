package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProfile(t *testing.T) {
	// start mock servers
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	srvWS := httptest.NewServer(wasp.MockWSServer{
		Sleep: 50 * time.Millisecond,
	})
	defer srvWS.Close()

	labels := map[string]string{
		"branch": "generator_healthcheck",
		"commit": "generator_healthcheck",
	}

	_, err := wasp.NewProfile().
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.RPS,
			GenName:    "first API",
			Schedule:   wasp.Plain(1, 30*time.Second),
			Gun:        NewExampleHTTPGun(srv.URL()),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.RPS,
			GenName:    "second API",
			Schedule:   wasp.Plain(2, 30*time.Second),
			Gun:        NewExampleHTTPGun(srv.URL()),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.RPS,
			GenName:    "second API",
			Schedule:   wasp.Plain(4, 30*time.Second),
			Gun:        NewExampleHTTPGun(srv.URL()),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Run(true)
	require.NoError(t, err)

	_, err = wasp.NewProfile().
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.VU,
			GenName:    "first API",
			Schedule:   wasp.Plain(1, 30*time.Second),
			VU:         NewExampleWSVirtualUser(srvWS.URL),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.VU,
			GenName:    "second API",
			Schedule:   wasp.Plain(2, 30*time.Second),
			VU:         NewExampleWSVirtualUser(srvWS.URL),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Add(wasp.NewGenerator(&wasp.Config{
			T:          t,
			LoadType:   wasp.VU,
			GenName:    "third API",
			Schedule:   wasp.Plain(4, 30*time.Second),
			VU:         NewExampleWSVirtualUser(srvWS.URL),
			Labels:     labels,
			LokiConfig: wasp.NewEnvLokiConfig(),
		})).
		Run(true)
	require.NoError(t, err)
}
