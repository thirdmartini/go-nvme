package nvme

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/thirdmartini/go-nvme/internal/buffers"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/stream"
	"github.com/thirdmartini/go-nvme/targets"
)

var controllersAvailable = uint64(0)

type Controller struct {
	SessionID string
	// Controller registers
	REGCtrlCaps   uint64
	REGCtrlConfig uint32

	// Bit.5   Processing Paused
	// Bit.4   Reset Occurred
	// Bit.2-3 Shutdown Status
	// Bit.1   Controller Fatal Status
	// Bit.0   Ready
	REGCtrlStatus    uint32
	Version          uint32
	ConnectedHostNQN string
	ConnectedSubNQN  string
	ControllerID     uint16

	AEC uint32

	NCQR uint16
	NSQR uint16

	// FlowControlDisabled manages whether the controller will perform flow control
	// through use of SQHD/SQCUR.  This is set in the FabricConnect command
	//  Preference: we would prefer this is always disabled
	FlowControlDisabled bool
	SQHD                uint16 // head of queue (next available slot)
	SQCUR               uint16 // current slot being processed

	QueueID   uint16
	QueueSize uint16
	Queue     []NVMERequest

	Server    *Server
	Subsystem Subsystem

	RequestTime  time.Duration
	RequestCount uint64

	conn net.Conn
	in   *stream.Reader
	out  *stream.Writer

	Log tracer.Tracer

	bufferManager *buffers.Buffers

	// queue for handling bottom half
	waiting     chan *NVMERequest
	completions chan *NVMERequest

	// close needs to wait for all sun routines to exit
	wg sync.WaitGroup
}

func (c *Controller) ProcessResponse(w *NVMEResponse) (error, bool) {
	// TODO: we really need to deal with this in the SGL style
	sgl := w.sgl
	offset := uint32(0)
	for idx := 0; idx < w.sglc; idx++ {
		flags := uint8(0)
		if idx+1 == len(sgl) {
			flags = 0x1 << 2 // last transfer
		}

		data := sgl[idx].Data
		dataLen := uint32(len(data))

		w.C2H.CCCID = w.CID
		w.C2H.DATAO = offset
		w.C2H.DATAL = dataLen
		offset += dataLen

		c.Log.TraceProtocol(tracer.TraceData, &w.C2H)

		c.out.MarshalWithData2(protocol.C2HData, flags, &w.C2H, 16, data)
		err := c.out.Flush()
		if err != nil {
			tracer.Fatal(err.Error())
		}
	}

	if c.FlowControlDisabled {
		w.Response.SQHD = 0xFFFF
	}

	c.Log.TraceProtocol(tracer.TraceCapsule, &w.Response)
	if w.NoReply {
		return nil, true
	}

	c.out.Send(protocol.CapsuleResp, &w.Response, 16)
	return c.out.Flush(), true
}

func (c *Controller) Serve(conn net.Conn) error {
	c.conn = conn
	c.in = stream.NewReader(conn)
	c.out = stream.NewWriter(conn)

	quit := make(chan bool, 10)

	c.wg.Add(2)
	go func() {
		err := c.CompletionHandler()
		fmt.Printf("Completion Handler Exiting %v\n", err)
		conn.Close()
		c.wg.Done()
	}()
	defer func() {
		// need to wait for all outstanding commands to drain
		fmt.Printf("Signal Controler(%d:%s/%s).Serve exit\n", c.ControllerID, c.Subsystem.GetNQN(), c.SessionID)

		fmt.Printf("Controler(%d).WaitDrain(%d/%d)\n", c.ControllerID, len(c.waiting), cap(c.waiting))
		count := 0
		for _ = range c.waiting {
			count++
			//fmt.Printf("Controler(%d).Draining %d/%d\n", c.ControllerID, count, cap(c.waiting))
			if count == cap(c.waiting) {
				break
			}
		}
		fmt.Printf("Controler(%d).Serve request queue drained\n", c.ControllerID)
		close(c.completions)
		c.wg.Done()
	}()

	c.bufferManager = buffers.New(128, protocol.MaxH2CPDUSize*8)
	for {
		hdr, err := c.in.Dequeue()
		if err != nil {
			if err != io.EOF {
				c.Log.Trace(0x01, "Session Error: %s", err.Error())
			}
			break
		}

		stime := time.Now()
		c.Log.TraceProtocol(tracer.TraceCommands, hdr)
		switch hdr.Type {
		case protocol.ICReq:
			req := protocol.ICRequest{}

			err = c.in.Receive(&req)
			if err != nil {
				c.Log.Trace(tracer.TraceCommands, "Session Error: %s", err.Error())
				return err
			}
			c.Log.TraceProtocol(tracer.TraceCommands, &req)

			rsp := protocol.ICResponse{
				PDUDataDigest:    0,      // no digests for now
				MaxH2CDataLength: 0x8000, // 32K
			}

			err = c.out.Send(protocol.ICResp, &rsp, 120)
			if err != nil {
				c.Log.Trace(tracer.TraceCommands, "Session Error: %s", err.Error())
				return err
			}

			c.Log.TraceProtocol(tracer.TraceCommands, &rsp)
			err = c.out.Flush()
			if err != nil {
				c.Log.Trace(tracer.TraceCommands, "Session Error: %s", err.Error())
				return err
			}

		case protocol.CapsuleCmd:
			err = c.HandleCapsule(conn, quit)
			if err != nil {
				c.Log.Trace(tracer.TraceCommands, "Session Error: %s", err.Error())
				return err
			}

		case protocol.H2CTermReq:
			term := protocol.C2HTermRequest{}
			err = c.in.Receive(&term)
			if err != nil {
				return err
			}
			return fmt.Errorf("controller terminated connection: %x:%x", term.FatalErrorStatus, term.FatalErrorInformation)

		default:
			c.Log.Todo("Unknown h.Type: 0x%x", hdr.Type)
			break
		}
		c.RequestTime += time.Now().Sub(stime)
		c.RequestCount++
		c.Log.End()
	}
	return nil
}

func (c *Controller) HandleCapsule(conn net.Conn, quit chan bool) error {
	var req *NVMERequest

	select {
	case <-quit:
		return nil
	case req = <-c.waiting:
	}
	c.SQHD++
	if c.SQHD > c.QueueSize {
		c.SQHD = 0
	}

	req.completion = c.completions
	err := c.in.Receive(&req.capsule)
	if err != nil {
		c.Log.Trace(0x01, "Session Error: %s", err.Error())
		return err
	}

	dataLen := c.in.Length()
	if dataLen != 0 {
		req.payload = c.bufferManager.Get()
		req.payloadLength = int(dataLen)
		c.in.ReceiveData(req.payload[0:dataLen])
	}

	// we can do all this in dequeue
	w := &req.response
	w.Response.CID = req.capsule.CID
	w.Response.SQHD = c.SQHD
	w.SetStatus(protocol.SCSuccess)
	w.State = 0

	capsule := req.Capsule()
	w.Type = protocol.CapsuleResp
	w.CID = capsule.CID

	// QueueID:0 is reserved for admin commands
	if c.QueueID == 0 {
		c.Log.TraceCapsule(true, capsule)
		err = c.handleAdminCapsule(w, req)
		req.Complete(targets.TargetErrorNone)
	} else {
		c.Log.TraceCapsule(false, capsule)
		err = c.handleIOCapsule(w, req)
	}

	return err
}

// CompletionHandler sends the request responses back to the initiator
func (c *Controller) CompletionHandler() error {
	var req *NVMERequest

	for req = range c.completions {
		err, done := c.ProcessResponse(&req.response)
		if err != nil {
			fmt.Printf("CompletionError: %s\n", err.Error())
			if req.payload != nil {
				c.bufferManager.Put(req.payload)
				req.payload = nil
			}
			req.active = false
			req.response.Reset()
			c.waiting <- req
			continue
		}
		if done {
			if req.payload != nil {
				c.bufferManager.Put(req.payload)
				req.payload = nil
			}
			req.active = false
			req.response.Reset()
			c.waiting <- req
		} // else we are waiting for an R2T to requeue this command
	}
	return nil
}

func (c *Controller) Close() error {
	c.conn.Close()
	c.wg.Wait()
	return nil
}
