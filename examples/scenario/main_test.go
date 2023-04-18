package main

import (
	"github.com/smartcontractkit/wasp"
	"testing"
	"time"
)

func TestScenario(t *testing.T) {
	srv := wasp.NewHTTPMockServer(nil)
	srv.Run()

	p, err := wasp.NewVUProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileVUPart{
			{
				Name: "first API",
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
	p.Run(true)
}
