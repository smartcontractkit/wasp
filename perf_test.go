package wasp

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"context"
	//nolint
	_ "net/http/pprof"
	"runtime"

	"github.com/pyroscope-io/client/pyroscope"
)

/* This tests can also be used as a performance validation of a tool itself or as a dashboard data filler */

func stdPyro(t *testing.T) {
	t.Helper()
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "wasp",
		ServerAddress:   "http://localhost:4040",
		Logger:          pyroscope.StandardLogger,
		Tags:            map[string]string{"test": "wasp-trace-1"},

		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	require.NoError(t, err)
}

func TestPyroscopeLocalTrace(t *testing.T) {
	// run like
	// go test -run TestLocalTrace -trace trace.out
	// to have all in one, then
	// go tool trace trace.out
	stdPyro(t)
	t.Parallel()
	t.Run("trace test", func(t *testing.T) {
		t.Parallel()
		pyroscope.TagWrapper(context.Background(), pyroscope.Labels("scope", "loadgen_impl"), func(c context.Context) {
			gen, err := NewGenerator(&Config{
				T:          t,
				LokiConfig: NewEnvLokiConfig(),
				Labels: map[string]string{
					"cluster":    "sdlc",
					"namespace":  "load-dummy-test",
					"app":        "dummy",
					"test_group": "generator_healthcheck",
					"test_id":    "dummy-healthcheck-pyro-1",
				},
				CallTimeout: 100 * time.Millisecond,
				LoadType:    RPSScheduleType,
				Schedule:    Plain(100, 10*time.Second),
				Gun: NewMockGun(&MockGunConfig{
					CallSleep: 50 * time.Millisecond,
				}),
			})
			require.NoError(t, err)
			//nolint
			gen.Run(true)
		})
	})
}

func TestRenderLokiRPSRun(t *testing.T) {
	t.Parallel()
	t.Run("can_report_to_loki", func(t *testing.T) {
		t.Parallel()
		gen, err := NewGenerator(&Config{
			T:          t,
			LokiConfig: NewEnvLokiConfig(),
			Labels: map[string]string{
				"branch":   "generator_healthcheck",
				"commit":   "generator_healthcheck",
				"gen_name": "rps",
			},
			CallTimeout: 100 * time.Millisecond,
			LoadType:    RPSScheduleType,
			Schedule: CombineAndRepeat(
				2,
				Line(1, 100, 30*time.Second),
				Plain(200, 30*time.Second),
				Line(100, 1, 30*time.Second),
			),
			Gun: NewMockGun(&MockGunConfig{
				TimeoutRatio: 1,
				CallSleep:    50 * time.Millisecond,
			}),
		})
		require.NoError(t, err)
		gen.Run(true)
	})
}

func TestRenderLokiVUsRun(t *testing.T) {
	t.Parallel()
	t.Run("can_report_to_loki", func(t *testing.T) {
		t.Parallel()
		gen, err := NewGenerator(&Config{
			T:          t,
			LokiConfig: NewEnvLokiConfig(),
			Labels: map[string]string{
				"branch":   "generator_healthcheck",
				"commit":   "generator_healthcheck",
				"gen_name": "vu",
			},
			CallTimeout: 100 * time.Millisecond,
			LoadType:    VUScheduleType,
			Schedule: CombineAndRepeat(
				2,
				Line(1, 20, 30*time.Second),
				Plain(30, 30*time.Second),
				Line(20, 1, 30*time.Second),
			),
			VU: NewMockVU(MockVirtualUserConfig{
				CallSleep: 100 * time.Millisecond,
			}),
		})
		require.NoError(t, err)
		gen.Run(true)
	})
}

func TestRenderLokiSpikeMaxLoadRun(t *testing.T) {
	t.Skip("This test is for manual run with or without Loki to measure max RPS")
	t.Parallel()
	t.Run("max_spike", func(t *testing.T) {
		t.Parallel()
		gen, err := NewGenerator(&Config{
			T:          t,
			LokiConfig: NewEnvLokiConfig(),
			Labels: map[string]string{
				"branch":   "generator_healthcheck",
				"commit":   "generator_healthcheck",
				"gen_name": "spike",
			},
			CallTimeout: 100 * time.Millisecond,
			LoadType:    RPSScheduleType,
			Schedule:    Plain(5000, 20*time.Second),
			Gun: NewMockGun(&MockGunConfig{
				CallSleep: 50 * time.Millisecond,
			}),
		})
		require.NoError(t, err)
		gen.Run(true)
	})
}

func TestRenderWS(t *testing.T) {
	t.Skip("This test is for manual run to measure max WS messages/s")
	s := httptest.NewServer(MockWSServer{
		Sleep: 50 * time.Millisecond,
		Logf:  t.Logf,
	})
	defer s.Close()

	gen, err := NewGenerator(&Config{
		T:          t,
		LokiConfig: NewEnvLokiConfig(),
		Labels: map[string]string{
			"branch":   "generator_healthcheck",
			"commit":   "generator_healthcheck",
			"gen_name": "ws",
		},
		LoadType: VUScheduleType,
		Schedule: []*Segment{
			{
				From:         10,
				Increase:     20,
				Steps:        10,
				StepDuration: 10 * time.Second,
			},
		},
		VU: NewWSMockVU(WSMockVUConfig{TargetURl: s.URL}),
	})
	require.NoError(t, err)
	gen.Run(true)
}

func TestRenderHTTP(t *testing.T) {
	t.Skip("This test is for manual run to measure max HTTP RPS")
	srv := NewHTTPMockServer(nil)
	srv.Run()

	gen, err := NewGenerator(&Config{
		T:          t,
		LokiConfig: NewEnvLokiConfig(),
		Labels: map[string]string{
			"branch":   "generator_healthcheck",
			"commit":   "generator_healthcheck",
			"gen_name": "http",
		},
		LoadType: RPSScheduleType,
		Schedule: Line(10, 400, 500*time.Second),
		Gun:      NewHTTPMockGun(&MockHTTPGunConfig{TargetURL: "http://localhost:8080"}),
	})
	require.NoError(t, err)
	gen.Run(true)
}
