package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/smartcontractkit/wasp"
	"time"
)

type VirtualUser struct {
	target string
	Data   []string
	client *resty.Client
	stop   chan struct{}
}

func NewExampleScenario(target string) *VirtualUser {
	return &VirtualUser{
		target: target,
		client: resty.New().SetBaseURL(target),
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m *VirtualUser) Clone(_ *wasp.Generator) wasp.VirtualUser {
	return &VirtualUser{
		target: m.target,
		client: resty.New().SetBaseURL(m.target),
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m *VirtualUser) Setup(_ *wasp.Generator) error {
	return nil
}

func (m *VirtualUser) Teardown(_ *wasp.Generator) error {
	return nil
}

func (m *VirtualUser) Call(l *wasp.Generator) {
	{
		var result map[string]interface{}
		r, err := m.client.R().
			SetResult(&result).
			Get(m.target)
		if err != nil {
			l.ResponsesChan <- wasp.CallResult{Duration: r.Time(), Data: r.Body()}
			return
		}
		l.ResponsesChan <- wasp.CallResult{Duration: r.Time(), Data: r.Body()}
	}
	time.Sleep(1 * time.Second)
	{
		var result map[string]interface{}
		r, err := m.client.R().
			SetResult(&result).
			Get(m.target)
		if err != nil {
			l.ResponsesChan <- wasp.CallResult{Duration: r.Time(), Data: r.Body()}
			return
		}
		l.ResponsesChan <- wasp.CallResult{Duration: r.Time(), Data: r.Body()}
	}
}

func (m *VirtualUser) Stop(_ *wasp.Generator) {
	m.stop <- struct{}{}
}

func (m *VirtualUser) StopChan() chan struct{} {
	return m.stop
}
