package client

import (
	"fmt"

	"github.com/thirdmartini/go-nvme/protocol"
)

// IOQueue exports io queue functionality from the client
type IOQueue struct {
	*Queue
	ready chan *CapsuleRequest
}

func (q *IOQueue) Write(lba uint64, data []byte) error {
	req := <-q.ready
	req.Request = &req.capsule
	req.capsule.OpCode = protocol.CapsuleCmdWrite
	req.capsule.D10 = uint32(lba)
	req.capsule.D11 = uint32(lba >> 32)
	req.capsule.D12 = (uint32(len(data)) / 512) - 1
	req.SendData = data
	q.QueueCapsule(req)
	req.Wait()
	req.SendData = nil
	q.ready <- req

	return req.GetStatus().AsError()
}

func (q *IOQueue) WriteZero(lba uint64, count uint16) error {
	if count == 0 {
		return fmt.Errorf("invalid parameter")
	}

	req := <-q.ready
	req.Request = &req.capsule
	req.capsule.OpCode = protocol.CapsuleCmdWriteZeros
	req.capsule.D10 = uint32(lba)
	req.capsule.D11 = uint32(lba >> 32)
	req.capsule.D12 = uint32(count - 1)
	q.QueueCapsule(req)
	req.Wait()
	req.SendData = nil
	q.ready <- req
	return req.GetStatus().AsError()
}

func (q *IOQueue) Trim(lba uint64, count uint16) error {
	if count == 0 {
		return fmt.Errorf("invalid parameter")
	}
	req := <-q.ready
	req.Request = &req.capsule
	req.capsule.OpCode = protocol.CapsuleCmdWriteZeros
	req.capsule.D10 = uint32(lba)
	req.capsule.D11 = uint32(lba >> 32)
	req.capsule.D12 = uint32(count-1) | protocol.CommandBitDeallocateSet
	q.QueueCapsule(req)
	req.Wait()
	req.SendData = nil
	q.ready <- req
	return req.GetStatus().AsError()
}

func (q *IOQueue) Read(lba uint64, data []byte) error {
	req := <-q.ready

	req.Request = &req.capsule
	req.capsule.OpCode = protocol.CapsuleCmdRead
	req.capsule.D10 = uint32(lba)
	req.capsule.D11 = uint32(lba >> 32)
	req.capsule.D12 = (uint32(len(data)) / 512) - 1
	req.RecvData = data
	q.QueueCapsule(req)
	req.Wait()
	req.RecvData = nil
	q.ready <- req
	return req.GetStatus().AsError()
}

func (q *IOQueue) Flush(lba uint64, data []byte) error {
	req := <-q.ready
	req.Request = &req.capsule
	req.capsule.OpCode = protocol.CapsuleCmdFlush
	q.QueueCapsule(req)
	req.Wait()
	q.ready <- req
	return req.GetStatus().AsError()
}
