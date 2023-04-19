package wasp

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	dskit "github.com/grafana/dskit/flagext"
	lokiAPI "github.com/grafana/loki/clients/pkg/promtail/api"
	lokiClient "github.com/grafana/loki/clients/pkg/promtail/client"
	lokiProto "github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type LocalLogger struct{}

func (m *LocalLogger) Log(kvars ...interface{}) error {
	for _, v := range kvars {
		log.Debug().Interface("Key", v).Msg("Loki client internal log")
	}
	return nil
}

// LokiClient is a Loki/Promtail client wrapper
type LokiClient struct {
	lokiClient.Client
}

// Handle handles adding a new label set and a message to the batch
func (m *LokiClient) Handle(ls model.LabelSet, t time.Time, s string) error {
	log.Debug().
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
	log.Debug().
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

// LokiConfig Loki/Promtail client configuration
type LokiConfig struct {
	// URL url to Loki endpoint
	URL string `yaml:"url"`
	// Token is Loki authorization token
	Token string `yaml:"token"`
	// BatchWait max time to wait until sending a new batch
	BatchWait time.Duration `yaml:"batch_wait"`
	// BatchSize size of a messages batch
	BatchSize int `yaml:"batch_size"`
	// Timeout is a batch send timeout
	Timeout time.Duration `yaml:"timeout"`
}

func NewDefaultLokiConfig(url string, token string) *LokiConfig {
	return &LokiConfig{
		URL:       url,
		Token:     token,
		BatchWait: 5 * time.Second,
		BatchSize: 500 * 1024,
		Timeout:   20 * time.Second,
	}
}

func NewEnvLokiConfig() *LokiConfig {
	return &LokiConfig{
		URL:       os.Getenv("LOKI_URL"),
		Token:     os.Getenv("LOKI_TOKEN"),
		BatchWait: 5 * time.Second,
		BatchSize: 500 * 1024,
		Timeout:   20 * time.Second,
	}
}

// NewLokiClient creates a new Loki/Promtail client
func NewLokiClient(extCfg *LokiConfig) (*LokiClient, error) {
	serverURL := dskit.URLValue{}
	err := serverURL.Set(extCfg.URL)
	if err != nil {
		return nil, err
	}
	cfg := lokiClient.Config{
		URL:       serverURL,
		BatchWait: extCfg.BatchWait,
		BatchSize: extCfg.BatchSize,
		Timeout:   extCfg.Timeout,
		Client:    config.HTTPClientConfig{BearerToken: config.Secret(extCfg.Token)},
	}
	c, err := lokiClient.New(lokiClient.NewMetrics(nil), cfg, 10, 999999, false, &LocalLogger{})
	if err != nil {
		return nil, err
	}
	return &LokiClient{
		Client: c,
	}, nil
}
