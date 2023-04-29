package wasp

import (
	"github.com/go-resty/resty/v2"
)

/* handy wrappers to use with resty in scenario (VU) tests */

const (
	CallGroupLabel = "call_group"
)

type Responses struct {
	ch chan CallResult
}

func NewResponses(ch chan CallResult) *Responses {
	return &Responses{ch}
}

func (m *Responses) OK(r *resty.Response, group string) {
	m.ch <- CallResult{
		Duration: r.Time(),
		Group:    group,
		Data:     r.Body(),
	}
}

func (m *Responses) Err(r *resty.Response, group string, err error) {
	m.ch <- CallResult{
		Failed:   true,
		Error:    err.Error(),
		Duration: r.Time(),
		Group:    group,
		Data:     r.Body(),
	}
}
