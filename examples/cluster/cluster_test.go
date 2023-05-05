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
		Timeout:   5 * time.Minute,
		// default chart "oci://registry-1.docker.io/f4hrenh9it/wasp:v0.1.7"
		ChartPath: "../../charts/wasp",
		//KeepJobs:  true,
		HelmValues: map[string]string{
			"env.loki.url":              os.Getenv("LOKI_URL"),
			"env.loki.token":            os.Getenv("LOKI_TOKEN"),
			"test.name":                 "TestNodeRPS",
			"test.timeout":              "24h",
			"image":                     "f4hrenh9it/wasp_test:latest",
			"jobs":                      "10",
			"resources.requests.cpu":    "1000m",
			"resources.requests.memory": "512Mi",
			"resources.limits.cpu":      "1000m",
			"resources.limits.memory":   "512Mi",
			"env.wasp.log_level":        "debug",
		},
	})
	require.NoError(t, err)
	err = p.Run()
	require.NoError(t, err)
}
