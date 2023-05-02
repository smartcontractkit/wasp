package wasp

import (
	"context"
	"os"
	"strconv"
	"testing"
)

type ProfileGunPart struct {
	Name     string
	Schedule []*Segment
	Gun      Gun
}

type ProfileVUPart struct {
	Name     string
	Schedule []*Segment
	VU       VirtualUser
}

// Profile is a set of concurrent generators forming some workload profile
type Profile struct {
	Generators []*Generator
}

// Run runs all generators and wait until they finish
func (m *Profile) Run(wait bool) error {
	if err := waitSyncGroupReady(); err != nil {
		return err
	}
	for _, g := range m.Generators {
		g.Run(false)
	}
	if wait {
		m.Wait()
	}
	return nil
}

// Wait waits until all generators have finished the workload
func (m *Profile) Wait() {
	for _, g := range m.Generators {
		g.Wait()
	}
}

// NewProfile creates new VU or Gun profile from parts
func NewProfile(t *testing.T, labels map[string]string, parts interface{}) (*Profile, error) {
	gens := make([]*Generator, 0)
	switch parts := parts.(type) {
	case []*ProfileVUPart:
		for _, p := range parts {
			gen, err := NewGenerator(&Config{
				T:          t,
				LoadType:   VU,
				GenName:    p.Name,
				Schedule:   p.Schedule,
				VU:         p.VU,
				Labels:     labels,
				LokiConfig: NewEnvLokiConfig(),
			})
			if err != nil {
				return nil, err
			}
			gens = append(gens, gen)
		}
	case []*ProfileGunPart:
		for _, p := range parts {
			gen, err := NewGenerator(&Config{
				T:          t,
				GenName:    p.Name,
				LoadType:   RPS,
				Schedule:   p.Schedule,
				Gun:        p.Gun,
				Labels:     labels,
				LokiConfig: NewEnvLokiConfig(),
			})
			if err != nil {
				return nil, err
			}
			gens = append(gens, gen)
		}
	default:
		panic("profile parts should be either []*ProfileVUPart or []*ProfileGunPart")
	}
	return &Profile{Generators: gens}, nil
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
