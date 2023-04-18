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

func NewExampleWSVirtualUser(target string) WSVirtualUser {
	return WSVirtualUser{
		target: target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m WSVirtualUser) Clone(l *wasp.Generator) wasp.VirtualUser {
	return WSVirtualUser{
		target: m.target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m WSVirtualUser) Setup(l *wasp.Generator) error {
	var err error
	m.conn, _, err = websocket.Dial(context.Background(), m.target, &websocket.DialOptions{})
	if err != nil {
		l.Log.Error().Err(err).Msg("failed to connect from vu")
		//nolint
		m.conn.Close(websocket.StatusInternalError, "")
		return err
	}
	return nil
}

func (m WSVirtualUser) Teardown(l *wasp.Generator) error {
	return m.conn.Close(websocket.StatusInternalError, "")
}

func (m WSVirtualUser) Call(l *wasp.Generator) {
	l.ResponsesWaitGroup.Add(1)
	c, _, err := websocket.Dial(context.Background(), m.target, &websocket.DialOptions{})
	if err != nil {
		l.Log.Error().Err(err).Msg("failed to connect from vu")
		//nolint
		c.Close(websocket.StatusInternalError, "")
	}
	go func() {
		defer l.ResponsesWaitGroup.Done()
		for {
			select {
			case <-l.ResponsesCtx.Done():
				//nolint
				c.Close(websocket.StatusNormalClosure, "")
				return
			case <-m.stop:
				return
			default:
				startedAt := time.Now()
				v := map[string]string{}
				err = wsjson.Read(context.Background(), c, &v)
				if err != nil {
					l.Log.Error().Err(err).Msg("failed read ws msg from vu")
				}
				l.ResponsesChan <- wasp.CallResult{StartedAt: &startedAt, Data: v}
			}
		}
	}()
}

func (m WSVirtualUser) Stop(l *wasp.Generator) {
	m.stop <- struct{}{}
}

func (m WSVirtualUser) StopChan() chan struct{} {
	return m.stop
}
