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
		HelmValues: map[string]string{
			"env.loki.url":              os.Getenv("LOKI_URL"),
			"env.loki.token":            os.Getenv("LOKI_TOKEN"),
			"env.loki.basic_auth":       os.Getenv("LOKI_BASIC_AUTH"),
			"env.loki.tenant_id":        os.Getenv("LOKI_TENANT_ID"),
			"test.name":                 "TestNodeRPS",
			"test.timeout":              "24h",
			"image":                     "public.ecr.aws/chainlink/wasp-test:latest",
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
