package main

import (
	"context"
	"time"

	"github.com/smartcontractkit/wasp"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WSVirtualUser struct {
	target string
	conn   *websocket.Conn
	Data   []string
	stop   chan struct{}
}

func NewExampleWSVirtualUser(target string) *WSVirtualUser {
	return &WSVirtualUser{
		target: target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m *WSVirtualUser) Clone(_ *wasp.Generator) wasp.VirtualUser {
	return &WSVirtualUser{
		target: m.target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m *WSVirtualUser) Setup(l *wasp.Generator) error {
	var err error
	m.conn, _, err = websocket.Dial(context.Background(), m.target, &websocket.DialOptions{})
	if err != nil {
		l.Log.Error().Err(err).Msg("failed to connect from vu")
		//nolint
		_ = m.conn.Close(websocket.StatusInternalError, "")
		return err
	}
	return nil
}

func (m *WSVirtualUser) Teardown(_ *wasp.Generator) error {
	return m.conn.Close(websocket.StatusInternalError, "")
}

func (m *WSVirtualUser) Call(l *wasp.Generator) {
	startedAt := time.Now()
	v := map[string]string{}
	err := wsjson.Read(context.Background(), m.conn, &v)
	if err != nil {
		l.Log.Error().Err(err).Msg("failed read ws msg from vu")
	}
	l.ResponsesChan <- wasp.CallResult{StartedAt: &startedAt, Data: v}
}

func (m *WSVirtualUser) Stop(_ *wasp.Generator) {
	m.stop <- struct{}{}
}

func (m *WSVirtualUser) StopChan() chan struct{} {
	return m.stop
}
