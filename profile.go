package wasp

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-testing-framework/grafana"
)

// Profile is a set of concurrent generators forming some workload profile
type Profile struct {
	Generators                  []*Generator
	testEndedWg                 *sync.WaitGroup
	bootstrapErr                error
	grafanaAPI                  *grafana.Client
	annotateDashboardUIDs       []string
	checkAlertsForDashboardUIDs []string
	startTime                   time.Time
	endTime                     time.Time
	waitBeforeAlertCheck        time.Duration // Cooldown period to wait before annotating and checking for grafana alerts
}

// Run runs all generators and wait until they finish
func (m *Profile) Run(wait bool) (*Profile, error) {
	if m.bootstrapErr != nil {
		return m, m.bootstrapErr
	}
	if err := waitSyncGroupReady(); err != nil {
		return m, err
	}
	m.startTime = time.Now()
	if len(m.annotateDashboardUIDs) > 0 {
		m.annotateRunStartOnGrafana()
	}
	for _, g := range m.Generators {
		g.Run(false)
	}
	if wait {
		m.Wait()
	}
	m.endTime = time.Now()
	if len(m.annotateDashboardUIDs) > 0 {
		m.annotateRunEndOnGrafana()
	}
	if len(m.checkAlertsForDashboardUIDs) > 0 {
		if m.waitBeforeAlertCheck > 0 {
			log.Info().Msgf("Waiting %s before checking for alerts..", m.waitBeforeAlertCheck)
			time.Sleep(m.waitBeforeAlertCheck)
			m.annotateAlertCheckOnGrafana()
		}

		alerts, err := CheckDashboardAlerts(m.grafanaAPI, m.startTime, time.Now(), m.checkAlertsForDashboardUIDs)
		if len(alerts) > 0 {
			log.Info().Msgf("Alerts found\n%s", grafana.FormatAlertsTable(alerts))
		}
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

func (m *Profile) annotateRunStartOnGrafana() {
	for _, dashboardID := range m.annotateDashboardUIDs {
		a := grafana.PostAnnotation{
			DashboardUID: dashboardID,
			Time:         &m.startTime,
			Text:         "Load test started",
		}
		_, err := m.grafanaAPI.PostAnnotation(a)
		if err != nil {
			log.Warn().Msgf("could not annotate on Grafana: %s", err)
		}
	}
}

func (m *Profile) annotateRunEndOnGrafana() {
	for _, dashboardID := range m.annotateDashboardUIDs {
		a := grafana.PostAnnotation{
			DashboardUID: dashboardID,
			Time:         &m.endTime,
			Text:         "Load test ended",
		}
		_, err := m.grafanaAPI.PostAnnotation(a)
		if err != nil {
			log.Warn().Msgf("could not annotate on Grafana: %s", err)
		}
	}
}

func (m *Profile) annotateAlertCheckOnGrafana() {
	t := time.Now()
	for _, dashboardID := range m.annotateDashboardUIDs {
		a := grafana.PostAnnotation{
			DashboardUID: dashboardID,
			Time:         &t,
			Text:         "Grafana alert check after load test",
		}
		_, err := m.grafanaAPI.PostAnnotation(a)
		if err != nil {
			log.Warn().Msgf("could not annotate on Grafana: %s", err)
		}
	}
}

// Pause pauses execution of all generators
func (m *Profile) Pause() {
	for _, g := range m.Generators {
		g.Pause()
	}
}

// Resume resumes execution of all generators
func (m *Profile) Resume() {
	for _, g := range m.Generators {
		g.Resume()
	}
}

// Wait waits until all generators have finished the workload
func (m *Profile) Wait() {
	for _, g := range m.Generators {
		g := g
		m.testEndedWg.Add(1)
		go func() {
			defer m.testEndedWg.Done()
			g.Wait()
		}()
	}
	m.testEndedWg.Wait()
}

// NewProfile creates new VU or Gun profile from parts
func NewProfile() *Profile {
	return &Profile{Generators: make([]*Generator, 0), testEndedWg: &sync.WaitGroup{}}
}

func (m *Profile) Add(g *Generator, err error) *Profile {
	if err != nil {
		m.bootstrapErr = err
		return m
	}
	m.Generators = append(m.Generators, g)
	return m
}

type GrafanaOpts struct {
	GrafanaURL                   string        `toml:"grafana_url"`
	GrafanaToken                 string        `toml:"grafana_token_secret"`
	WaitBeforeAlertCheck         time.Duration `toml:"grafana_wait_before_alert_check"`                  // Cooldown period to wait before checking for alerts
	AnnotateDashboardUIDs        []string      `toml:"grafana_annotate_dashboard_uids"`                  // Grafana dashboardUIDs to annotate start and end of the run
	CheckDashboardAlertsAfterRun []string      `toml:"grafana_check_alerts_after_run_on_dashboard_uids"` // Grafana dashboardIds to check for alerts after run
}

func (m *Profile) WithGrafana(opts *GrafanaOpts) *Profile {
	m.grafanaAPI = grafana.NewGrafanaClient(opts.GrafanaURL, opts.GrafanaToken)
	m.annotateDashboardUIDs = opts.AnnotateDashboardUIDs
	m.checkAlertsForDashboardUIDs = opts.CheckDashboardAlertsAfterRun
	m.waitBeforeAlertCheck = opts.WaitBeforeAlertCheck
	return m
}

// waitSyncGroupReady awaits other pods with WASP_SYNC label to start before starting the test
func waitSyncGroupReady() error {
	if os.Getenv("WASP_NODE_ID") != "" {
		kc := NewK8sClient()
		jobNum, err := strconv.Atoi(os.Getenv("WASP_JOBS"))
		if err != nil {
			return err
		}
		if err := kc.waitSyncGroup(context.Background(), os.Getenv("WASP_NAMESPACE"), os.Getenv("WASP_SYNC"), jobNum); err != nil {
			return err
		}
	}
	return nil
}
