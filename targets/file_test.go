package targets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileTarget(t *testing.T) {

	f, err := os.CreateTemp("", "targets-file-")
	require.Nil(t, err)
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	var data [512]byte
	for i := 0; i < 2000; i++ {
		cnt, err := f.Write(data[0:])
		require.Nil(t, err)
		require.Equal(t, len(data), cnt)
	}

	options := make(Options)
	options["image"] = f.Name()

	target, err := New("file", options)
	require.Nil(t, err)

	TestTarget(t, target)
}
