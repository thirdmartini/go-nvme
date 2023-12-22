package nvme

import (
	"fmt"

	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
)

const (
	SQFlowControlDisabled = 1 << 2
)

func (c *Controller) handleFabricCommand(w *NVMEResponse, r *NVMERequest) error {
	capsule := r.Capsule()

	c.Log.Trace(tracer.TraceFabric, "%s", capsule)

	switch capsule.FCType {
	// 3.3 Connect Command and Response
	case protocol.FabricCmdConnect:
		// When fabric connects, it should always connect with controller set to 0xFFFF and QID set to 0
		//   This is the admin queue
		//   we will then be probed for setings and configuration and will get a connection for
		//   a specific controller and Queue ID. ...
		//    this wil then be an IO connection

		// reinterpret the command
		fcc := protocol.ConnectCommand{}
		if r.ReinterpretCapsule(&fcc) != nil {
			tracer.Fatal("reinterpret error")
		}

		ds := serialize.NewDeserializer(r.Payload())
		fcd := protocol.ConnectData{}
		if err := ds.Deserialize(&fcd); err != nil {
			tracer.Fatal("deserialize error:" + err.Error())
		}

		//utilPrettyPrint(fcc)
		c.Log.Trace(tracer.TraceCapsuleDetail, "      SQSIZE: %d", fcc.QueueSize+1) // zero based
		c.Log.Trace(tracer.TraceCapsuleDetail, "       CATTR: %x", fcc.CATTR)
		c.Log.Trace(tracer.TraceCapsuleDetail, "        KATO: %d", fcc.KATO)
		c.Log.Trace(tracer.TraceCapsuleDetail, "  CNTLID/QID: 0x%x (%d) / %d", fcd.CNTLID, fcd.CNTLID, fcc.QueueID)
		c.Log.Trace(tracer.TraceCapsuleDetail, "    HOST NQN: %s", fcd.HostNQN)
		c.Log.Trace(tracer.TraceCapsuleDetail, "     SUB NQN: %s", fcd.SubNQN)

		c.ConnectedHostNQN = fcd.HostNQN
		c.ConnectedSubNQN = fcd.SubNQN

		// TODO: rework this, we will alias ourselves to this queue now
		//  if queue is 0 its ALWAYS the admin queue
		c.QueueID = fcc.QueueID

		// did not find the subsystem we need
		subsys := c.Server.GetSubSystem(c.ConnectedSubNQN)
		if subsys == nil {
			offset := uint64(256) | 0x10000 // offset of NQN in request + set the start of data
			w.SetStatus(protocol.SCNamespaceNotReady)
			sm := serialize.New(w.Response.FabricResponse[0:])
			sm.Serialize(&offset)
			return nil
		}

		//  New
		c.Subsystem = subsys

		// we can resize past our controller limits!
		if int(fcc.QueueSize+1) > len(c.Queue) {
			tracer.Fatal("protocol.FabricCmdConnect: cant resize past queue size: %+v", capsule)
		}
		c.QueueSize = fcc.QueueSize + 1

		// Controller ID (CNTLID): Specifies the controller ID allocated to the host. If a
		// particular controller was specified in the CNTLID field of the Connect command,
		// then this field shall contain the same value
		if fcd.CNTLID == 0xFFFF {
			if c.ControllerID == 0 {
				c.ControllerID = 7
				//atomic.AddInt32(controllersAvailable)
				//controllersAvailable++
				val := uint32(c.ControllerID)
				sm := serialize.New(w.Response.FabricResponse[:])
				sm.Serialize(&val)
				fmt.Printf("C: %d  FCR: %+v\n", c.ControllerID, w.Response.FabricResponse)
			}
		} else {
			// FIXME: this may not be the correct thing to do
			c.ControllerID = fcd.CNTLID
		}

		if fcc.CATTR&SQFlowControlDisabled == SQFlowControlDisabled {
			c.FlowControlDisabled = true
		}

		w.Response.QueueID = c.QueueID
		// w.Response.QueueID = 0

	// 3.5 Property Get Command and Response
	case protocol.FabricCmdPropertyGet:
		// D10 contains the size: 000:4bytes | 001:8bytes
		c.Log.Trace(tracer.TraceCommands, "     Property: 0x%0x (Size:%x)", capsule.D11, capsule.D10&0x7)

		val := uint64(0)
		switch capsule.D11 {
		case protocol.PropertyControllerCapabilities: // Controller Capabilities Register
			val = c.REGCtrlCaps

		case protocol.PropertyVersion: // Controller Version Register
			val = uint64(c.Version)

		case protocol.PropertyControllerStatus: // Controller Status Register (0x1c)
			c.Log.Trace(tracer.TraceAll, "Get: protocol.PropertyControllerStatus: 0x%x\n", c.REGCtrlStatus)
			val = uint64(c.REGCtrlStatus)

		case protocol.PropertyControllerConfiguration: // Controller Configuration
			val = uint64(c.REGCtrlConfig)

		case protocol.PropertySubsystemReset: // RESET
			// FIXME: pretend we reset

		default:
			tracer.Fatal("protocol.FabricCmdPropertyGet %+v", capsule)
		}

		sm := serialize.New(w.Response.FabricResponse[:])
		sm.Serialize(&val)

	case protocol.FabricCmdPropertySet:
		c.Log.Trace(tracer.TraceCapsuleDetail, "     Property: 0x%0x   Value: 0x%x", capsule.D11, capsule.D12)

		switch capsule.D11 {
		case protocol.PropertyControllerConfiguration: // Controller Configuration

			// bits 04:06  000 (NVM Command Set) | 111 Admin Command Set
			// bits 07:10 MPS (2 ^ (12 + MPS)).
			// RW bits 11:13 AMS
			// RW bits 14:15 SHN (00 no notification, 01 normal, 10 Abrupt )
			// RW bits 16:19 IO Q Size ( 2^n )
			// RW Completion Q Size (2^n)

			c.REGCtrlConfig = capsule.D12
			if c.REGCtrlConfig&0xc000 != 0 {
				// we instantly go to shutdown mode
				c.REGCtrlStatus |= 0x2 << 2
				c.REGCtrlStatus &= ^uint32(1) // clear ready bit
				// We're being asked for a shutdown

				c.Log.Trace(tracer.TraceCommands, "     Shutdown: %x", c.REGCtrlStatus)
				//w.Shutdown = true
				return nil
			}

			// 1.0 The controller starts with REGCtrlStatus:0 indicating the controller is offline
			if c.REGCtrlStatus == 0 {
				// (1.1) host has to first set  REGCtrlConfig:1 indicating it wants to enable the controller
				if c.REGCtrlConfig&0x1 == 0x1 {
					// (1.2) in response the controller (we) set  c.REGCtrlStatus |= 0x1
					c.REGCtrlStatus |= 0x1 // enable controller
				}
			}

		default:
			tracer.Fatal("protocol.FabricCmdPropertySet %+v", capsule)
		}

	case protocol.FabricCmdAuthenticationReceive:
		c.Log.Todo("protocol.FabricCmdAuthenticationReceive: %+v", capsule)
		// Nothing to do here

	case protocol.FabricCmdDisconnect:
		c.Log.Todo("protocol.FabricCmdDisconnect %+v", capsule)
		// Nothing to do here, target about to close the session

	default:
		c.Log.Todo("protocol.FabricCmdAuthenticationReceive: %+v", capsule)
		tracer.Fatal("unknown fabric command %x", capsule.FCType)
	}

	return nil
}
