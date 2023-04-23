package wasp

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strings"
)

// ClusterConfig defines k8s jobs settings
type ClusterConfig struct {
	ChartPath  string
	Namespace  string
	HelmValues map[string]string
}

// ClusterProfile is a k8s cluster test for some workload profile
type ClusterProfile struct {
	cfg *ClusterConfig
}

// NewClusterProfile creates new cluster profile
func NewClusterProfile(cfg *ClusterConfig) *ClusterProfile {
	return &ClusterProfile{cfg: cfg}
}

// Run starts a new test
func (m *ClusterProfile) Run() error {
	var cmd strings.Builder
	testName := uuid.NewString()[0:5]
	cmd.WriteString(fmt.Sprintf("helm install wasp-%s %s", testName, m.cfg.ChartPath))
	for k, v := range m.cfg.HelmValues {
		cmd.WriteString(fmt.Sprintf(" --set %s=%s", k, v))
	}
	cmd.WriteString(fmt.Sprintf(" -n %s", m.cfg.Namespace))
	log.Info().Str("Cmd", cmd.String()).Msg("Deploying jobs")
	return ExecCmd(cmd.String())
}
