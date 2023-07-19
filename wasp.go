package wasp

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/rs/zerolog"
	"go.uber.org/ratelimit"
)

const (
	DefaultCallTimeout           = 1 * time.Minute
	DefaultStatsPollInterval     = 5 * time.Second
	DefaultRateLimitUnitDuration = 1 * time.Second
	DefaultCallResultBufLen      = 50000
	DefaultGenName               = "Generator"
)

var (
	ErrNoCfg               = errors.New("config is nil")
	ErrNoImpl              = errors.New("either \"gun\" or \"vu\" implementation must provided")
	ErrNoSchedule          = errors.New("no schedule segments were provided")
	ErrInvalidScheduleType = errors.New("schedule type must be either of wasp.RPS, wasp.VU, use package constants")
	ErrCallTimeout         = errors.New("generator request call timeout")
	ErrStartFrom           = errors.New("from must be > 0")
	ErrInvalidSteps        = errors.New("both \"Steps\" and \"StepsDuration\" must be defined in a schedule segment")
	ErrNoGun               = errors.New("rps load scheduleSegments selected but gun implementation is nil")
	ErrNoVU                = errors.New("vu load scheduleSegments selected but vu implementation is nil")
)

// Gun is basic interface for some synthetic load test implementation
// Call performs one request according to some RPS schedule
type Gun interface {
	Call(l *Generator) *CallResult
}

// VirtualUser is basic interface to run virtual users load
// you should use it if:
// - your protocol is stateful, ex.: ws, grpc
// - you'd like to have some VirtualUser modelling, perform sequential requests
type VirtualUser interface {
	Call(l *Generator)
	Clone(l *Generator) VirtualUser
	Setup(l *Generator) error
	Teardown(l *Generator) error
	Stop(l *Generator)
	StopChan() chan struct{}
}

// CallResult represents basic call result info
type CallResult struct {
	Failed     bool          `json:"failed,omitempty"`
	Timeout    bool          `json:"timeout,omitempty"`
	Duration   time.Duration `json:"duration"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	Group      string        `json:"group"`
	Data       interface{}   `json:"data,omitempty"`
	Error      string        `json:"error,omitempty"`
}

type ScheduleType string

const (
	RPS ScheduleType = "rps_schedule"
	VU  ScheduleType = "vu_schedule"
)

// Segment load test schedule segment
type Segment struct {
	From         int64
	Increase     int64
	Steps        int64
	StepDuration time.Duration
}

func (ls *Segment) Validate() error {
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
	T                     *testing.T
	GenName               string
	LoadType              ScheduleType
	Labels                map[string]string
	LokiConfig            *LokiConfig
	Schedule              []*Segment
	RateLimitUnitDuration time.Duration
	CallResultBufLen      int
	StatsPollInterval     time.Duration
	CallTimeout           time.Duration
	Gun                   Gun
	VU                    VirtualUser
	Logger                zerolog.Logger
	SharedData            interface{}
	// calculated fields
	duration time.Duration
	// only available in cluster mode
	nodeID        string
	SamplerConfig *SamplerConfig
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
	if lgc.LoadType != RPS && lgc.LoadType != VU {
		return ErrInvalidScheduleType
	}
	if lgc.LoadType == RPS && lgc.Gun == nil {
		return ErrNoGun
	}
	if lgc.LoadType == VU && lgc.VU == nil {
		return ErrNoVU
	}
	if lgc.RateLimitUnitDuration == 0 {
		lgc.RateLimitUnitDuration = DefaultRateLimitUnitDuration
	}
	return nil
}

// Stats basic generator load stats
type Stats struct {
	// TODO: update json labels with dashboards on major release
	CurrentRPS      atomic.Int64 `json:"currentRPS"`
	CurrentTimeUnit int64        `json:"current_time_unit"`
	CurrentVUs      atomic.Int64 `json:"currentVUs"`
	LastSegment     atomic.Int64 `json:"last_segment"`
	CurrentSegment  atomic.Int64 `json:"current_schedule_segment"`
	CurrentStep     atomic.Int64 `json:"current_schedule_step"`
	SamplesRecorded atomic.Int64 `json:"samples_recorded"`
	SamplesSkipped  atomic.Int64 `json:"samples_skipped"`
	RunStopped      atomic.Bool  `json:"runStopped"`
	RunFailed       atomic.Bool  `json:"runFailed"`
	Success         atomic.Int64 `json:"success"`
	Failed          atomic.Int64 `json:"failed"`
	CallTimeout     atomic.Int64 `json:"callTimeout"`
	Duration        int64        `json:"load_duration"`
}

// ResponseData includes any request/response data that a gun might store
// ok* slices usually contains successful responses and their verifications if their done async
// fail* slices contains CallResult with response data and an error
type ResponseData struct {
	okDataMu        *sync.Mutex
	OKData          *SliceBuffer[any]
	okResponsesMu   *sync.Mutex
	OKResponses     *SliceBuffer[*CallResult]
	failResponsesMu *sync.Mutex
	FailResponses   *SliceBuffer[*CallResult]
}

// Generator generates load with some RPS
type Generator struct {
	cfg                *Config
	sampler            *Sampler
	Log                zerolog.Logger
	labels             model.LabelSet
	rl                 atomic.Pointer[ratelimit.Limiter]
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
	ResponsesChan      chan *CallResult
	Responses          *Responses
	responsesData      *ResponseData
	errsMu             *sync.Mutex
	errs               *SliceBuffer[string]
	stats              *Stats
	loki               *LokiClient
	lokiResponsesChan  chan *CallResult
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
		if err := s.Validate(); err != nil {
			return nil, err
		}
	}
	for _, s := range cfg.Schedule {
		segmentTotal := time.Duration(s.Steps) * s.StepDuration
		cfg.duration += segmentTotal
	}
	l := GetLogger(cfg.T, cfg.GenName)

	ls := LabelsMapToModel(cfg.Labels)
	if cfg.T != nil {
		ls = ls.Merge(model.LabelSet{
			"go_test_name": model.LabelValue(cfg.T.Name()),
		})
	}
	if cfg.GenName != "" {
		ls = ls.Merge(model.LabelSet{
			"gen_name": model.LabelValue(cfg.GenName),
		})
	}
	cfg.nodeID = os.Getenv("WASP_NODE_ID")
	// context for all requests/responses and vus
	responsesCtx, responsesCancel := context.WithTimeout(context.Background(), cfg.duration)
	// context for all the collected data
	dataCtx, dataCancel := context.WithCancel(context.Background())
	rch := make(chan *CallResult)
	g := &Generator{
		cfg:                cfg,
		sampler:            NewSampler(cfg.SamplerConfig),
		scheduleSegments:   cfg.Schedule,
		ResponsesWaitGroup: &sync.WaitGroup{},
		dataWaitGroup:      &sync.WaitGroup{},
		ResponsesCtx:       responsesCtx,
		responsesCancel:    responsesCancel,
		dataCtx:            dataCtx,
		dataCancel:         dataCancel,
		gun:                cfg.Gun,
		vu:                 cfg.VU,
		Responses:          NewResponses(rch),
		ResponsesChan:      rch,
		labels:             ls,
		responsesData: &ResponseData{
			okDataMu:        &sync.Mutex{},
			OKData:          NewSliceBuffer[any](DefaultCallResultBufLen),
			okResponsesMu:   &sync.Mutex{},
			OKResponses:     NewSliceBuffer[*CallResult](DefaultCallResultBufLen),
			failResponsesMu: &sync.Mutex{},
			FailResponses:   NewSliceBuffer[*CallResult](DefaultCallResultBufLen),
		},
		errsMu:            &sync.Mutex{},
		errs:              NewSliceBuffer[string](DefaultCallResultBufLen),
		stats:             &Stats{},
		Log:               l,
		lokiResponsesChan: make(chan *CallResult, 50000),
	}
	var err error
	if cfg.LokiConfig != nil {
		g.loki, err = NewLokiClient(cfg.LokiConfig)
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}

// setupSchedule set up initial data for both RPS and VirtualUser load types
func (g *Generator) setupSchedule() {
	g.currentSegment = g.scheduleSegments[0]
	g.stats.LastSegment.Store(int64(len(g.scheduleSegments)))
	switch g.cfg.LoadType {
	case RPS:
		g.ResponsesWaitGroup.Add(1)
		g.stats.CurrentRPS.Store(g.currentSegment.From)
		newRateLimit := ratelimit.New(int(g.currentSegment.From), ratelimit.Per(g.cfg.RateLimitUnitDuration))
		g.rl.Store(&newRateLimit)
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
	case VU:
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
		//pyroscope.TagWrapper(context.Background(), pyroscope.Labels("scope", "vuCall"), func(c context.Context) {
		for {
			startedAt := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), g.cfg.CallTimeout)
			vuChan := make(chan struct{})
			go func() {
				vu.Call(g)
				select {
				case <-ctx.Done():
					return
				default:
					vuChan <- struct{}{}
				}
			}()
			select {
			case <-g.ResponsesCtx.Done():
				if err := vu.Teardown(g); err != nil {
					g.Stop()
				}
				cancel()
				return
			case <-vu.StopChan():
				if err := vu.Teardown(g); err != nil {
					g.Stop()
				}
				cancel()
				return
			case <-ctx.Done():
				g.ResponsesChan <- &CallResult{StartedAt: &startedAt, Error: ErrCallTimeout.Error(), Timeout: true}
				cancel()
			case <-vuChan:
			}
		}
		//})
	}()
}

// processSegment change RPS or VUs accordingly
// changing both internal and Stats values to report
func (g *Generator) processSegment() bool {
	if g.stats.CurrentStep.Load() == g.currentSegment.Steps {
		g.stats.CurrentSegment.Add(1)
		g.stats.CurrentStep.Store(0)
		if g.stats.CurrentSegment.Load() == g.stats.LastSegment.Load() {
			return true
		}
		g.currentSegment = g.scheduleSegments[g.stats.CurrentSegment.Load()]
		switch g.cfg.LoadType {
		case RPS:
			newRateLimit := ratelimit.New(int(g.currentSegment.From), ratelimit.Per(g.cfg.RateLimitUnitDuration))
			g.rl.Store(&newRateLimit)
			g.stats.CurrentRPS.Store(g.currentSegment.From)
		case VU:
			for idx := range g.vus {
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
		Msg("Schedule segment")
	return false
}

func (g *Generator) processStep() {
	defer g.stats.CurrentStep.Add(1)
	switch g.cfg.LoadType {
	case RPS:
		newRPS := g.stats.CurrentRPS.Load() + g.currentSegment.Increase
		if newRPS <= 0 {
			newRPS = 1
		}
		newRateLimit := ratelimit.New(int(newRPS), ratelimit.Per(g.cfg.RateLimitUnitDuration))
		g.rl.Store(&newRateLimit)
		g.stats.CurrentRPS.Store(newRPS)
	case VU:
		if g.currentSegment.Increase == 0 {
			g.Log.Info().Msg("No VUs changes, constant load step")
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
					g.Log.Info().Msg("VUs can't be 0, keeping at least one VU up")
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
				g.Log.Info().
					Int64("Segment", g.stats.CurrentSegment.Load()).
					Int64("Step", g.stats.CurrentStep.Load()).
					Int64("VUs", g.stats.CurrentVUs.Load()).
					Int64("RPS", g.stats.CurrentRPS.Load()).
					Msg("Finished all schedule segments")
				g.Log.Info().Msg("Scheduler exited")
				return
			default:
				if g.processSegment() {
					return
				}
				g.processStep()
				time.Sleep(g.currentSegment.StepDuration)
			}
		}
	}()
}

// storeCallResult stores local metrics for CallResult, pushed them to Loki stream too if Loki is on
func (g *Generator) storeCallResult(res *CallResult) {
	if g.cfg.CallTimeout > 0 && res.Duration > g.cfg.CallTimeout && !res.Timeout {
		return
	}
	if !g.sampler.ShouldRecord(res, g.stats) {
		return
	}
	if g.cfg.LokiConfig != nil {
		g.lokiResponsesChan <- res
	}
	g.responsesData.okDataMu.Lock()
	g.responsesData.failResponsesMu.Lock()
	g.errsMu.Lock()
	if res.Failed {
		g.stats.RunFailed.Store(true)
		g.stats.Failed.Add(1)
		g.errs.Append(res.Error)
		g.responsesData.FailResponses.Append(res)
		g.Log.Error().Str("Err", res.Error).Msg("load generator request failed")
	} else if res.Timeout {
		g.stats.RunFailed.Store(true)
		g.stats.CallTimeout.Add(1)
		g.stats.Failed.Add(1)
		g.errs.Append(res.Error)
		g.responsesData.FailResponses.Append(res)
		g.Log.Error().Msg("load generator request timed out")
	} else {
		g.stats.Success.Add(1)
		g.responsesData.OKData.Append(res.Data)
		g.responsesData.OKResponses.Append(res)
	}
	g.responsesData.okDataMu.Unlock()
	g.responsesData.failResponsesMu.Unlock()
	g.errsMu.Unlock()
}

// collectVUResults collects CallResult from all the VUs
func (g *Generator) collectVUResults() {
	if g.cfg.LoadType == RPS {
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
				g.storeCallResult(res)
			}
		}
	}()
}

// pacedCall calls a gun according to a scheduleSegments or plain RPS
func (g *Generator) pacedCall() {
	//pyroscope.TagWrapper(context.Background(), pyroscope.Labels("scope", "rpsCall"), func(c context.Context) {
	l := *g.rl.Load()
	l.Take()
	result := make(chan *CallResult)
	requestCtx, cancel := context.WithTimeout(context.Background(), g.cfg.CallTimeout)
	callStartTS := time.Now()
	g.ResponsesWaitGroup.Add(1)
	go func() {
		defer g.ResponsesWaitGroup.Done()
		select {
		case result <- g.gun.Call(g):
		case <-requestCtx.Done():
			ts := time.Now()
			cr := &CallResult{Duration: time.Since(callStartTS), FinishedAt: &ts, Timeout: true, Error: ErrCallTimeout.Error()}
			g.storeCallResult(cr)
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
			g.storeCallResult(res)
		}
		cancel()
	}()
	//})
}

// Run runs load loop until timeout or stop
func (g *Generator) Run(wait bool) (interface{}, bool) {
	g.Log.Info().Msg("Load generator started")
	g.printStatsLoop()
	if g.cfg.LokiConfig != nil {
		g.storeCallResultLoki()
		g.runPromtailStats()
	}
	g.setupSchedule()
	g.collectVUResults()
	g.runSchedule()
	if wait {
		return g.Wait()
	}
	return nil, false
}

// Stop stops load generator, waiting for all calls for either finish or timeout
// this method is external so Gun/VU implementations can stop the generator
func (g *Generator) Stop() (interface{}, bool) {
	if g.stats.RunStopped.Load() {
		return nil, true
	}
	g.stats.RunStopped.Store(true)
	g.stats.RunFailed.Store(true)
	g.Log.Warn().Msg("Graceful stop")
	g.responsesCancel()
	return g.Wait()
}

// Wait waits until test ends
func (g *Generator) Wait() (interface{}, bool) {
	g.Log.Info().Msg("Waiting for all responses to finish")
	g.ResponsesWaitGroup.Wait()
	g.stats.Duration = g.cfg.duration.Nanoseconds()
	g.stats.CurrentTimeUnit = g.cfg.RateLimitUnitDuration.Nanoseconds()
	if g.cfg.LokiConfig != nil {
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
// adding custom CallResult labels if present
func (g *Generator) handleLokiResponsePayload(cr *CallResult) {
	labels := g.labels.Merge(model.LabelSet{
		"test_data_type": "responses",
		CallGroupLabel:   model.LabelValue(cr.Group),
	})
	// we are removing time.Time{} because when it marshalled to string it creates N responses for some Loki queries
	// and to minimize the payload, duration is already calculated at that point
	ts := cr.FinishedAt
	cr.StartedAt = nil
	cr.FinishedAt = nil
	err := g.loki.HandleStruct(labels, *ts, cr)
	if err != nil {
		g.Log.Err(err).Send()
	}
}

// handleLokiStatsPayload handles StatsJSON payload with adding default labels
// this stream serves as a debug data and shouldn't be customized with additional labels
func (g *Generator) handleLokiStatsPayload() {
	ls := g.labels.Merge(model.LabelSet{
		"test_data_type": "stats",
	})
	err := g.loki.HandleStruct(ls, time.Now(), g.StatsJSON())
	if err != nil {
		g.Log.Err(err).Send()
	}
}

// storeCallResultLoki pushes CallResult to Loki using Promtail
func (g *Generator) storeCallResultLoki() {
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

// runPromtailStats pushes Stats payloads to Loki
func (g *Generator) runPromtailStats() {
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
		"node_id":           g.cfg.nodeID,
		"current_rps":       g.stats.CurrentRPS.Load(),
		"current_instances": g.stats.CurrentVUs.Load(),
		"samples_recorded":  g.stats.SamplesRecorded.Load(),
		"samples_skipped":   g.stats.SamplesSkipped.Load(),
		"run_stopped":       g.stats.RunStopped.Load(),
		"run_failed":        g.stats.RunFailed.Load(),
		"failed":            g.stats.Failed.Load(),
		"success":           g.stats.Success.Load(),
		"callTimeout":       g.stats.CallTimeout.Load(),
		"load_duration":     g.stats.Duration,
		"current_time_unit": g.stats.CurrentTimeUnit,
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

type SamplerConfig struct {
	SuccessfulCallResultRecordRatio int
}

// Sampler is a CallResult filter that stores a percentage of successful call results
// errored and timed out results are always stored
type Sampler struct {
	cfg *SamplerConfig
}

// NewSampler creates new Sampler
func NewSampler(cfg *SamplerConfig) *Sampler {
	if cfg == nil {
		cfg = &SamplerConfig{SuccessfulCallResultRecordRatio: 100}
	}
	return &Sampler{cfg: cfg}
}

// ShouldRecord return true if we should save CallResult
func (m *Sampler) ShouldRecord(cr *CallResult, s *Stats) bool {
	if cr.Error != "" || cr.Failed || cr.Timeout {
		s.SamplesRecorded.Add(1)
		return true
	}
	if m.cfg.SuccessfulCallResultRecordRatio == 0 {
		s.SamplesSkipped.Add(1)
		return false
	}
	if m.cfg.SuccessfulCallResultRecordRatio == 100 {
		s.SamplesRecorded.Add(1)
		return true
	}
	r := rand.Intn(100)
	if cr.Error == "" && r < m.cfg.SuccessfulCallResultRecordRatio {
		s.SamplesRecorded.Add(1)
		return true
	}
	s.SamplesSkipped.Add(1)
	return false
}
