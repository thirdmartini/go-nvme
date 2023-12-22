package client

import (
	"sync"
	"testing"

	uuid2 "github.com/google/uuid"
	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/targets"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	testNQN := "nqn.2020-20.com.thirdmartini.nvme:null"

	s, err := nvme.New("localhost:4444")
	require.Nil(t, err)
	require.NotNil(t, s)

	target, err := targets.New("null", nil)
	require.Nil(t, err)
	require.NotNil(t, target)

	err = target.Start()
	require.Nil(t, err)

	subsys := &nvme.TargetSubsystem{
		NQN:    testNQN,
		Target: target,
	}
	uuid, err := uuid2.NewUUID()
	require.Nil(t, err)
	copy(subsys.UUID[:], uuid[:])
	s.AddSubSystem(subsys)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.Serve()
		wg.Done()
	}()

	c, err := New("localhost:4444", testNQN)
	require.Nil(t, err)

	status := c.Login()
	require.False(t, status.IsError())

	err = c.AdminQueue().KeepAlive()
	require.Nil(t, err)

	ioq, status := c.OpenIOQueue(1)
	require.Nil(t, status.AsError())

	data := make([]byte, 512, 512)
	err = ioq.Write(0, data)
	require.Nil(t, err)

	err = c.CloseQueue(1)
	require.Nil(t, err)

	err = c.Close()
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	wg.Wait()
}
