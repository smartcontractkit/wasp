package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestClusterScenario(t *testing.T) {
	// create a new namespace before that
	// kubectl create ns wasp
	p, err := wasp.NewClusterProfile(&wasp.ClusterConfig{
		ChartPath: "../../charts/wasp/",
		Namespace: "wasp",
		Timeout:   10 * time.Minute,
		HelmValues: map[string]string{
			"env.loki.url":       os.Getenv("LOKI_URL"),
			"env.loki.token":     os.Getenv("LOKI_TOKEN"),
			"image":              "f4hrenh9it/wasp_test:latest",
			"jobs":               "9",
			"sync":               "TestClusterScenario",
			"env.wasp.log_level": "debug",
		},
	})
	require.NoError(t, err)
	err = p.Run()
	require.NoError(t, err)
}
