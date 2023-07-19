package wasp

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func lokiLogTupleMsg() []interface{} {
	// Loki log entry is a tuple, 9th element is Status, 13th is errors.Error
	logMsg := make([]interface{}, 13)
	for i := 0; i < 13; i++ {
		logMsg = append(logMsg, errors.New("fatal error"))
	}
	return logMsg
}

func TestSmokeLokiNoExitOnStreamError(t *testing.T) {
	t.Parallel()
	ge := os.Getenv("TEST_IGNORE_ERRORS")
	if ge == "" {
		ge = "true"
	}
	ignoreErrors, err := strconv.ParseBool(ge)
	require.NoError(t, err)
	lc, err := NewLokiClient(&LokiConfig{
		IgnoreErrors: ignoreErrors,
	})
	require.NoError(t, err)
	_ = lc.logWrapper.Log(lokiLogTupleMsg()...)
}

func TestSmokeLokiExitOnStreamError(t *testing.T) {
	// must exit 1 if IgnoreErrors = false
	cmd := exec.Command(os.Args[0], "-test.run=TestSmokeLokiNoExitOnStreamError")
	cmd.Env = append(os.Environ(), "TEST_IGNORE_ERRORS=false")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	assert.Equal(t, true, ok)
	assert.Equal(t, "exit status 1", e.Error())
}
