package wasp

import (
	"context"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// WSMockVUConfig ws mock config
type WSMockVUConfig struct {
	TargetURl string
}

// WSMockVU ws mock virtual user
type WSMockVU struct {
	cfg  *WSMockVUConfig
	conn *websocket.Conn
	stop chan struct{}
	Data []string
}

// NewWSMockVU create a ws mock virtual user
func NewWSMockVU(cfg *WSMockVUConfig) *WSMockVU {
	return &WSMockVU{
		cfg:  cfg,
		stop: make(chan struct{}, 1),
		Data: make([]string, 0),
	}
}

func (m *WSMockVU) Clone(_ *Generator) VirtualUser {
	return &WSMockVU{
		cfg:  m.cfg,
		stop: make(chan struct{}, 1),
		Data: make([]string, 0),
	}
}

func (m *WSMockVU) Setup(l *Generator) error {
	var err error
	m.conn, _, err = websocket.Dial(context.Background(), m.cfg.TargetURl, &websocket.DialOptions{})
	if err != nil {
		l.Log.Error().Err(err).Msg("failed to connect from virtual user")
		//nolint
		_ = m.conn.Close(websocket.StatusInternalError, "")
		return err
	}
	return nil
}

func (m *WSMockVU) Teardown(_ *Generator) error {
	return m.conn.Close(websocket.StatusInternalError, "")
}

// Call create a virtual user firing read requests against mock ws server
func (m *WSMockVU) Call(l *Generator) {
	startedAt := time.Now()
	v := map[string]string{}
	err := wsjson.Read(context.Background(), m.conn, &v)
	if err != nil {
		l.Log.Error().Err(err).Msg("failed read ws msg from vu")
	}
	l.ResponsesChan <- &CallResult{StartedAt: &startedAt, Data: v}
}

func (m *WSMockVU) Stop(_ *Generator) {
	m.stop <- struct{}{}
}

func (m *WSMockVU) StopChan() chan struct{} {
	return m.stop
}
