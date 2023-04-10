package main

import (
	"context"
	"time"

	"github.com/smartcontractkit/wasp"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WSInstance struct {
	target string
	Data   []string
	stop   chan struct{}
}

func NewExampleWSInstance(target string) WSInstance {
	return WSInstance{
		target: target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m WSInstance) Clone(l *wasp.Generator) wasp.Instance {
	return WSInstance{
		target: m.target,
		stop:   make(chan struct{}, 1),
		Data:   make([]string, 0),
	}
}

func (m WSInstance) Run(l *wasp.Generator) {
	l.ResponsesWaitGroup.Add(1)
	c, _, err := websocket.Dial(context.Background(), m.target, &websocket.DialOptions{})
	if err != nil {
		l.Log.Error().Err(err).Msg("failed to connect from instanceTemplate")
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
					l.Log.Error().Err(err).Msg("failed read ws msg from instanceTemplate")
				}
				l.ResponsesChan <- wasp.CallResult{StartedAt: &startedAt, Data: v}
			}
		}
	}()
}

func (m WSInstance) Stop(l *wasp.Generator) {
	m.stop <- struct{}{}
}
