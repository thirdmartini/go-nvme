package client

import (
	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/protocol"
)

// AdminQueue provides the abstraction to an NVME targets AdminQueue
type AdminQueue struct {
	*Queue
}

// KeepAlive sends a keepalive to the target
func (q *AdminQueue) KeepAlive() error {
	req := CapsuleRequest{
		Request: &protocol.CapsuleCommand{
			OpCode: protocol.CapsuleCmdKeepAlive,
		},
		ready: make(chan bool),
	}
	q.QueueCapsule(&req)
	req.Wait()
	return req.GetStatus().AsError()
}

func (q *AdminQueue) GetProperty(property uint32, sz uint8) (uint64, error) {
	req := CapsuleRequest{
		Request: &protocol.CapsuleCommand{
			OpCode: protocol.CapsuleCmdFabric,
			FCType: protocol.FabricCmdPropertyGet,
			D11:    property,
			D10:    uint32(sz),
		},
		ready: make(chan bool),
	}
	q.QueueCapsule(&req)
	req.Wait()

	err := req.GetStatus().AsError()
	if err != nil {
		return 0, err
	}

	val := uint64(0)
	sm := serialize.NewDeserializer(req.Response.FabricResponse[:])
	err = sm.Deserialize(&val)
	return val, req.GetStatus().AsError()
}

func (q *AdminQueue) SetProperty(property uint32, value uint64, sz uint8) (uint64, error) {
	req := CapsuleRequest{
		Request: &protocol.CapsuleCommand{
			OpCode: protocol.CapsuleCmdFabric,
			FCType: protocol.FabricCmdPropertySet,
			D11:    property,
			D10:    uint32(sz),
			D12:    uint32(value & 0xFFFFFFFF), // fixme need to to big->little
			D13:    uint32(value >> 32),
		},
		ready: make(chan bool),
	}

	q.QueueCapsule(&req)
	req.Wait()

	err := req.GetStatus().AsError()
	if err != nil {
		return 0, err
	}

	val := uint64(0)
	sm := serialize.NewDeserializer(req.Response.FabricResponse[:])
	err = sm.Deserialize(&val)
	return val, req.GetStatus().AsError()
}

func (q *AdminQueue) IdentifyController() (protocol.IdentifyController, error) {
	req := CapsuleRequest{
		Request: &protocol.CapsuleCommand{
			OpCode: protocol.CapsuleCmdIdentify,
			D10:    0x01,
		},
		ready:    make(chan bool),
		RecvData: make([]byte, 4096, 4096),
	}

	q.QueueCapsule(&req)
	req.Wait()

	id := protocol.IdentifyController{}
	err := req.GetStatus().AsError()
	if err != nil {
		return id, err
	}

	return id, serialize.NewDeserializer(req.RecvData).Deserialize(&id)
}
