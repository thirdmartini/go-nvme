package targets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullTarget(t *testing.T) {
	target, err := New("null", nil)
	require.Nil(t, err)
	TestTarget(t, target)
}
