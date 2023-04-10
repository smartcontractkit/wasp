package wasp

import "testing"

type ProfileGunPart struct {
	Name     string
	Schedule []*Segment
	Gun      Gun
}

type ProfileInstancePart struct {
	Name     string
	Schedule []*Segment
	Instance Instance
}

// Profile is a set of concurrent generators forming some workload profile
type Profile struct {
	Generators []*Generator
}

// Run runs all generators and wait until they finish
func (m *Profile) Run(wait bool) error {
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

// NewRPSProfile creates new RPSProfile from parts
func NewRPSProfile(t *testing.T, labels map[string]string, pp []*ProfileGunPart) (*Profile, error) {
	gens := make([]*Generator, 0)
	for _, p := range pp {
		labels["gen_name"] = p.Name
		gen, err := NewGenerator(&Config{
			T:          t,
			LoadType:   RPSScheduleType,
			Schedule:   p.Schedule,
			Gun:        p.Gun,
			Labels:     labels,
			LokiConfig: NewEnvLokiConfig(),
		})
		if err != nil {
			panic(err)
		}
		gen.Run(false)
		gens = append(gens, gen)
	}
	return &Profile{Generators: gens}, nil
}

// NewInstanceProfile creates new InstanceProfile from parts
func NewInstanceProfile(t *testing.T, labels map[string]string, pp []*ProfileInstancePart) (*Profile, error) {
	gens := make([]*Generator, 0)
	for _, p := range pp {
		labels["gen_name"] = p.Name
		gen, err := NewGenerator(&Config{
			T:          t,
			LoadType:   InstancesScheduleType,
			Schedule:   p.Schedule,
			Instance:   p.Instance,
			Labels:     labels,
			LokiConfig: NewEnvLokiConfig(),
		})
		if err != nil {
			panic(err)
		}
		gen.Run(false)
		gens = append(gens, gen)
	}
	return &Profile{Generators: gens}, nil
}
