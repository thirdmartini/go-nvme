package nvme

import (
	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/targets"
)

// NVMERequest wrap our CapsuleCommand in needed queue structures
type NVMERequest struct {
	payload       []byte // additional payload
	payloadLength int

	capsule protocol.CapsuleCommand
	ior     targets.IORequest

	response   NVMEResponse
	completion chan *NVMERequest

	// The Queue ID that this request belongs to
	qid uint16

	// flag indicating if this request is active
	// fixme: we should not need this in the new processing model
	active bool

	// State tracking for any R2T requests that need to be made (Fixme this does not work right now)
	State     uint32
	R2TOffset uint32
	R2TLength uint32
}

// Capsule returns the unmarshaled capsule
func (r *NVMERequest) Capsule() *protocol.CapsuleCommand {
	return &r.capsule
}

// ReinterpretCapsule  reinterprets this capsule request into the passed in v type
func (r *NVMERequest) ReinterpretCapsule(v interface{}) error {
	var tmp [protocol.MaxHeaderSize]byte
	s := serialize.New(tmp[0:])
	s.Serialize(&r.capsule)
	ds := serialize.NewDeserializer(tmp[:])
	return ds.Deserialize(v)
}

// Payload returns payload data that followed our header
func (r *NVMERequest) Payload() []byte {
	return r.payload[0:r.payloadLength]
}

func (r *NVMERequest) SetStatus(code protocol.NVMEStatusCode) {
	r.response.SetStatus(code)
}

// Complete asynchronously complete this request to the bottom half
//
//	This essentially schedules the request for completion in the BH handler of the controller
func (r *NVMERequest) Complete(status targets.TargetError) {
	// fast path the status
	if status == 0 {
		r.completion <- r
		return
	}

	/* Convert errors (use a lookup table down the road */
	switch status {
	case targets.TargetErrorNone:
		r.SetStatus(protocol.SCSuccess)

	case targets.TargetErrorLbaOutOfRange:
		r.SetStatus(protocol.SCLBAOutOfRange)

	case targets.TargetErrorWrite:
		r.SetStatus(protocol.SCMediaWriteFault)

	case targets.TargetErrorRead:
		r.SetStatus(protocol.SCMediaUncorrectableReadError)

	case targets.TargetErrorUnsupported:
		r.SetStatus(protocol.SCInvalidCommandOpcode)

	default:
		r.SetStatus(protocol.SCInternalError)
	}
	r.completion <- r
}
