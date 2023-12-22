package targets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSleepyTarget(t *testing.T) {
	target, err := New("sleepy", nil)
	require.Nil(t, err)
	TestTarget(t, target)
}
