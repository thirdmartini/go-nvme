package nvme

import (
	"github.com/thirdmartini/go-nvme/protocol"
)

// DataSGLE is a Scatter gather entry for Data buffer writes
type DataSGLE struct {
	Data []byte
}

type SGL []DataSGLE

const (
	MaxSGL           = 16
	RequestNeedsData = 0x1
)

type NVMEResponse struct {
	Type uint8
	CID  uint16

	sgl  [MaxSGL]DataSGLE
	sglc int

	Response protocol.CapsuleResponse
	C2H      protocol.C2HDataTransfer

	State      uint32 // state flags on the request
	DataLength uint32 // how much data is needed to be received

	NoReply  bool // flags o // n the request
	Shutdown bool
}

func (w *NVMEResponse) Reset() {
	w.CID = 0
	w.sglc = 0
	w.NoReply = false
	w.Shutdown = false
}

// Write this just sets our response buffers
func (w *NVMEResponse) Write(data []byte) {
	w.sgl[w.sglc].Data = data
	w.sglc++
}

func (w *NVMEResponse) SetStatus(code protocol.NVMEStatusCode) {
	w.Response.SetStatus(code)
}
