package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestClusterScenario(t *testing.T) {
	p, err := wasp.NewClusterProfile(&wasp.ClusterConfig{
		Namespace: "wasp",
		Timeout:   10 * time.Minute,
		ChartPath: "oci://registry-1.docker.io/f4hrenh9it/wasp",
		HelmValues: map[string]string{
			"env.loki.url":       os.Getenv("LOKI_URL"),
			"env.loki.token":     os.Getenv("LOKI_TOKEN"),
			"image":              "f4hrenh9it/wasp_test:latest",
			"jobs":               "3",
			"sync":               "TestClusterScenario",
			"env.wasp.log_level": "debug",
		},
	})
	require.NoError(t, err)
	err = p.Run()
	require.NoError(t, err)
}
