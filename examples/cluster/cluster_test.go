package main

import (
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestClusterScenario(t *testing.T) {
	// modify CPU/MEM guards, test will end if threshold was reached
	// these are the defaults, so you can omit them
	wasp.ResourcesThresholdCheckInterval = 5 * time.Second
	wasp.CPUIdleThresholdPercentage = 20
	wasp.MEMFreeThresholdPercentage = 1

	p, err := wasp.NewClusterProfile(&wasp.ClusterConfig{
		Namespace: "wasp",
		Timeout:   5 * time.Minute,
		KeepJobs:  true,
		ChartPath: "../../charts/wasp",
		HelmValues: map[string]string{
			"env.loki.url":              os.Getenv("LOKI_URL"),
			"env.loki.token":            os.Getenv("LOKI_TOKEN"),
			"env.loki.basic_auth":       os.Getenv("LOKI_BASIC_AUTH"),
			"env.loki.tenant_id":        os.Getenv("LOKI_TENANT_ID"),
			"env.wasp.log_level":        "debug",
			"image":                     "323150190480.dkr.ecr.us-west-2.amazonaws.com/wasp-tests:self-test-amd64-multi-2",
			"jobs":                      "1",
			"resources.requests.cpu":    "1000m",
			"resources.requests.memory": "512Mi",
			"resources.limits.cpu":      "1000m",
			"resources.limits.memory":   "512Mi",
			"test.binaryName":           "cluster.test",
			"test.name":                 "TestNodeRPS",
			"test.timeout":              "24h",
			// other test vars can set like
			"test.MY_CUSTOM_VAR": "abc",
		},
	})
	require.NoError(t, err)
	err = p.Run()
	require.NoError(t, err)
}
