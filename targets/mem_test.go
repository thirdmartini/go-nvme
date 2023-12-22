package targets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemTarget(t *testing.T) {
	options := make(Options)
	options["size"] = "10485760"

	target, err := New("mem", options)
	require.Nil(t, err)

	require.Equal(t, uint64(10485760), target.GetSize())

	TestTarget(t, target)
}
