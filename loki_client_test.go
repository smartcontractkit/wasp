package wasp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestLokiProcessStreamErrors(t *testing.T) {
	t.Parallel()
	ge := os.Getenv("TEST_MAX_ERRORS")
	maxErrors, err := strconv.ParseInt(ge, 10, 64)
	require.NoError(t, err)
	lc, err := NewLokiClient(&LokiConfig{
		MaxErrors: int(maxErrors),
	})
	require.NoError(t, err)
	_ = lc.logWrapper.Log(lokiLogTupleMsg()...)
}

func TestSmokeLokiExitOnStreamErrors(t *testing.T) {
	type testcase struct {
		name      string
		maxErrors int
		mustExit  bool
	}

	tests := []testcase{
		{
			name:      "must exit with exit code 0, ignoring errors",
			maxErrors: 0,
		},
		{
			name:      "must exit with exit code 0, we have only 1 error happened",
			maxErrors: 2,
		},
		{
			name:      "must exit with exit code 1",
			mustExit:  true,
			maxErrors: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestLokiProcessStreamErrors")
			cmd.Env = append(os.Environ(), fmt.Sprintf("TEST_MAX_ERRORS=%d", tc.maxErrors))
			err := cmd.Run()
			if tc.mustExit {
				var e *exec.ExitError
				ok := errors.As(err, &e)
				assert.Equal(t, true, ok)
				assert.Equal(t, "exit status 1", e.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
