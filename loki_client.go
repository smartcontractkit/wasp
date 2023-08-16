package wasp

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"errors"
	"github.com/grafana/dskit/backoff"
	dskit "github.com/grafana/dskit/flagext"
	lokiAPI "github.com/grafana/loki/clients/pkg/promtail/api"
	lokiClient "github.com/grafana/loki/clients/pkg/promtail/client"
	lokiProto "github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
	"strings"
)

// LokiLogWrapper wraps Loki errors received through logs, handles them
type LokiLogWrapper struct {
	IgnoreErrors bool
	client       *LokiClient
}

func (m *LokiLogWrapper) SetClient(c *LokiClient) {
	m.client = c
}

func (m *LokiLogWrapper) Log(kvars ...interface{}) error {
	// in case any batch send can not succeed we exit immediately
	// test metrics may be rate-limited, or we can't push them
	// if IgnoreErrors = true we proceed in any case
	if _, ok := kvars[13].(error); ok {
		if kvars[13].(error) != nil {
			log.Error().
				Interface("Status", kvars[9]).
				Str("Error", kvars[13].(error).Error()).
				Msg("Loki error")
			if !m.IgnoreErrors {
				os.Exit(1)
			}
		}
	}
	log.Trace().Interface("Line", kvars).Msg("Loki client internal log")
	return nil
}

// LokiClient is a Loki/Promtail client wrapper
type LokiClient struct {
	logWrapper *LokiLogWrapper
	lokiClient.Client
}

// Handle handles adding a new label set and a message to the batch
func (m *LokiClient) Handle(ls model.LabelSet, t time.Time, s string) error {
	log.Trace().
		Interface("Labels", ls).
		Time("Time", t).
		Str("Data", s).
		Msg("Sending data to Loki")
	m.Client.Chan() <- lokiAPI.Entry{Labels: ls, Entry: lokiProto.Entry{Timestamp: t, Line: s}}
	return nil
}

// HandleStruct handles adding a new label set and a message to the batch, marshalling JSON from struct
func (m *LokiClient) HandleStruct(ls model.LabelSet, t time.Time, st interface{}) error {
	d, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("failed to marshal struct in response: %v", st)
	}
	log.Trace().
		Interface("Labels", ls).
		Time("Time", t).
		Str("Data", string(d)).
		Msg("Sending data to Loki")
	m.Client.Chan() <- lokiAPI.Entry{Labels: ls, Entry: lokiProto.Entry{Timestamp: t, Line: string(d)}}
	return nil
}

// Stop stops the client goroutine
func (m *LokiClient) Stop() {
	m.Client.Stop()
}

// LokiConfig is simplified subset of a Promtail client configuration
type LokiConfig struct {
	// URL url to Loki endpoint
	URL string `yaml:"url"`
	// Token is Loki authorization token
	Token string `yaml:"token"`
	// BasicAuth is a basic login:password auth string
	BasicAuth string `yaml:"basic_auth"`
	// IgnoreErrors ignore any loki client errors, do not fail the test
	IgnoreErrors bool
	// BatchWait max time to wait until sending a new batch
	BatchWait time.Duration
	// BatchSize size of a messages batch
	BatchSize int
	// Timeout is batch send timeout
	Timeout time.Duration
	// BackoffConfig backoff configuration
	BackoffConfig backoff.Config
	// Headers are additional request headers
	Headers map[string]string
	// The tenant ID to use when pushing logs to Loki (empty string means
	// single tenant mode)
	TenantID string
	// When enabled, Promtail will not retry batches that get a
	// 429 'Too Many Requests' response from the distributor. Helps
	// prevent HOL blocking in multitenant deployments.
	DropRateLimitedBatches bool
	// ExposePrometheusMetrics if enabled exposes Promtail Prometheus metrics
	ExposePrometheusMetrics bool
	MaxStreams              int
	MaxLineSize             int
	MaxLineSizeTruncate     bool
}

func NewEnvLokiConfig() *LokiConfig {
	return &LokiConfig{
		TenantID:                os.Getenv("LOKI_TENANT_ID"),
		URL:                     os.Getenv("LOKI_URL"),
		Token:                   os.Getenv("LOKI_TOKEN"),
		BasicAuth:               os.Getenv("LOKI_BASIC_AUTH"),
		IgnoreErrors:            true,
		BatchWait:               5 * time.Second,
		BatchSize:               500 * 1024,
		Timeout:                 20 * time.Second,
		DropRateLimitedBatches:  false,
		ExposePrometheusMetrics: false,
		MaxStreams:              30,
		MaxLineSize:             999999,
		MaxLineSizeTruncate:     false,
	}
}

// NewLokiClient creates a new Promtail client
func NewLokiClient(extCfg *LokiConfig) (*LokiClient, error) {
	serverURL := dskit.URLValue{}
	err := serverURL.Set(extCfg.URL)
	if err != nil {
		return nil, err
	}
	cfg := lokiClient.Config{
		URL:                    serverURL,
		BatchWait:              extCfg.BatchWait,
		BatchSize:              extCfg.BatchSize,
		Timeout:                extCfg.Timeout,
		DropRateLimitedBatches: extCfg.DropRateLimitedBatches,
		BackoffConfig:          extCfg.BackoffConfig,
		Headers:                extCfg.Headers,
		TenantID:               extCfg.TenantID,
		Client: config.HTTPClientConfig{
			TLSConfig: config.TLSConfig{InsecureSkipVerify: true},
		},
	}
	if extCfg.BasicAuth != "" {
		logpass := strings.Split(extCfg.BasicAuth, ":")
		if len(logpass) != 2 {
			return nil, errors.New("basic auth should be in login:password format")
		}
		cfg.Client.BasicAuth = &config.BasicAuth{
			Username: logpass[0],
			Password: config.Secret(logpass[1]),
		}
	}
	if extCfg.Token != "" {
		cfg.Client.BearerToken = config.Secret(extCfg.Token)
	}
	ll := &LokiLogWrapper{IgnoreErrors: extCfg.IgnoreErrors}
	c, err := lokiClient.New(lokiClient.NewMetrics(nil), cfg, extCfg.MaxStreams, extCfg.MaxLineSize, extCfg.MaxLineSizeTruncate, ll)
	if err != nil {
		return nil, err
	}
	lc := &LokiClient{
		logWrapper: ll,
		Client:     c,
	}
	ll.SetClient(lc)
	return lc, nil
}
