package main

import (
	"os"
	"testing"
	"time"

	"github.com/K-Phoen/grabana/alert"
	"github.com/smartcontractkit/wasp"
	"github.com/stretchr/testify/require"
)

const (
	FirstGenName                 = "first API"
	SecondGenName                = "second API"
	BaselineRequirementGroupName = "baseline"
	StressRequirementGroupName   = "stress"
)

func TestMain(m *testing.M) {
	srv := wasp.NewHTTPMockServer(
		&wasp.HTTPMockServerConfig{
			FirstAPILatency:   50 * time.Millisecond,
			FirstAPIHTTPCode:  500,
			SecondAPILatency:  50 * time.Millisecond,
			SecondAPIHTTPCode: 500,
		},
	)
	srv.Run()

	// we define 2 NFRs groups
	// - baseline - basic latency requirements for 99th percentiles and no errors
	// - stress - another custom NFRs for stress
	// CustomWaspAlert can be defined on per Generator level
	// usually, you define it once per project, generate your dashboard and upload it, it's here only for example purposes
	_, err := wasp.NewDashboard().Deploy(
		[]wasp.CustomWaspAlert{
			{
				Name:                 "99th latency percentile is out of SLO for first API",
				AlertType:            wasp.AlertTypeQuantile99,
				TestName:             "TestBaselineRequirements",
				GenName:              FirstGenName,
				RequirementGroupName: BaselineRequirementGroupName,
				AlertIf:              alert.IsAbove(50),
			},
			{
				Name:                 "first API has errors",
				AlertType:            wasp.AlertTypeErrors,
				TestName:             "TestBaselineRequirements",
				GenName:              FirstGenName,
				RequirementGroupName: BaselineRequirementGroupName,
				AlertIf:              alert.IsAbove(0),
			},
			{
				Name:                 "first API has errors > threshold",
				AlertType:            wasp.AlertTypeErrors,
				TestName:             "TestStressRequirements",
				GenName:              FirstGenName,
				RequirementGroupName: StressRequirementGroupName,
				AlertIf:              alert.IsAbove(10),
			},
			{
				Name:                 "99th latency percentile is out of SLO for second API",
				AlertType:            wasp.AlertTypeQuantile99,
				TestName:             "TestBaselineRequirements",
				GenName:              FirstGenName,
				RequirementGroupName: BaselineRequirementGroupName,
				AlertIf:              alert.IsAbove(50),
			},
			{
				Name:                 "second API has errors",
				AlertType:            wasp.AlertTypeErrors,
				TestName:             "TestBaselineRequirements",
				GenName:              FirstGenName,
				RequirementGroupName: BaselineRequirementGroupName,
				AlertIf:              alert.IsAbove(0),
			},
		},
	)
	if err != nil {
		panic(err)
	}
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestBaselineRequirements(t *testing.T) {
	p, err := wasp.NewRPSProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileGunPart{
			{
				Name:     FirstGenName,
				Gun:      NewExampleHTTPGun("http://localhost:8080/1"),
				Schedule: wasp.Plain(5, 20*time.Second),
			},
			{
				Name:     SecondGenName,
				Gun:      NewExampleHTTPGun("http://localhost:8080/2"),
				Schedule: wasp.Plain(10, 20*time.Second),
			},
		})
	if err != nil {
		panic(err)
	}
	p.Run(true)

	// we are checking all active alerts for dashboard with UUID = "wasp" which have label "requirement_name" = "baseline"
	// if any alerts of particular group, for example "baseline" were raised - we fail the test
	// change some data in NewHTTPMockServer to make alerts disappear
	_, err = wasp.NewAlertChecker(t).AnyAlerts(wasp.DefaultDashboardUUID, BaselineRequirementGroupName)
	require.NoError(t, err)
}

func TestStressRequirements(t *testing.T) {
	// we are testing the same APIs but for different NFRs group
	p, err := wasp.NewRPSProfile(
		t,
		map[string]string{
			"branch": "generator_healthcheck",
			"commit": "generator_healthcheck",
		}, []*wasp.ProfileGunPart{
			{
				Name:     FirstGenName,
				Gun:      NewExampleHTTPGun("http://localhost:8080/1"),
				Schedule: wasp.Plain(10, 20*time.Second),
			},
			{
				Name:     SecondGenName,
				Gun:      NewExampleHTTPGun("http://localhost:8080/2"),
				Schedule: wasp.Plain(20, 20*time.Second),
			},
		})
	if err != nil {
		panic(err)
	}
	p.Run(true)

	// we are checking all active alerts for dashboard with UUID = "wasp" which have label "requirement_name" = "stress"
	// if any alerts of particular group, for example "stress" were raised - we fail the test
	// change some data in NewHTTPMockServer to make alerts disappear
	_, err = wasp.NewAlertChecker(t).AnyAlerts(wasp.DefaultDashboardUUID, StressRequirementGroupName)
	require.NoError(t, err)
}
