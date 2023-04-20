package wasp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	InitDefaultLogging()
	os.Exit(m.Run())
}

func TestSmokeConcurrentGenerators(t *testing.T) {
	t.Parallel()
	gens := make([]*Generator, 0)
	for i := 0; i < 2; i++ {
		gen, err := NewGenerator(&Config{
			T:        t,
			LoadType: RPSScheduleType,
			Schedule: Plain(1, 1*time.Second),
			Gun: NewMockGun(&MockGunConfig{
				CallSleep: 50 * time.Millisecond,
			}),
		})
		require.NoError(t, err)
		gen.Run(false)
		gens = append(gens, gen)
	}
	for _, gen := range gens {
		_, failed := gen.Wait()
		require.Equal(t, false, failed)
		stats := gen.Stats()
		require.Equal(t, int64(1), stats.CurrentRPS.Load())
		// we do not check exact RPS, because ratelimit.Limiter implementation
		// compensate RPS only after several requests
		// see example_test.go for precision tests of long runs
		// https://github.com/uber-go/ratelimit/blob/a12885fa6127db0aa3c29d33fc8ddeeb1fa1530c/limiter_atomic.go#L54
		require.GreaterOrEqual(t, stats.Success.Load(), int64(2))

		okData, okResponses, failResponses := convertResponsesData(gen.GetData())
		require.Contains(t, okData, "successCallData")
		require.GreaterOrEqual(t, len(okData), 2)
		require.GreaterOrEqual(t, okResponses[0].Duration, 50*time.Millisecond)
		require.Equal(t, okResponses[0].Data.(string), "successCallData")
		require.Equal(t, okResponses[1].Data.(string), "successCallData")
		require.GreaterOrEqual(t, len(okResponses), 2)
		require.Empty(t, failResponses)
		require.Empty(t, gen.Errors())
	}
}

func TestSmokePositiveOneRequest(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: RPSScheduleType,
		Schedule: Plain(1, 100*time.Millisecond),
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 50 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	gen.Run(false)
	time.Sleep(50 * time.Millisecond)
	_, failed := gen.Stop()
	require.Equal(t, false, failed)
	gs := &Stats{}
	gs.LastSegment.Store(1)
	gs.CurrentSegment.Store(0)
	gs.CurrentStep.Store(1)
	gs.CurrentVUs.Store(0)
	gs.CurrentRPS.Store(1)
	gs.Success.Add(2)
	require.Equal(t, gs, gen.Stats())

	okData, okResponses, failResponses := convertResponsesData(gen.GetData())
	require.Equal(t, []string{"successCallData", "successCallData"}, okData)
	require.GreaterOrEqual(t, okResponses[0].Duration, 50*time.Millisecond)
	require.Equal(t, okResponses[0].Data.(string), "successCallData")
	require.Equal(t, okResponses[1].Data.(string), "successCallData")
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeFailedOneRequest(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:                 t,
		LoadType:          RPSScheduleType,
		StatsPollInterval: 1 * time.Second,
		Schedule:          Plain(1, 1*time.Second),
		Gun: NewMockGun(&MockGunConfig{
			FailRatio: 100,
			CallSleep: 50 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, _ = gen.Run(false)
	time.Sleep(40 * time.Millisecond)
	_, failed := gen.Stop()
	require.Equal(t, true, failed)
	gs := &Stats{}
	gs.LastSegment.Store(1)
	gs.CurrentSegment.Store(0)
	gs.CurrentStep.Store(1)
	gs.RunFailed.Store(true)
	gs.CurrentRPS.Store(1)
	gs.CurrentVUs.Store(0)
	gs.Failed.Add(2)
	require.Equal(t, gs, gen.Stats())

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.Empty(t, okData)
	require.GreaterOrEqual(t, failResponses[0].Duration, 50*time.Millisecond)
	require.GreaterOrEqual(t, failResponses[1].Duration, 50*time.Millisecond)
	require.Equal(t, failResponses[0].Data.(string), "failedCallData")
	require.Equal(t, failResponses[0].Error, "error")
	require.Equal(t, failResponses[1].Data.(string), "failedCallData")
	require.Equal(t, failResponses[1].Error, "error")
	require.Equal(t, []string{"error", "error"}, gen.Errors())
}

func TestSmokeLoadGenCallTimeout(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:           t,
		LoadType:    RPSScheduleType,
		Schedule:    Plain(1, 1*time.Second),
		CallTimeout: 400 * time.Millisecond,
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 500 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	gen.Run(false)
	time.Sleep(990 * time.Millisecond)
	_, failed := gen.Stop()
	require.Equal(t, true, failed)
	gs := &Stats{}
	gs.LastSegment.Store(1)
	gs.CurrentSegment.Store(0)
	gs.CurrentStep.Store(1)
	gs.CurrentRPS.Store(1)
	gs.CurrentVUs.Store(0)
	gs.RunFailed.Store(true)
	gs.CallTimeout.Add(2)
	require.Equal(t, gs, gen.Stats())

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.Empty(t, okData)
	require.Equal(t, failResponses[0].Data, nil)
	require.Equal(t, failResponses[0].Error, "generator request call timeout")
	require.Equal(t, []string{ErrCallTimeout.Error(), ErrCallTimeout.Error()}, gen.Errors())
}

func TestSmokeLoadGenCallTimeoutWait(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:           t,
		LoadType:    RPSScheduleType,
		Schedule:    Plain(1, 1*time.Second),
		CallTimeout: 50 * time.Millisecond,
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 55 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, true, failed)
	stats := gen.Stats()
	require.Equal(t, int64(1), stats.CurrentRPS.Load())
	require.GreaterOrEqual(t, stats.CallTimeout.Load(), int64(2))
	require.Equal(t, true, stats.RunFailed.Load())

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.Empty(t, okData)
	require.Equal(t, failResponses[0].Data, nil)
	require.Equal(t, failResponses[0].Error, "generator request call timeout")
	require.Contains(t, gen.Errors(), ErrCallTimeout.Error())
	require.GreaterOrEqual(t, len(gen.Errors()), 2)
}

func TestSmokeCancelledByDeadlineWait(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule:          Plain(1, 400*time.Millisecond),
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 500 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	gen.Run(false)
	before := time.Now()
	_, failed := gen.Wait()
	after := time.Now()
	elapsed := after.Sub(before)
	// execution time + last request
	require.Greater(t, elapsed, 1050*time.Millisecond)
	require.Equal(t, false, failed)
	gs := &Stats{}
	gs.LastSegment.Store(1)
	gs.CurrentSegment.Store(0)
	gs.CurrentStep.Store(1)
	gs.CurrentVUs.Store(0)
	gs.CurrentRPS.Store(1)
	gs.Success.Add(2)
	require.Equal(t, gs, gen.Stats())

	// in case of gen.Stop() if we don't have test duration or if gen.Wait() and we have a deadline
	// we are waiting for all requests, so result in that case must be successful
	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.Equal(t, []string{"successCallData", "successCallData"}, okData)
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeCancelledBeforeDeadline(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: RPSScheduleType,
		Schedule: Plain(1, 40*time.Millisecond),
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 50 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	gen.Run(false)
	before := time.Now()
	time.Sleep(20 * time.Millisecond)
	_, failed := gen.Stop()
	after := time.Now()
	elapsed := after.Sub(before)

	require.Greater(t, elapsed, 1050*time.Millisecond)
	require.Equal(t, false, failed)
	gs := &Stats{}
	gs.LastSegment.Store(1)
	gs.CurrentSegment.Store(0)
	gs.CurrentStep.Store(1)
	gs.CurrentVUs.Store(0)
	gs.CurrentRPS.Store(1)
	gs.Success.Add(2)
	require.Equal(t, gs, gen.Stats())

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.Equal(t, []string{"successCallData", "successCallData"}, okData)
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeStaticRPSSchedulePrecision(t *testing.T) {
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: RPSScheduleType,
		Schedule: Plain(1000, 1*time.Second),
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 50 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, false, failed)
	require.GreaterOrEqual(t, gen.Stats().Success.Load(), int64(995))
	require.LessOrEqual(t, gen.Stats().Success.Load(), int64(1009))
	require.Equal(t, gen.Stats().Failed.Load(), int64(0))
	require.Equal(t, gen.Stats().CallTimeout.Load(), int64(0))

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.GreaterOrEqual(t, len(okData), 995)
	require.LessOrEqual(t, len(okData), 1009)
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeStaticRPSScheduleIsNotBlocking(t *testing.T) {
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: RPSScheduleType,
		Schedule: Plain(1000, 1*time.Second),
		Gun: NewMockGun(&MockGunConfig{
			// call time must not affect the load scheduleSegments
			CallSleep: 1 * time.Second,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, false, failed)
	require.GreaterOrEqual(t, gen.Stats().Success.Load(), int64(998))
	require.LessOrEqual(t, gen.Stats().Success.Load(), int64(1009))
	require.Equal(t, gen.Stats().Failed.Load(), int64(0))
	require.Equal(t, gen.Stats().CallTimeout.Load(), int64(0))

	okData, _, failResponses := convertResponsesData(gen.GetData())
	require.GreaterOrEqual(t, len(okData), 998)
	require.LessOrEqual(t, len(okData), 1009)
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeLoadScheduleSegmentRPSIncrease(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule: []*Segment{
			{
				From:         1,
				Increase:     1,
				Steps:        7,
				StepDuration: 1 * time.Second,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, false, failed)
	require.GreaterOrEqual(t, gen.Stats().Success.Load(), int64(28))
}

func TestSmokeLoadScheduleSegmentRPSDecrease(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule: []*Segment{
			{
				From:         5,
				Increase:     -1,
				Steps:        5,
				StepDuration: 1 * time.Second,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, false, failed)
	require.GreaterOrEqual(t, gen.Stats().Success.Load(), int64(20))
}

func TestSmokeValidation(t *testing.T) {
	t.Parallel()
	_, err := NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule: []*Segment{
			{
				From:         0,
				Increase:     1,
				Steps:        1,
				StepDuration: 1 * time.Second,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.Equal(t, ErrStartFrom, err)
	_, err = NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule: []*Segment{
			{
				From:         1,
				Increase:     1,
				StepDuration: 1 * time.Second,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.Equal(t, ErrInvalidSteps, err)
	_, err = NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          RPSScheduleType,
		Schedule: []*Segment{
			{
				From:     1,
				Increase: 1,
				Steps:    1,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.Equal(t, ErrInvalidSteps, err)
	_, err = NewGenerator(nil)
	require.Equal(t, ErrNoCfg, err)
	_, err = NewGenerator(&Config{
		T:        t,
		LoadType: RPSScheduleType,
		Schedule: []*Segment{
			{
				From: 1,
			},
		},
		Gun: nil,
	})
	require.Equal(t, ErrNoImpl, err)
	_, err = NewGenerator(&Config{
		T:        t,
		LoadType: "arbitrary_load_type",
		Schedule: []*Segment{
			{
				From: 1,
			},
		},
		Gun: NewMockGun(&MockGunConfig{
			CallSleep: 10 * time.Millisecond,
		}),
	})
	require.Equal(t, ErrWrongScheduleType, err)
}

func TestSmokeVUsIncrease(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: VUScheduleType,
		Schedule: []*Segment{
			{
				From:         1,
				Increase:     1,
				Steps:        10,
				StepDuration: 100 * time.Millisecond,
			},
		},
		VU: NewMockVU(MockVirtualUserConfig{
			CallSleep: 100 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	stats := gen.Stats()
	require.Equal(t, false, failed)
	require.Equal(t, int64(11), stats.CurrentVUs.Load())

	okData, okResponses, failResponses := convertResponsesData(gen.GetData())
	require.GreaterOrEqual(t, okResponses[0].Duration, 50*time.Millisecond)
	require.Greater(t, len(okResponses), 50)
	require.Greater(t, len(okData), 50)
	require.Equal(t, okResponses[0].Data.(string), "successCallData")
	require.Equal(t, okResponses[50].Data.(string), "successCallData")
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeVUsDecrease(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:        t,
		LoadType: VUScheduleType,
		Schedule: []*Segment{
			{
				From:         10,
				Increase:     -1,
				Steps:        10,
				StepDuration: 1 * time.Second,
			},
		},
		VU: NewMockVU(MockVirtualUserConfig{
			CallSleep: 50 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	stats := gen.Stats()
	require.Equal(t, false, failed)
	require.Equal(t, int64(1), stats.CurrentVUs.Load())

	okData, okResponses, failResponses := convertResponsesData(gen.GetData())
	require.GreaterOrEqual(t, okResponses[0].Duration, 50*time.Millisecond)
	require.Greater(t, len(okResponses), 50)
	require.Greater(t, len(okData), 50)
	require.Equal(t, okResponses[0].Data.(string), "successCallData")
	require.Equal(t, okResponses[50].Data.(string), "successCallData")
	require.Empty(t, failResponses)
	require.Empty(t, gen.Errors())
}

func TestSmokeVUsSetupTeardown(t *testing.T) {
	t.Parallel()
	gen, err := NewGenerator(&Config{
		T:                 t,
		StatsPollInterval: 1 * time.Second,
		LoadType:          VUScheduleType,
		Schedule: Combine(
			Line(1, 10, 10*time.Second),
			Line(10, 1, 10*time.Second),
		),
		VU: NewMockVU(MockVirtualUserConfig{
			CallSleep: 100 * time.Millisecond,
		}),
	})
	require.NoError(t, err)
	_, failed := gen.Run(true)
	require.Equal(t, false, failed)
	require.GreaterOrEqual(t, gen.Stats().Success.Load(), int64(1099))
}
