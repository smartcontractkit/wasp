package wasp

import (
	"context"
	"os"
	"strconv"
)

// Profile is a set of concurrent generators forming some workload profile
type Profile struct {
	Generators   []*Generator
	bootstrapErr error
}

// Run runs all generators and wait until they finish
func (m *Profile) Run(wait bool) (*Profile, error) {
	if m.bootstrapErr != nil {
		return m, m.bootstrapErr
	}
	if err := waitSyncGroupReady(); err != nil {
		return m, err
	}
	for _, g := range m.Generators {
		g.Run(false)
	}
	if wait {
		m.Wait()
	}
	return m, nil
}

// Wait waits until all generators have finished the workload
func (m *Profile) Wait() {
	for _, g := range m.Generators {
		g.Wait()
	}
}

// NewProfile creates new VU or Gun profile from parts
func NewProfile() *Profile {
	return &Profile{Generators: make([]*Generator, 0)}
}

func (m *Profile) Add(g *Generator, err error) *Profile {
	if err != nil {
		m.bootstrapErr = err
		return m
	}
	m.Generators = append(m.Generators, g)
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
