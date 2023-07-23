package wasp

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultArchiveName = "wasp-0.1.7.tgz"
)

//go:embed charts/wasp-0.1.7.tgz
var defaultChart []byte

var (
	ErrNoNamespace = errors.New("namespace is empty")
	ErrNoTimeout   = errors.New("timeout shouldn't be zero")
	ErrNoJobs      = errors.New("HelmValues should contain \"jobs\" field used to scale your cluster jobs, jobs must be > 1")
)

// ClusterConfig defines k8s jobs settings
type ClusterConfig struct {
	ChartPath       string
	Namespace       string
	Timeout         time.Duration
	KeepJobs        bool
	HelmValues      map[string]string
	tmpHelmFilePath string
}

func (m *ClusterConfig) Defaults() error {
	m.HelmValues["namespace"] = m.Namespace
	// nolint
	m.HelmValues["sync"] = fmt.Sprintf("%s", uuid.NewString()[0:5])
	if m.ChartPath == "" {
		log.Info().Msg("Using default embedded chart")
		f, err := os.CreateTemp(".", defaultArchiveName)
		//nolint
		defer f.Close()
		if err != nil {
			return err
		}
		if _, err := f.Write(defaultChart); err != nil {
			return err
		}
		m.tmpHelmFilePath, m.ChartPath = f.Name(), f.Name()
	}
	return nil
}

func (m *ClusterConfig) Validate() (err error) {
	if m.Namespace == "" {
		err = errors.Join(err, ErrNoNamespace)
	}
	if m.Timeout == 0 {
		err = errors.Join(err, ErrNoTimeout)
	}
	if m.HelmValues["jobs"] == "" || m.HelmValues["jobs"] == "1" {
		err = errors.Join(err, ErrNoJobs)
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
	if err := cfg.Defaults(); err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), cfg.Timeout)
	return &ClusterProfile{
		cfg:    cfg,
		c:      NewK8sClient(),
		Ctx:    ctx,
		Cancel: cancelFunc,
	}, nil
}

func (m *ClusterProfile) deployHelm(testName string) error {
	//nolint
	defer os.Remove(m.cfg.tmpHelmFilePath)
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
