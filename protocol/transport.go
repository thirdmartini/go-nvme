package protocol

import (
	"encoding/binary"
	"fmt"
)

// C2HDataTransfer (C2HData) represents the header sent to a host alongside a data transfer from the controller.
// This is used for all READ operations
type C2HDataTransfer struct {
	CCCID uint16 `offset:"0"`
	DATAO uint32 `offset:"4"`
	DATAL uint32 `offset:"8"`
}

// String returns a pretty string representation of the C2HData PDU
func (c *C2HDataTransfer) String() string {
	return fmt.Sprintf("[C2H   ] CID:%d  Ofs:%d Len:%d", c.CCCID, c.DATAO, c.DATAL)
}

// Marshal marshals the data structure onto a stream buffer
func (c *C2HDataTransfer) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.CCCID)
	binary.LittleEndian.PutUint32(data[4:], c.DATAO)
	binary.LittleEndian.PutUint32(data[8:], c.DATAL)
}

// Unmarshal decodes the structure from a stream buffer
func (c *C2HDataTransfer) Unmarshal(data []byte) {
	c.CCCID = binary.LittleEndian.Uint16(data[0:])
	c.DATAO = binary.LittleEndian.Uint32(data[4:])
	c.DATAL = binary.LittleEndian.Uint32(data[8:])
}

// R2TRequest (R2T) represents a request to transmit message from the controller to host.
// This is sent when the controller needs additional data PDU from the host for a Capsule command
type R2TRequest struct {
	CCCID uint16 `offset:"0"`
	TTAG  uint16 `offset:"2"`
	DATAO uint32 `offset:"4"`
	DATAL uint32 `offset:"8"`
}

// String returns a pretty string representation of the R2T
func (c *R2TRequest) String() string {
	return fmt.Sprintf("[R2T   ] CID:%d  Ofs:%d Len:%d Tag:%d", c.CCCID, c.DATAO, c.DATAL, c.TTAG)
}

// Marshal marshals the data structure onto a stream buffer
func (c *R2TRequest) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.CCCID)
	binary.LittleEndian.PutUint16(data[2:], c.TTAG)
	binary.LittleEndian.PutUint32(data[4:], c.DATAO)
	binary.LittleEndian.PutUint32(data[8:], c.DATAL)
}

// Unmarshal decodes the structure from a stream buffer
func (c *R2TRequest) Unmarshal(data []byte) {
	c.CCCID = binary.LittleEndian.Uint16(data[0:])
	c.TTAG = binary.LittleEndian.Uint16(data[2:])
	c.DATAO = binary.LittleEndian.Uint32(data[4:])
	c.DATAL = binary.LittleEndian.Uint32(data[8:])
}

// C2HTermRequest (C2HTermReq) represents a Controller to host termination request.
// This request may be accompanied by a data buffer containing the offending PDU/Capsule
type C2HTermRequest struct {
	FatalErrorStatus      uint16
	FatalErrorInformation uint16
}

// String returns a pretty string representation of the C2HTerm
func (c *C2HTermRequest) String() string {
	return fmt.Sprintf("[C2H   ] FES:0x%x  FEI:0x%x", c.FatalErrorStatus, c.FatalErrorInformation)
}

// Marshal marshals the data structure onto a stream buffer
func (c *C2HTermRequest) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.FatalErrorStatus)
	binary.LittleEndian.PutUint16(data[2:], c.FatalErrorInformation)
}

// Unmarshal decodes the structure from a stream buffer
func (c *C2HTermRequest) Unmarshal(data []byte) {
	c.FatalErrorStatus = binary.LittleEndian.Uint16(data[0:])
	c.FatalErrorInformation = binary.LittleEndian.Uint16(data[2:])
}

// H2CTermRequest (H2CTermReq) represents a Host to Controller termination request.
// This request may be accompanied by a data buffer containing the offending PDU/Capsule
type H2CTermRequest struct {
	FatalErrorStatus      uint16
	FatalErrorInformation uint16
}

// String returns a pretty string representation of the H2CTerm
func (c *H2CTermRequest) String() string {
	return fmt.Sprintf("[H2C   ] FES:0x%x  FEI:0x%x", c.FatalErrorStatus, c.FatalErrorInformation)
}

// Marshal marshals the data structure onto a stream buffer
func (c *H2CTermRequest) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.FatalErrorStatus)
	binary.LittleEndian.PutUint16(data[2:], c.FatalErrorInformation)
}

// Unmarshal decodes the structure from a stream buffer
func (c *H2CTermRequest) Unmarshal(data []byte) {
	c.FatalErrorStatus = binary.LittleEndian.Uint16(data[0:])
	c.FatalErrorInformation = binary.LittleEndian.Uint16(data[2:])
}
