package nvme

import (
	"log"

	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
)

// handleCapsule services capsule commands
func (c *Controller) handleAdminCapsule(w *NVMEResponse, r *NVMERequest) error {
	capsule := r.Capsule()

	switch capsule.OpCode {
	case protocol.CapsuleCmdFabric: // Fabrics command.
		c.handleFabricCommand(w, r)

	//Figure 247: Identify – Identify Controller header Structure
	case protocol.CapsuleCmdIdentify:
		CNTID := uint16(capsule.D10 >> 16 & 0xFFFF)
		CNS := uint8(capsule.D10 & 0xFF)
		c.Log.Trace(tracer.TraceCapsuleDetail, "    CTRL:%d/%d CNS:%d NSID:%d (%s)", CNTID, c.ControllerID, CNS, capsule.FCType, c.Subsystem.GetNQN())

		data, err := c.Subsystem.Identify(c.ControllerID, CNS)
		if err != nil {
			log.Printf("protocol.CapsuleCmdIdentify: identify failure from subsystem: %s | err:%s\n", c.Subsystem.GetNQN(), err.Error())
			//tracer.Fatal("protocol.CapsuleCmdIdentify: identify failure from subsystem", capsule)
			w.SetStatus(protocol.CapsuleCmdInvalid)
			return nil
		}
		w.Write(data)

	case protocol.CapsuleCmdSetFeatures:
		fid := capsule.D10 & 0xff

		switch fid {
		case 0x20: // controller reset
			// FIXME: handle proper controller restet

		// 5.21.1.7 Number of Queues (Feature Identifier 07h)
		case protocol.FeatureNumberOfQueues:
			c.NCQR = uint16((capsule.D11 >> 16) & 0xFFFF)
			c.NSQR = uint16(capsule.D11 & 0xFFFF)

			c.Log.Trace(tracer.TraceCapsuleDetail, "    NCQR:%d  NSQR:%d", c.NCQR, c.NSQR)

			// limit queue count to 1 for now
			//c.NCQR = 0x0
			//c.NSQR = 0x0

			sm := serialize.New(w.Response.FabricResponse[:])
			v := uint64(c.NCQR)<<16 | uint64(c.NSQR)
			sm.Serialize(&v)
			//sm.Serialize(&c.NSQS)

		case protocol.FeatureAsyncEventConfig:
			v := capsule.D11 // contains teh feature word to set

			c.Log.Trace(tracer.TraceCapsuleDetail, "    AEC:%d/%x", v, v)
			// Enable async notification
			if v&(1<<31) != 0 {
				//c.EnableAsyncDiscoveryNotification = true
			}
			c.AEC = v
			sm := serialize.New(w.Response.FabricResponse[:])
			sm.Serialize(&v)

		default:
			w.SetStatus(protocol.SCCmdFeatureNotChangeable)
			c.Log.Todo("unsupported set features register: 0x%x", capsule.D10&0xFF)
		}

	case protocol.CapsuleCmdGetLogPage:
		val := capsule.D10 & 0xFF
		c.Log.Trace(tracer.TraceCapsuleDetail, "    LP:0x%0x", val)

		lpc := protocol.GetLogPageCommand{}
		if r.ReinterpretCapsule(&lpc) != nil {
			tracer.Fatal("protocol.CapsuleCmdGetLogPage: reinterpret failure")
		}

		dataLen := lpc.GetReturnBufferLength()
		c.Log.Trace(tracer.TraceCapsuleDetail, "    LP:0x%0x  Offset:%d Len:%d", val, lpc.GetReturnOffset(), dataLen)

		// Get the correct set based on what subsystem we are connected to
		data, err := c.Subsystem.GetLogPage(int(val), lpc.GetReturnOffset(), int(dataLen))
		if err != nil {
			log.Printf("protocol.CapsuleCmdGetLogPage err:%s\n", err.Error())
			tracer.Fatal("protocol.CapsuleCmdGetLogPage: %+v", capsule)
		}
		w.Write(data)

	case protocol.CapsuleCmdGetFeatures:
		feature := capsule.D10 & 0xff
		switch feature {
		case protocol.FeatureKeepAliveTimer:
			v := uint64(60000)
			sm := serialize.New(w.Response.FabricResponse[:])
			sm.Serialize(&v)

		case protocol.FeatureAsyncEventConfig:
			v := uint64(c.AEC)
			sm := serialize.New(w.Response.FabricResponse[:])
			sm.Serialize(&v)

		default:
			w.SetStatus(protocol.SCInternalError)
			c.Log.Todo("unsupported capsule feature command %+v", capsule)
		}

	case protocol.CapsuleCmdAsyncEventRequest:
		w.NoReply = true

	case protocol.CapsuleCmdKeepAlive:
		// Nothing to do

	case protocol.CapsuleCmdSecurityRecv:
		w.SetStatus(protocol.CapsuleCmdInvalid)

	default:
		w.SetStatus(protocol.SCInternalError)
		c.Log.Todo("unsupported capsule command %+v", capsule)
	}
	return nil
}
