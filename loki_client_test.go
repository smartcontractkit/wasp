package wasp

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func lokiLogTupleMsg() []interface{} {
	// Loki log entry is a tuple, 9th element is Status, 13th is errors.Error
	logMsg := make([]interface{}, 13)
	for i := 0; i < 13; i++ {
		logMsg = append(logMsg, errors.New("fatal error"))
	}
	return logMsg
}

func TestSmokeDefaultLokiConfig(t *testing.T) {
	cfg := NewEnvLokiConfig()
	cfg.MaxErrors = 1
	lc, err := NewLokiClient(cfg)
	defer lc.StopNow()
	require.NoError(t, err)
	q := struct {
		Name string
	}{
		Name: "test",
	}
	var errAtSomePoint error
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		errAtSomePoint = lc.HandleStruct(nil, time.Now(), q)
		if errAtSomePoint != nil {
			return
		}
	}
	require.Error(t, err)
}

func TestSmokeLokiErrors(t *testing.T) {
	type testcase struct {
		name      string
		maxErrors int
		mustError bool
	}

	tests := []testcase{
		{
			name:      "must ignore all the errors",
			maxErrors: -1,
		},
		{
			name:      "must continue, but log errors",
			maxErrors: 2,
		},
		{
			name:      "must return an error",
			mustError: true,
			maxErrors: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lc, err := NewLokiClient(&LokiConfig{
				MaxErrors: tc.maxErrors,
			})
			defer lc.StopNow()
			require.NoError(t, err)
			_ = lc.logWrapper.Log(lokiLogTupleMsg()...)
			_ = lc.logWrapper.Log(lokiLogTupleMsg()...)
			q := struct {
				Name string
			}{
				Name: "test",
			}
			err = lc.HandleStruct(nil, time.Now(), q)
			if tc.mustError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
