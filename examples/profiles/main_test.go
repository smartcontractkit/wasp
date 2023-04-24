package main

import (
	"github.com/smartcontractkit/wasp"
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

	p, err := wasp.NewProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileVUPart{
			{
				Name:     "first API",
				VU:       NewExampleWSVirtualUser(srvWS.URL),
				Schedule: wasp.Plain(1, 30*time.Second),
			},
			{
				Name:     "second API",
				VU:       NewExampleWSVirtualUser(srvWS.URL),
				Schedule: wasp.Plain(2, 30*time.Second),
			},
			{
				Name:     "third API",
				VU:       NewExampleWSVirtualUser(srvWS.URL),
				Schedule: wasp.Plain(4, 30*time.Second),
			},
		})
	if err != nil {
		panic(err)
	}
	_ = p.Run(true)

	p, err = wasp.NewProfile(
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
	_ = p.Run(true)
}
