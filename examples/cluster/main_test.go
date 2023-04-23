package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestClusterScenario(t *testing.T) {
	// create a new namespace before that
	// kubectl create ns wasp
	p := wasp.NewClusterProfile(&wasp.ClusterConfig{
		ChartPath: "../../charts/wasp/",
		Namespace: "wasp",
		HelmValues: map[string]string{
			"image":              "f4hrenh9it/wasp_test:latest",
			"jobs":               "1",
			"sync_label":         "Test",
			"env.vars.log_level": "trace",
			"env.loki.url":       os.Getenv("LOKI_URL"),
			"env.loki.token":     os.Getenv("LOKI_TOKEN"),
		},
	})

	err := p.Run()
	require.NoError(t, err)
}
