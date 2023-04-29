package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/smartcontractkit/wasp"
)

const (
	GroupAuth = "auth"
	GroupUser = "user"
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

func (m *VirtualUser) requestOne(l *wasp.Generator) {
	var result map[string]interface{}
	r, err := m.client.R().
		SetResult(&result).
		Get(m.target)
	if err != nil {
		l.Responses.Err(r, GroupAuth, err)
		return
	}
	l.Responses.OK(r, GroupAuth)
}

func (m *VirtualUser) requestTwo(l *wasp.Generator) {
	var result map[string]interface{}
	r, err := m.client.R().
		SetResult(&result).
		Get(m.target)
	if err != nil {
		l.Responses.Err(r, GroupUser, err)
		return
	}
	l.Responses.OK(r, GroupUser)
}

func (m *VirtualUser) Call(l *wasp.Generator) {
	m.requestOne(l)
	m.requestTwo(l)
}

func (m *VirtualUser) Stop(_ *wasp.Generator) {
	m.stop <- struct{}{}
}

func (m *VirtualUser) StopChan() chan struct{} {
	return m.stop
}
