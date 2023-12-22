package nvme

import (
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/targets"
)

func (c *Controller) handleIOCapsule(w *NVMEResponse, r *NVMERequest) error {
	status := targets.TargetErrorInternal
	capsule := r.Capsule()

	switch capsule.OpCode {
	case protocol.CapsuleCmdFlush:
		req := r.ior.Init(targets.IORequestCmdFlush, 0, 0, r.Complete)
		status = c.Subsystem.QueueIO(req)

	case protocol.CapsuleCmdRead:
		req := r.ior.Init(targets.IORequestCmdRead, capsule.Lba(), capsule.LbaLength()*512, r.Complete)
		r.payload = c.bufferManager.Get()
		r.payloadLength = int(req.Length)
		req.AddBuffer(r.Payload())
		w.Write(r.Payload())
		status = c.Subsystem.QueueIO(req)

	case protocol.CapsuleCmdWrite:
		req := r.ior.Init(targets.IORequestCmdWrite, capsule.Lba(), capsule.LbaLength()*512, r.Complete)
		req.AddBuffer(r.Payload())

		// fixme: alloc c.Log.Trace(TraceCapsuleDetail, "    Lba: %d  Length: %d  [in Payload: %d]", req.Lba, req.Length, len(req.SGL[0].Data))
		if uint32(len(req.SGL[0].Data)) == req.Length {
			status = c.Subsystem.QueueIO(req)
		} else {
			tracer.Fatal("Implement R2T")
			// Need to request data segments
			/*
				req.SGL[0].Data = make([]byte, req.Length, req.Length)
				r.State |= RequestNeedsData
				w.DataLength = req.Length
			*/
			// fixme: alloc c.Log.Trace(TraceCapsuleDetail, "    Lba: %d  Length: %d  [in Payload: %d] -- Needs Data", req.Lba, req.Length, len(req.SGL[0].Data))
		}

	case protocol.CapsuleCmdWriteZeros:
		var cmd targets.TargetCommand
		if capsule.D12&protocol.CommandBitDeallocateSet != 0 {
			cmd = targets.IORequestCmdTrim
		} else {
			cmd = targets.IORequestCmdWriteZero
		}

		req := r.ior.Init(cmd, capsule.Lba(), capsule.LbaLength()*512, r.Complete)
		status = c.Subsystem.QueueIO(req)

	case protocol.CapsuleCmdDatasetMgmt:
		// FIXME: this is typically a trim operation but there are additonal things that usefull (such as cache region ops)
		status = targets.TargetErrorNone
		c.Log.Todo("protocol.CapsuleCmdDatasetMgmt: %+v", capsule)
		r.Complete(targets.TargetErrorNone)

	default:
		tracer.Fatal("OpCode:default %+v", capsule)
		r.Complete(targets.TargetErrorUnsupported)
	}

	if status != targets.TargetErrorNone {
		w.SetStatus(protocol.SCInternalError)
		tracer.Fatal("bad status")
	}
	return nil
}
