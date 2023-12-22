package client

import (
	"fmt"
	"net"

	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
)

type Client struct {
	adminQueue AdminQueue

	address string
	conn    net.Conn

	hostNQN   string
	targetNQN string

	queues map[uint16]*IOQueue

	log tracer.Tracer
}

func (c *Client) Close() error {
	if c.adminQueue.Queue != nil {
		c.adminQueue.Close()
	}
	for idx := range c.queues {
		c.CloseQueue(idx)
	}
	return nil
}

func (c *Client) CloseQueue(id uint16) error {
	if c.queues[id] != nil {
		err := c.queues[id].Close()
		c.queues[id] = nil
		return err
	}
	return nil
}

func (c *Client) OpenIOQueue(id uint16) (*IOQueue, protocol.NVMEStatusCode) {
	if id == 0 {
		return nil, protocol.SCInvalidQueueId
	}

	if ioq, ok := c.queues[id]; ok {
		return ioq, protocol.SCSuccess
	}

	q, status := c.openQueue(id)
	if status.IsError() {
		return nil, status
	}
	/*
		fc := protocol.ConnectCommand{
			OpCode:    protocol.CapsuleCmdFabric,
			FCType:    protocol.FabricCmdConnect,
			CATTR:     nvme.SQFlowControlDisabled,
			QueueID:   id,
			QueueSize: 32,
		}

		fcd := protocol.ConnectData{
			HostIdentifier: [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			CNTLID:         0xFFFF,
			SubNQN:         c.targetNQN,
			HostNQN:        c.hostNQN,
		}
		fc.SetConnectData(&fcd)

		req := CapsuleRequest{
			Request:  &fc,
			SendData: fc.Data(),
			ready:    make(chan bool),
		}

		q.QueueCapsule(&req)
		req.Wait()

		if req.Response.Status != protocol.SCSuccess {
			q.Close()
			return nil, errors.New("connect failure")
		}
	*/
	ioq := &IOQueue{
		Queue: q,
		ready: make(chan *CapsuleRequest, 128),
	}

	for i := 0; i < 128; i++ {
		ioq.ready <- &CapsuleRequest{
			ready: make(chan bool),
		}
	}
	fmt.Printf("Opened Queue: %d\n", id)

	c.queues[id] = ioq
	return ioq, protocol.SCSuccess
}

func (c *Client) AdminQueue() *AdminQueue {
	return &c.adminQueue
}

func (c *Client) openQueue(id uint16) (*Queue, protocol.NVMEStatusCode) {
	conn, err := net.Dial("tcp", c.address)
	if err != err {
		return nil, protocol.SCConnectionFailure
	}

	q := &Queue{
		id:   id,
		conn: conn,

		requests:     make(map[uint16]*CapsuleRequest),
		requestQueue: make(chan *CapsuleRequest, 32),
		log:          c.log,
	}

	err = q.init()
	if err != nil {
		conn.Close()
		return nil, protocol.SCConnectionFailure
	}

	fc := protocol.ConnectCommand{
		OpCode:    protocol.CapsuleCmdFabric,
		FCType:    protocol.FabricCmdConnect,
		CATTR:     nvme.SQFlowControlDisabled,
		QueueID:   id,
		QueueSize: 32,
	}

	fcd := protocol.ConnectData{
		HostIdentifier: [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		CNTLID:         0xFFFF,
		SubNQN:         c.targetNQN,
		HostNQN:        c.hostNQN,
	}
	fc.SetConnectData(&fcd)

	req := CapsuleRequest{
		Request:  &fc,
		SendData: fc.Data(),
		ready:    make(chan bool),
	}

	q.QueueCapsule(&req)
	req.Wait()

	if req.GetStatus() != protocol.SCSuccess {
		q.Close()
		return nil, req.GetStatus()
	}

	return q, protocol.SCSuccess
}

func (c *Client) Login() protocol.NVMEStatusCode {
	q, status := c.openQueue(0)
	if status.IsError() {
		return status
	}

	c.adminQueue.Queue = q
	return protocol.SCSuccess
}

func (c *Client) WithTracer(t tracer.Tracer) *Client {
	c.log = t
	return c
}

func New(address string, nqn string) (*Client, error) {
	c := &Client{
		address:   address,
		hostNQN:   "nqn.2020-20.com.thirdmartini.nvme:initiator0",
		targetNQN: nqn,
		queues:    make(map[uint16]*IOQueue),
		log:       &tracer.NullTracer{},
	}

	return c, nil
}
