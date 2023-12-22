package targets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTestableTarget(t *testing.T) {
	options := make(Options)
	options["size"] = "10485760"

	target, err := New("testable", options)
	require.Nil(t, err)

	require.Equal(t, uint64(defaultTestableTargetSize), target.GetSize())

	TestTarget(t, target)
}
