package wasp

import (
	"math/rand"
	"time"
)

// MockVirtualUserConfig configures a mock virtual user
type MockVirtualUserConfig struct {
	// FailRatio in percentage, 0-100
	FailRatio int
	// TimeoutRatio in percentage, 0-100
	TimeoutRatio int
	// CallSleep time spent waiting inside a call
	CallSleep time.Duration
}

// MockVirtualUser is a mock virtual user
type MockVirtualUser struct {
	cfg  *MockVirtualUserConfig
	stop chan struct{}
	Data []string
}

// NewMockVU create a mock virtual user
func NewMockVU(cfg *MockVirtualUserConfig) *MockVirtualUser {
	return &MockVirtualUser{
		cfg:  cfg,
		stop: make(chan struct{}, 1),
		Data: make([]string, 0),
	}
}

func (m *MockVirtualUser) Clone(_ *Generator) VirtualUser {
	return &MockVirtualUser{
		cfg:  m.cfg,
		stop: make(chan struct{}, 1),
		Data: make([]string, 0),
	}
}

func (m *MockVirtualUser) Setup(_ *Generator) error {
	return nil
}

func (m *MockVirtualUser) Teardown(_ *Generator) error {
	return nil
}

func (m *MockVirtualUser) Call(l *Generator) {
	startedAt := time.Now()
	time.Sleep(m.cfg.CallSleep)
	if m.cfg.FailRatio > 0 && m.cfg.FailRatio <= 100 {
		//nolint
		r := rand.Intn(100)
		if r <= m.cfg.FailRatio {
			l.ResponsesChan <- &CallResult{StartedAt: &startedAt, Data: "failedCallData", Error: "error", Failed: true}
		}
	}
	if m.cfg.TimeoutRatio > 0 && m.cfg.TimeoutRatio <= 100 {
		//nolint
		r := rand.Intn(100)
		if r <= m.cfg.TimeoutRatio {
			time.Sleep(m.cfg.CallSleep + 20*time.Millisecond)
		}
	}
	l.ResponsesChan <- &CallResult{StartedAt: &startedAt, Data: "successCallData"}
}

func (m *MockVirtualUser) Stop(_ *Generator) {
	m.stop <- struct{}{}
}

func (m *MockVirtualUser) StopChan() chan struct{} {
	return m.stop
}
