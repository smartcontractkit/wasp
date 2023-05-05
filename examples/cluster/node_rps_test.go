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

	p, err := wasp.NewProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileGunPart{
			{
				Name: "Alpha",
				Gun:  NewExampleHTTPGun("http://localhost:8080/1"),
				Schedule: wasp.Combine(
					wasp.Line(10, 20, 100*time.Second),
				),
			},
			{
				Name: "Beta",
				Gun:  NewExampleHTTPGun("http://localhost:8080/2"),
				Schedule: wasp.Combine(
					wasp.Line(10, 40, 100*time.Second),
				),
			},
		})
	if err != nil {
		panic(err)
	}
	err = p.Run(true)
	require.NoError(t, err)
}
