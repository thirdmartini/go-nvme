package client

import (
	"fmt"
	"net"
	"sync"

	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/stream"
)

type Command interface {
	protocol.PDU
	SetCID(id uint16)
}

type CapsuleRequest struct {
	Request  Command
	Response protocol.CapsuleResponse
	SendData []byte
	RecvData []byte

	// capsule for use
	capsule protocol.CapsuleCommand
	ready   chan bool
}

func (c CapsuleRequest) SetStatus(code protocol.NVMEStatusCode) {
	c.Response.SetStatus(code)
}

func (c CapsuleRequest) GetStatus() protocol.NVMEStatusCode {
	return c.Response.GetStatus()
}

func (c *CapsuleRequest) Done() {
	c.ready <- true
}

func (c *CapsuleRequest) Wait() {
	<-c.ready
}

func NewCapsuleRequest(c Command, r, w []byte) *CapsuleRequest {
	return &CapsuleRequest{
		Request:  c,
		ready:    make(chan bool),
		SendData: w,
		RecvData: r,
	}
}

type Queue struct {
	id   uint16
	conn net.Conn

	cid          uint16
	lock         sync.Mutex
	requests     map[uint16]*CapsuleRequest
	requestQueue chan *CapsuleRequest
	freeQueue    protocol.CapsuleCommand

	// protocol control
	ic  protocol.ICResponse
	log tracer.Tracer
}

func (c *Queue) Close() error {
	close(c.requestQueue)
	return c.conn.Close()
}

func (c *Queue) receiver(wg *sync.WaitGroup) error {
	capResponse := protocol.CapsuleResponse{}
	c2Data := protocol.C2HDataTransfer{}

	defer func() {
		fmt.Printf("receiver terminating\n")
		c.conn.Close()

		c.lock.Lock()
		defer c.lock.Unlock()
		for k := range c.requests {
			r := c.requests[k]
			delete(c.requests, k)
			r.SetStatus(protocol.SCInternalError)
			r.Done()
		}
		fmt.Printf("receiver termination done\n")
	}()

	defer wg.Done()
	stream := stream.NewReader(c.conn)
	for {
		hdr, err := stream.Dequeue()
		if err != nil {
			return err
		}

		c.log.TraceProtocol(tracer.TraceCommands, hdr)
		switch hdr.Type {
		case protocol.CapsuleResp:
			err = stream.Receive(&capResponse)
			if err != nil {
				return err
			}

			//c.log.Trace(nvme.TraceCapsule,"%s", &capResponse)
			c.lock.Lock()
			r, ok := c.requests[capResponse.CID]
			if !ok {
				tracer.Fatal("bad cid: %d", capResponse.CID)
			}
			delete(c.requests, capResponse.CID)
			c.lock.Unlock()

			r.Response = capResponse
			r.Done()

		case protocol.C2HData:
			err = stream.Receive(&c2Data)
			if err != nil {
				return err
			}

			c.log.TraceProtocol(tracer.TraceData, &c2Data)

			c.lock.Lock()
			r, ok := c.requests[c2Data.CCCID]
			if !ok {
				tracer.Fatal("bad cid: %d", c2Data.CCCID)

			}
			c.lock.Unlock()
			stream.ReceiveData(r.RecvData[c2Data.DATAO:])

		case protocol.C2HTermReq:
			term := protocol.C2HTermRequest{}
			err = stream.Receive(&term)
			if err != nil {
				return err
			}
			return fmt.Errorf("controller terminated connection: %x:%x", term.FatalErrorStatus, term.FatalErrorInformation)

		default:
			tracer.Fatal("got bad code: %+v\n", hdr)
		}
		c.log.End()
	}
}

func (c *Queue) transmitter(wg *sync.WaitGroup) {
	defer wg.Done()

	out := stream.NewWriter(c.conn)
	for cmd := range c.requestQueue {
		c.lock.Lock()
		cmd.Request.SetCID(c.cid)
		c.requests[c.cid] = cmd
		c.cid++
		c.lock.Unlock()

		err := out.MarshalWithData2(protocol.CapsuleCmd, 0, cmd.Request, 64, cmd.SendData)
		if err != nil {
			tracer.Fatal("failed to marshal data")
		}

		c.log.TraceProtocol(tracer.TraceCapsule, cmd.Request)
		err = out.Flush()
		if err != nil {
			fmt.Printf("client error: %s\n", err.Error())
			c.lock.Lock()
			delete(c.requests, c.cid)
			c.lock.Unlock()
			cmd.SetStatus(protocol.SCInternalError) // need to indicate connection error
			cmd.Done()
		}
	}
}

func (c *Queue) QueueCapsule(cap *CapsuleRequest) error {
	//	cap.wg.Add(1)
	c.requestQueue <- cap
	return nil
}

func (c *Queue) SendH2C(h2c *protocol.H2CTermRequest) error {
	out := stream.NewWriter(c.conn)
	out.Send(protocol.ICReq, h2c, 16)
	return out.Flush()
}

func (c *Queue) init() error {
	out := stream.NewWriter(c.conn)
	in := stream.NewReader(c.conn)

	req := protocol.ICRequest{}

	out.Send(protocol.ICReq, &req, 120)

	err := out.Flush()
	if err != nil {
		return err
	}

	hdr, err := in.Dequeue()
	if err != nil {
		return err
	}

	if hdr.Type != protocol.ICResp {
		return nil
	}

	err = in.Receive(&c.ic)
	if err != nil {
		return err
	}

	if c.ic.MaxH2CDataLength != 0x8000 {
		fmt.Printf("%x\n", c.ic.MaxH2CDataLength)
		tracer.Fatal("Not our target")
	}

	// start monitor func that cleans up any pending io
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go c.receiver(&wg)
		go c.transmitter(&wg)
		wg.Wait()

		// cleanup stuck requests on exit
		c.lock.Lock()
		for idx := range c.requests {
			r := c.requests[idx]

			r.SetStatus(protocol.SCInternalError) // need to indicate connection error
			delete(c.requests, idx)
		}
		c.lock.Unlock()
	}()
	return nil
}
