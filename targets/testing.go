package targets

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCompleter struct {
	status TargetError
	Wait   sync.WaitGroup
}

func (t *TestCompleter) Complete(status TargetError) {
	t.status = status
	t.Wait.Done()
}

func TestRequest(t Target, r *IORequest) TargetError {
	c := TestCompleter{}
	c.Wait.Add(1)
	r.CompleteRequest = c.Complete
	t.Queue(r)
	c.Wait.Wait()
	return c.status
}

func TestTarget(t *testing.T, target Target) {
	err := target.Start()
	require.Nil(t, err)

	sz := target.GetSize()
	require.NotEqual(t, uint64(0), sz)

	status := TestRequest(target, &IORequest{
		Command: IORequestCmdFlush,
	})
	require.Equal(t, TargetErrorNone, status)
}
