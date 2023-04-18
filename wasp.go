package wasp

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
)

const (
	DefaultCallTimeout       = 1 * time.Minute
	DefaultStatsPollInterval = 5 * time.Second
	DefaultCallResultBufLen  = 50000
	DefaultGenName           = "Generator"
)

var (
	ErrNoCfg             = errors.New("config is nil")
	ErrNoImpl            = errors.New("either \"gun\" or \"vu\" implementation must provided")
	ErrNoSchedule        = errors.New("no schedule segments were provided")
	ErrWrongScheduleType = errors.New("schedule type must be RPSScheduleType or VUScheduleType, use package constants")
	ErrCallTimeout       = errors.New("generator request call timeout")
	ErrStartFrom         = errors.New("from must be > 0")
	ErrInvalidSteps      = errors.New("both \"Steps\" and \"StepsDuration\" must be defined in a schedule segment")
	ErrNoGun             = errors.New("rps load scheduleSegments selected but gun implementation is nil")
	ErrNoVU              = errors.New("vu load scheduleSegments selected but vu implementation is nil")
)

// Gun is basic interface to some synthetic load test
// Call one request with some RPS schedule
type Gun interface {
	Call(l *Generator) CallResult
}

// VirtualUser is basic interface to run virtual users load
// you should use it if:
// - your protocol is stateful, ex.: ws, grpc
// - you'd like to have some VirtualUser modelling
type VirtualUser interface {
	Call(l *Generator)
	Stop(l *Generator)
	Clone(l *Generator) VirtualUser
	Setup(l *Generator) error
	Teardown(l *Generator) error
	StopChan() chan struct{}
}

// CallResult represents basic call result info
type CallResult struct {
	Failed     bool          `json:"failed,omitempty"`
	Timeout    bool          `json:"timeout,omitempty"`
	Duration   time.Duration `json:"duration"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	Data       interface{}   `json:"data,omitempty"`
	Error      string        `json:"error,omitempty"`
}

const (
	RPSScheduleType string = "rps_schedule"
	VUScheduleType  string = "vu_schedule"
)

// Segment load test schedule segment
type Segment struct {
	From         int64
	Increase     int64
	Steps        int64
	StepDuration time.Duration
	rl           ratelimit.Limiter
}

func (ls *Segment) Validate(cfg *Config) error {
	if ls.From <= 0 {
		return ErrStartFrom
	}
	if ls.Steps < 0 || (ls.Steps != 0 && ls.StepDuration == 0) || (ls.StepDuration != 0 && ls.Steps == 0) {
		return ErrInvalidSteps
	}
	return nil
}

// Config is for shared load test data and configuration
type Config struct {
	T                 *testing.T
	GenName           string
	LoadType          string
	Labels            map[string]string
	LokiConfig        *LokiConfig
	Schedule          []*Segment
	CallResultBufLen  int
	StatsPollInterval time.Duration
	CallTimeout       time.Duration
	Gun               Gun
	VU                VirtualUser
	Logger            zerolog.Logger
	SharedData        interface{}
	// calculated fields
	duration time.Duration
}

func (lgc *Config) Validate() error {
	if lgc.CallTimeout == 0 {
		lgc.CallTimeout = DefaultCallTimeout
	}
	if lgc.StatsPollInterval == 0 {
		lgc.StatsPollInterval = DefaultStatsPollInterval
	}
	if lgc.CallResultBufLen == 0 {
		lgc.CallResultBufLen = DefaultCallResultBufLen
	}
	if lgc.GenName == "" {
		lgc.GenName = DefaultGenName
	}
	if lgc.Gun == nil && lgc.VU == nil {
		return ErrNoImpl
	}
	if lgc.Schedule == nil {
		return ErrNoSchedule
	}
	if lgc.LoadType != RPSScheduleType && lgc.LoadType != VUScheduleType {
		return ErrWrongScheduleType
	}
	if lgc.LoadType == RPSScheduleType && lgc.Gun == nil {
		return ErrNoGun
	}
	if lgc.LoadType == VUScheduleType && lgc.VU == nil {
		return ErrNoVU
	}
	return nil
}

// Stats basic generator load stats
type Stats struct {
	CurrentRPS     atomic.Int64 `json:"currentRPS"`
	CurrentVUs     atomic.Int64 `json:"currentVUs"`
	LastSegment    atomic.Int64 `json:"last_segment"`
	CurrentSegment atomic.Int64 `json:"current_schedule_segment"`
	CurrentStep    atomic.Int64 `json:"current_schedule_step"`
	RunStopped     atomic.Bool  `json:"runStopped"`
	RunFailed      atomic.Bool  `json:"runFailed"`
	Success        atomic.Int64 `json:"success"`
	Failed         atomic.Int64 `json:"failed"`
	CallTimeout    atomic.Int64 `json:"callTimeout"`
}

// ResponseData includes any request/response data that a gun might store
// ok* slices usually contains successful responses and their verifications if their done async
// fail* slices contains CallResult with response data and an error
type ResponseData struct {
	okDataMu        *sync.Mutex
	OKData          *SliceBuffer[any]
	okResponsesMu   *sync.Mutex
	OKResponses     *SliceBuffer[CallResult]
	failResponsesMu *sync.Mutex
	FailResponses   *SliceBuffer[CallResult]
}

// Generator generates load with some RPS
type Generator struct {
	cfg                *Config
	Log                zerolog.Logger
	labels             model.LabelSet
	scheduleSegments   []*Segment
	currentSegment     *Segment
	ResponsesWaitGroup *sync.WaitGroup
	dataWaitGroup      *sync.WaitGroup
	ResponsesCtx       context.Context
	responsesCancel    context.CancelFunc
	dataCtx            context.Context
	dataCancel         context.CancelFunc
	gun                Gun
	vu                 VirtualUser
	vus                []VirtualUser
	ResponsesChan      chan CallResult
	responsesData      *ResponseData
	errsMu             *sync.Mutex
	errs               *SliceBuffer[string]
	stats              *Stats
	loki               ExtendedLokiClient
	lokiResponsesChan  chan CallResult
}

// NewGenerator creates a new generator,
// shoots for scheduled RPS until timeout, test logic is defined through Gun or VirtualUser
func NewGenerator(cfg *Config) (*Generator, error) {
	InitDefaultLogging()
	if cfg == nil {
		return nil, ErrNoCfg
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	for _, s := range cfg.Schedule {
		if err := s.Validate(cfg); err != nil {
			return nil, err
		}
	}
	for _, s := range cfg.Schedule {
		segmentTotal := time.Duration(s.Steps) * s.StepDuration
		cfg.duration += segmentTotal
	}
	l := GetLogger(cfg.T, cfg.GenName)

	var loki ExtendedLokiClient
	var err error
	if cfg.LokiConfig != nil {
		if cfg.LokiConfig.URL == "" {
			l.Warn().Msg("Loki config is set but URL is empty, saving results in memory!")
			loki = NewMockPromtailClient()
			if err != nil {
				return nil, err
			}
		} else {
			loki, err = NewLokiClient(cfg.LokiConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	ls := LabelsMapToModel(cfg.Labels)
	if cfg.T != nil {
		ls = ls.Merge(model.LabelSet{
			"go_test_name": model.LabelValue(cfg.T.Name()),
		})
	}
	// context for all requests/responses and vus
	responsesCtx, responsesCancel := context.WithTimeout(context.Background(), cfg.duration)
	// context for all the collected data
	dataCtx, dataCancel := context.WithCancel(context.Background())
	return &Generator{
		cfg:                cfg,
		scheduleSegments:   cfg.Schedule,
		ResponsesWaitGroup: &sync.WaitGroup{},
		dataWaitGroup:      &sync.WaitGroup{},
		ResponsesCtx:       responsesCtx,
		responsesCancel:    responsesCancel,
		dataCtx:            dataCtx,
		dataCancel:         dataCancel,
		gun:                cfg.Gun,
		vu:                 cfg.VU,
		ResponsesChan:      make(chan CallResult),
		labels:             ls,
		responsesData: &ResponseData{
			okDataMu:        &sync.Mutex{},
			OKData:          NewSliceBuffer[any](DefaultCallResultBufLen),
			okResponsesMu:   &sync.Mutex{},
			OKResponses:     NewSliceBuffer[CallResult](DefaultCallResultBufLen),
			failResponsesMu: &sync.Mutex{},
			FailResponses:   NewSliceBuffer[CallResult](DefaultCallResultBufLen),
		},
		errsMu:            &sync.Mutex{},
		errs:              NewSliceBuffer[string](DefaultCallResultBufLen),
		stats:             &Stats{},
		loki:              loki,
		Log:               l,
		lokiResponsesChan: make(chan CallResult, 50000),
	}, nil
}

// setupSchedule set up initial data for both RPS and VirtualUser load types
func (g *Generator) setupSchedule() {
	g.currentSegment = g.scheduleSegments[0]
	g.stats.LastSegment.Store(int64(len(g.scheduleSegments)))
	switch g.cfg.LoadType {
	case RPSScheduleType:
		g.ResponsesWaitGroup.Add(1)
		g.currentSegment.rl = ratelimit.New(int(g.currentSegment.From))
		g.stats.CurrentRPS.Store(g.currentSegment.From)

		// we run pacedCall controlled by stats.CurrentRPS
		go func() {
			for {
				select {
				case <-g.ResponsesCtx.Done():
					g.ResponsesWaitGroup.Done()
					g.Log.Info().Msg("RPS generator stopped")
					return
				default:
					g.pacedCall()
				}
			}
		}()
	case VUScheduleType:
		g.stats.CurrentVUs.Store(g.currentSegment.From)
		// we start all vus once
		vus := g.stats.CurrentVUs.Load()
		for i := 0; i < int(vus); i++ {
			inst := g.vu.Clone(g)
			g.runVU(inst)
			g.vus = append(g.vus, inst)
		}
	}
}

// runVU performs virtual user lifecycle
func (g *Generator) runVU(vu VirtualUser) {
	g.ResponsesWaitGroup.Add(1)
	if err := vu.Setup(g); err != nil {
		g.Stop()
	}
	go func() {
		defer g.ResponsesWaitGroup.Done()
		for {
			select {
			case <-g.ResponsesCtx.Done():
				if err := vu.Teardown(g); err != nil {
					g.Stop()
				}
				return
			case <-vu.StopChan():
				if err := vu.Teardown(g); err != nil {
					g.Stop()
				}
				return
			default:
				vu.Call(g)
			}
		}
	}()
}

// processSegment change RPS or VUs accordingly
// changing both internal and Stats values to report
func (g *Generator) processSegment() bool {
	if g.stats.CurrentStep.Load() == g.currentSegment.Steps {
		g.stats.CurrentSegment.Add(1)
		g.stats.CurrentStep.Store(0)
		if g.stats.CurrentSegment.Load() == g.stats.LastSegment.Load() {
			g.Log.Info().Msg("Finished all schedule segments")
			return true
		}
		g.currentSegment = g.scheduleSegments[g.stats.CurrentSegment.Load()]
		switch g.cfg.LoadType {
		case RPSScheduleType:
			g.currentSegment.rl = ratelimit.New(int(g.currentSegment.From))
			g.stats.CurrentRPS.Store(g.currentSegment.From)
		case VUScheduleType:
			for idx := range g.vus {
				log.Debug().Msg("Removing vus")
				g.vus[idx].Stop(g)
			}
			g.vus = g.vus[len(g.vus):]
			g.stats.CurrentVUs.Store(g.currentSegment.From)
			for i := 0; i < int(g.currentSegment.From); i++ {
				inst := g.vu.Clone(g)
				g.runVU(inst)
				g.vus = append(g.vus, inst)
			}
		}
	}
	g.Log.Info().
		Int64("Segment", g.stats.CurrentSegment.Load()).
		Int64("Step", g.stats.CurrentStep.Load()).
		Int64("VUs", g.stats.CurrentVUs.Load()).
		Int64("RPS", g.stats.CurrentRPS.Load()).
		Msg("Scheduler step")
	return false
}

func (g *Generator) processStep() {
	defer g.stats.CurrentStep.Add(1)
	switch g.cfg.LoadType {
	case RPSScheduleType:
		newRPS := g.stats.CurrentRPS.Load() + g.currentSegment.Increase
		if newRPS <= 0 {
			newRPS = 1
		}
		g.currentSegment.rl = ratelimit.New(int(newRPS))
		g.stats.CurrentRPS.Store(newRPS)
	case VUScheduleType:
		if g.currentSegment.Increase == 0 {
			g.Log.Info().Msg("No vus changes, passing the step")
			return
		}
		if g.currentSegment.Increase > 0 {
			for i := 0; i < int(g.currentSegment.Increase); i++ {
				inst := g.vu.Clone(g)
				g.runVU(inst)
				g.vus = append(g.vus, inst)
				g.stats.CurrentVUs.Store(g.stats.CurrentVUs.Load() + 1)
			}
		} else {
			absInst := int(math.Abs(float64(g.currentSegment.Increase)))
			for i := 0; i < absInst; i++ {
				if g.stats.CurrentVUs.Load()+g.currentSegment.Increase <= 0 {
					g.Log.Info().Msg("VUs can't be 0, keeping one VU")
					continue
				}
				g.vus[0].Stop(g)
				g.vus = g.vus[1:]
				g.stats.CurrentVUs.Store(g.stats.CurrentVUs.Load() - 1)
			}
		}
	}
}

// runSchedule runs scheduling loop
// processing steps inside segments
// processing segments inside the whole schedule
func (g *Generator) runSchedule() {
	g.ResponsesWaitGroup.Add(1)
	go func() {
		defer g.ResponsesWaitGroup.Done()
		for {
			select {
			case <-g.ResponsesCtx.Done():
				g.Log.Info().Msg("Scheduler exited")
				return
			default:
				time.Sleep(g.currentSegment.StepDuration)
				if g.processSegment() {
					return
				}
				g.processStep()
			}
		}
	}()
}

// handleCallResult stores local metrics for CallResult, pushed them to Loki stream too if Loki is on
func (g *Generator) handleCallResult(res CallResult) {
	if g.cfg.LokiConfig != nil {
		g.lokiResponsesChan <- res
	}
	if res.Error != "" {
		g.stats.RunFailed.Store(true)
		g.stats.Failed.Add(1)

		g.errsMu.Lock()
		g.responsesData.failResponsesMu.Lock()
		g.errs.Append(res.Error)
		g.responsesData.FailResponses.Append(res)
		g.errsMu.Unlock()
		g.responsesData.failResponsesMu.Unlock()

		g.Log.Error().Str("Err", res.Error).Msg("load generator request failed")
	} else {
		g.stats.Success.Add(1)
		g.responsesData.okDataMu.Lock()
		g.responsesData.OKData.Append(res.Data)
		g.responsesData.okResponsesMu.Lock()
		g.responsesData.OKResponses.Append(res)
		g.responsesData.okDataMu.Unlock()
		g.responsesData.okResponsesMu.Unlock()
	}
}

// collectResults collects CallResult from all the VUs
func (g *Generator) collectResults() {
	if g.cfg.LoadType == RPSScheduleType {
		return
	}
	g.dataWaitGroup.Add(1)
	go func() {
		defer g.dataWaitGroup.Done()
		for {
			select {
			case <-g.dataCtx.Done():
				g.Log.Info().Msg("Collect data exited")
				return
			case res := <-g.ResponsesChan:
				if res.StartedAt != nil {
					res.Duration = time.Since(*res.StartedAt)
				}
				tn := time.Now()
				res.FinishedAt = &tn
				g.handleCallResult(res)
			}
		}
	}()
}

// pacedCall calls a gun according to a scheduleSegments or plain RPS
func (g *Generator) pacedCall() {
	g.currentSegment.rl.Take()
	result := make(chan CallResult)
	requestCtx, cancel := context.WithTimeout(context.Background(), g.cfg.CallTimeout)
	callStartTS := time.Now()
	g.ResponsesWaitGroup.Add(1)
	go func() {
		defer g.ResponsesWaitGroup.Done()
		select {
		case result <- g.gun.Call(g):
		case <-requestCtx.Done():
			ts := time.Now()
			cr := CallResult{Duration: time.Since(callStartTS), FinishedAt: &ts, Timeout: true, Error: ErrCallTimeout.Error()}
			if g.cfg.LokiConfig != nil {
				g.lokiResponsesChan <- cr
			}
			g.stats.RunFailed.Store(true)
			g.stats.CallTimeout.Add(1)

			g.errsMu.Lock()
			defer g.errsMu.Unlock()
			g.errs.Append(ErrCallTimeout.Error())

			g.responsesData.failResponsesMu.Lock()
			defer g.responsesData.failResponsesMu.Unlock()
			g.responsesData.FailResponses.Append(cr)
			return
		}
	}()
	g.ResponsesWaitGroup.Add(1)
	go func() {
		defer g.ResponsesWaitGroup.Done()
		select {
		case <-requestCtx.Done():
			return
		case res := <-result:
			defer close(result)
			res.Duration = time.Since(callStartTS)
			ts := time.Now()
			res.FinishedAt = &ts
			g.handleCallResult(res)
		}
		cancel()
	}()
}

// Run runs load loop until timeout or stop
func (g *Generator) Run(wait bool) (interface{}, bool) {
	g.Log.Info().Msg("Load generator started")
	g.printStatsLoop()
	if g.cfg.LokiConfig != nil {
		g.runLokiPromtailResponses()
		g.runLokiPromtailStats()
	}
	g.setupSchedule()
	g.collectResults()
	g.runSchedule()
	if wait {
		return g.Wait()
	}
	return nil, false
}

// Stop stops load generator, waiting for all calls for either finish or timeout
func (g *Generator) Stop() (interface{}, bool) {
	g.responsesCancel()
	return g.Wait()
}

// Wait waits until test ends
func (g *Generator) Wait() (interface{}, bool) {
	g.Log.Info().Msg("Waiting for all responses to finish")
	g.ResponsesWaitGroup.Wait()
	if g.cfg.LokiConfig != nil {
		g.handleLokiStatsPayload()
		g.dataCancel()
		g.dataWaitGroup.Wait()
		g.stopLokiStream()
	}
	return g.GetData(), g.stats.RunFailed.Load()
}

// InputSharedData returns the SharedData passed in Generator config
func (g *Generator) InputSharedData() interface{} {
	return g.cfg.SharedData
}

// Errors get all calls errors
func (g *Generator) Errors() []string {
	return g.errs.Data
}

// GetData get all calls data
func (g *Generator) GetData() *ResponseData {
	return g.responsesData
}

// Stats get all load stats
func (g *Generator) Stats() *Stats {
	return g.stats
}

/* Loki's methods to handle CallResult/Stats and stream it to Loki */

// stopLokiStream stops the Loki stream client
func (g *Generator) stopLokiStream() {
	if g.cfg.LokiConfig != nil && g.cfg.LokiConfig.URL != "" {
		g.Log.Info().Msg("Stopping Loki")
		g.loki.Stop()
		g.Log.Info().Msg("Loki exited")
	}
}

// handleLokiResponsePayload handles CallResult payload with adding default labels
func (g *Generator) handleLokiResponsePayload(cr CallResult) {
	ls := g.labels.Merge(model.LabelSet{
		"test_data_type": "responses",
	})
	// we are removing time.Time{} because when it marshalled to string it creates N responses for some Loki queries
	// and to minimize the payload, duration is already calculated at that point
	ts := cr.FinishedAt
	cr.StartedAt = nil
	cr.FinishedAt = nil
	err := g.loki.HandleStruct(ls, *ts, cr)
	if err != nil {
		g.Log.Err(err).Send()
	}
}

// handleLokiStatsPayload handles StatsJSON payload with adding default labels
func (g *Generator) handleLokiStatsPayload() {
	ls := g.labels.Merge(model.LabelSet{
		"test_data_type": "stats",
	})
	err := g.loki.HandleStruct(ls, time.Now(), g.StatsJSON())
	if err != nil {
		g.Log.Err(err).Send()
	}
}

// runLokiPromtailResponses pushes CallResult to Loki
func (g *Generator) runLokiPromtailResponses() {
	g.Log.Info().
		Str("URL", g.cfg.LokiConfig.URL).
		Interface("DefaultLabels", g.cfg.Labels).
		Msg("Streaming data to Loki")
	g.dataWaitGroup.Add(1)
	go func() {
		defer g.dataWaitGroup.Done()
		for {
			select {
			case <-g.dataCtx.Done():
				g.Log.Info().Msg("Loki responses exited")
				return
			case r := <-g.lokiResponsesChan:
				g.handleLokiResponsePayload(r)
			}
		}
	}()
}

// runLokiPromtailStats pushes Stats payloads to Loki
func (g *Generator) runLokiPromtailStats() {
	g.dataWaitGroup.Add(1)
	go func() {
		defer g.dataWaitGroup.Done()
		for {
			select {
			case <-g.dataCtx.Done():
				g.Log.Info().Msg("Loki stats exited")
				return
			default:
				time.Sleep(g.cfg.StatsPollInterval)
				g.handleLokiStatsPayload()
			}
		}
	}()
}

/* Local logging methods */

// StatsJSON get all load stats for export
func (g *Generator) StatsJSON() map[string]interface{} {
	return map[string]interface{}{
		"current_rps":       g.stats.CurrentRPS.Load(),
		"current_instances": g.stats.CurrentVUs.Load(),
		"run_stopped":       g.stats.RunStopped.Load(),
		"run_failed":        g.stats.RunFailed.Load(),
		"failed":            g.stats.Failed.Load(),
		"success":           g.stats.Success.Load(),
		"callTimeout":       g.stats.CallTimeout.Load(),
	}
}

// printStatsLoop prints stats periodically, with Config.StatsPollInterval
func (g *Generator) printStatsLoop() {
	g.ResponsesWaitGroup.Add(1)
	go func() {
		defer g.ResponsesWaitGroup.Done()
		for {
			select {
			case <-g.ResponsesCtx.Done():
				g.Log.Info().Msg("Stats loop exited")
				return
			default:
				time.Sleep(g.cfg.StatsPollInterval)
				g.Log.Info().
					Int64("Success", g.stats.Success.Load()).
					Int64("Failed", g.stats.Failed.Load()).
					Int64("CallTimeout", g.stats.CallTimeout.Load()).
					Msg("Load stats")
			}
		}
	}()
}

// LabelsMapToModel create model.LabelSet from map of labels
func LabelsMapToModel(m map[string]string) model.LabelSet {
	ls := model.LabelSet{}
	for k, v := range m {
		ls[model.LabelName(k)] = model.LabelValue(v)
	}
	return ls
}
