package main

import (
	"testing"
	"time"

	"github.com/smartcontractkit/wasp"
)

func TestProfile(t *testing.T) {
	// start mock servers
	srv := wasp.NewHTTPMockServer(nil)
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
}
