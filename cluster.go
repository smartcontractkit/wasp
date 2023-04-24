package wasp

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoChart     = errors.New("cluster chart path is empty")
	ErrNoNamespace = errors.New("namespace is empty")
	ErrNoTimeout   = errors.New("timeout shouldn't be zero")
	ErrNoSync      = errors.New("HelmValues should contain \"sync\" field used to track your cluster jobs")
	ErrNoJobs      = errors.New("HelmValues should contain \"jobs\" field used to scale your cluster jobs")
)

// ClusterConfig defines k8s jobs settings
type ClusterConfig struct {
	ChartPath  string
	Namespace  string
	Timeout    time.Duration
	KeepJobs   bool
	HelmValues map[string]string
}

func (m *ClusterConfig) Defaults() {
	m.HelmValues["namespace"] = m.Namespace
}

func (m *ClusterConfig) Validate() (err error) {
	if m.ChartPath == "" {
		_ = errors.Join(err, ErrNoChart)
	}
	if m.Namespace == "" {
		_ = errors.Join(err, ErrNoNamespace)
	}
	if m.Timeout == 0 {
		_ = errors.Join(err, ErrNoTimeout)
	}
	if m.HelmValues["sync"] == "" {
		_ = errors.Join(err, ErrNoSync)
	}
	if m.HelmValues["jobs"] == "" {
		_ = errors.Join(err, ErrNoJobs)
	}
	return
}

// ClusterProfile is a k8s cluster test for some workload profile
type ClusterProfile struct {
	cfg    *ClusterConfig
	c      *K8sClient
	Ctx    context.Context
	Cancel context.CancelFunc
}

// NewClusterProfile creates new cluster profile
func NewClusterProfile(cfg *ClusterConfig) (*ClusterProfile, error) {
	InitDefaultLogging()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.Defaults()
	ctx, cancelFunc := context.WithTimeout(context.Background(), cfg.Timeout)
	return &ClusterProfile{
		cfg:    cfg,
		c:      NewK8sClient(),
		Ctx:    ctx,
		Cancel: cancelFunc,
	}, nil
}

func (m *ClusterProfile) deployHelm(testName string) error {
	var cmd strings.Builder
	cmd.WriteString(fmt.Sprintf("helm install %s %s", testName, m.cfg.ChartPath))
	for k, v := range m.cfg.HelmValues {
		cmd.WriteString(fmt.Sprintf(" --set %s=%s", k, v))
	}
	cmd.WriteString(fmt.Sprintf(" -n %s", m.cfg.Namespace))
	log.Info().Str("Cmd", cmd.String()).Msg("Deploying jobs")
	return ExecCmd(cmd.String())
}

// Run starts a new test
func (m *ClusterProfile) Run() error {
	testName := uuid.NewString()[0:5]
	if err := m.deployHelm(testName); err != nil {
		return err
	}
	jobNum, err := strconv.Atoi(m.cfg.HelmValues["jobs"])
	if err != nil {
		return err
	}
	return m.c.TrackJobs(m.Ctx, m.cfg.Namespace, m.cfg.HelmValues["sync"], jobNum, m.cfg.KeepJobs)
}
