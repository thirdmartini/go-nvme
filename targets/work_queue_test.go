package targets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkQueue(t *testing.T) {
	ioCount := 16

	testTarget := NewTestableTarget(nil)
	wqTarget := NewWorkQueue(nil, testTarget)

	err := wqTarget.Start()
	require.Nil(t, err)

	c := TestCompleter{}
	c.Wait.Add(ioCount)

	for i := 0; i < ioCount; i++ {
		r := &IORequest{
			Command:         IORequestCmdRead,
			Lba:             uint64(i),
			Length:          1,
			CompleteRequest: c.Complete,
		}
		s := wqTarget.Queue(r)
		assert.Equal(t, TargetErrorNone, s)
	}
	c.Wait.Wait()
	assert.Equal(t, uint64(ioCount), testTarget.ReadCount)
	assert.Equal(t, uint64(0), testTarget.WriteCount)

	err = wqTarget.Close()
	assert.Nil(t, err)
}
