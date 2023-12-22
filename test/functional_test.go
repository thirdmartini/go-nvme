package test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/client"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/targets"
)

const (
	testNQN           = "nqn.2020-20.com.thirdmartini.nvme:null"
	testServerAddress = "localhost:4444"
)

func TestTargetFunctions(t *testing.T) {
	s, err := nvme.New(testServerAddress)
	require.Nil(t, err)
	require.NotNil(t, s)

	options := make(targets.Options).With("size", 1024*1024*1024)

	target, err := targets.New("testable", options)
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
	s.SetDebugLevel(tracer.TraceAll)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.Serve()
		wg.Done()
	}()

	c, err := client.New(testServerAddress, testNQN)
	require.Nil(t, err)
	require.NotNil(t, c)

	c.WithTracer(&tracer.LogTracer{
		Id:    "test-client",
		Level: tracer.TraceAll,
	})

	status := c.Login()
	require.Equal(t, protocol.SCSuccess, status)

	v, err := c.AdminQueue().GetProperty(protocol.RegisterControllerVersion, 1) // 0:32Bit 1:64Bit
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x10300), v)

	v, err = c.AdminQueue().GetProperty(protocol.RegisterControllerStatus, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x0), v)

	v, err = c.AdminQueue().SetProperty(protocol.PropertyControllerConfiguration, 1, 4)
	assert.Nil(t, err)

	v, err = c.AdminQueue().GetProperty(protocol.RegisterControllerStatus, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x1), v)

	id, err := c.AdminQueue().IdentifyController()
	assert.Nil(t, err)
	fmt.Printf("Id: %+v\n", id)

	ioq, status := c.OpenIOQueue(1)
	require.False(t, status.IsError())
	require.NotNil(t, ioq)

	err = ioq.WriteZero(5, 100)
	require.Nil(t, err)

	var zero [512]byte
	var data [512]byte
	var verify [512]byte

	for i := range data {
		data[i] = 0x55
	}

	//IO above lba 1024 returns various errors (see testable)
	err = ioq.Write(5, data[0:])
	require.Nil(t, err)

	err = ioq.Read(5, verify[0:])
	require.Nil(t, err)
	assert.Equal(t, data, verify)

	err = ioq.Read(0, verify[0:])
	require.Nil(t, err)
	assert.Equal(t, zero, verify)

	err = ioq.Trim(5, 100)
	require.Nil(t, err)

	err = ioq.Read(5, verify[0:])
	require.Nil(t, err)
	assert.Equal(t, zero, verify)

	err = ioq.Write(5, data[0:])
	require.Nil(t, err)
	err = ioq.Read(5, verify[0:])
	require.Nil(t, err)
	assert.Equal(t, data, verify)

	err = ioq.WriteZero(0, 100)
	require.Nil(t, err)
	err = ioq.Read(5, verify[0:])
	require.Nil(t, err)
	assert.Equal(t, zero, verify)

	// TODO: do some io here for testing
	err = c.CloseQueue(1)
	assert.Nil(t, err)

	err = c.Close()
	assert.Nil(t, err)

	err = s.Close()
	assert.Nil(t, err)
	wg.Wait()
}

func TestSafeShutdown(t *testing.T) {
	s, err := nvme.New(testServerAddress)
	require.Nil(t, err)
	require.NotNil(t, s)

	options := make(targets.Options).With("size", 1024*1024*1024).With("sleep", 10)

	testable := targets.NewTestableTarget(options)
	require.Nil(t, err)
	require.NotNil(t, testable)

	target := targets.NewWorkQueue(nil, testable)

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
	s.SetDebugLevel(tracer.TraceAll)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.Serve()
		wg.Done()
	}()

	c, err := client.New(testServerAddress, testNQN)
	require.Nil(t, err)
	require.NotNil(t, c)

	c.WithTracer(&tracer.LogTracer{
		Id:    "test-client",
		Level: tracer.TraceAll,
	})

	status := c.Login()
	require.False(t, status.IsError())

	v, err := c.AdminQueue().GetProperty(protocol.RegisterControllerVersion, 1) // 0:32Bit 1:64Bit
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x10300), v)

	v, err = c.AdminQueue().GetProperty(protocol.RegisterControllerStatus, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x0), v)

	v, err = c.AdminQueue().SetProperty(protocol.PropertyControllerConfiguration, 1, 4)
	assert.Nil(t, err)

	v, err = c.AdminQueue().GetProperty(protocol.RegisterControllerStatus, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x1), v)

	id, err := c.AdminQueue().IdentifyController()
	assert.Nil(t, err)
	fmt.Printf("Id: %+v\n", id)

	ioq, status := c.OpenIOQueue(1)
	require.False(t, status.IsError())
	require.NotNil(t, ioq)

	wgio := sync.WaitGroup{}
	wgio.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			err = ioq.WriteZero(5, 100)
			wgio.Done()
		}()
	}

	for {
		if testable.ZeroCount > 10 {
			break
		}
		time.Sleep(time.Millisecond * 5)
	}

	err = s.RemoveSubSystem(testNQN)
	require.Nil(t, err)

	// wait for all the io to finish
	wgio.Wait()
	fmt.Printf("ZC: %d\n", testable.ZeroCount)

}
