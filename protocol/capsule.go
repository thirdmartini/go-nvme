package protocol

import (
	"encoding/binary"
	"fmt"
)

type CapsuleCommand struct {
	OpCode uint8    `offset:"0"`
	PRP    uint8    `offset:"1"`
	CID    uint16   `offset:"2"`
	FCType uint8    `offset:"4"`
	DPTR   [16]byte `offset:"24" length:"16"` // FIXME what is the offset of DPTR ?
	D10    uint32   `offset:"40"`
	D11    uint32   `offset:"44"`
	D12    uint32   `offset:"48"`
	D13    uint32   `offset:"52"`
	D14    uint32   `offset:"56"`
	D15    uint32   `offset:"60"`
}

func (c *CapsuleCommand) Unmarshal(data []byte) {
	c.OpCode = data[0]
	c.PRP = data[1]
	c.CID = binary.LittleEndian.Uint16(data[2:])
	c.FCType = data[4]
	copy(c.DPTR[0:16], data[24:])
	c.D10 = binary.LittleEndian.Uint32(data[40:])
	c.D11 = binary.LittleEndian.Uint32(data[44:])
	c.D12 = binary.LittleEndian.Uint32(data[48:])
	c.D13 = binary.LittleEndian.Uint32(data[52:])
	c.D14 = binary.LittleEndian.Uint32(data[56:])
	c.D15 = binary.LittleEndian.Uint32(data[60:])
}

func (c *CapsuleCommand) Marshal(data []byte) {
	data[0] = c.OpCode
	data[1] = c.PRP
	binary.LittleEndian.PutUint16(data[2:], c.CID)
	data[4] = c.FCType
	copy(data[24:16+24], c.DPTR[0:16])
	binary.LittleEndian.PutUint32(data[40:], c.D10)
	binary.LittleEndian.PutUint32(data[44:], c.D11)
	binary.LittleEndian.PutUint32(data[48:], c.D12)
	binary.LittleEndian.PutUint32(data[52:], c.D13)
	binary.LittleEndian.PutUint32(data[56:], c.D14)
	binary.LittleEndian.PutUint32(data[60:], c.D15)
}

func (c *CapsuleCommand) SetCID(id uint16) {
	c.CID = id
}

func (c *CapsuleCommand) String() string {
	return fmt.Sprintf("[CapReq] Op: 0x%x  CID:%d", c.OpCode, c.CID)
}

// Lba returns the logical block (in BLOCK SIZE ) blocks of the offset of the io
func (c *CapsuleCommand) Lba() uint64 {
	return uint64(c.D11)<<32 | uint64(c.D10)
}

// LbaLength returns the length of the request in LBA Blocks
//
//	target needs to adjust this for the BLOCK size it is uinsg
func (c *CapsuleCommand) LbaLength() uint32 {
	return c.D12&0xFFFF + 1
}

type CapsuleResponse struct {
	FabricResponse [8]uint8 `offset:"0" length:"8"`
	SQHD           uint16   `offset:"8"`
	QueueID        uint16   `offset:"10"` // TODO: see Fabric command but this is wrong
	CID            uint16   `offset:"12"`
	Status         uint16   `offset:"14"`
}

func (c *CapsuleResponse) String() string {
	return fmt.Sprintf("[CapRes] SQHD:%d  QID:%d  CID:%d  Status:0x%x",
		c.SQHD,
		c.QueueID,
		c.CID,
		c.Status)
}

func (c *CapsuleResponse) Unmarshal(data []byte) {
	copy(c.FabricResponse[0:8], data[0:8])
	c.SQHD = binary.LittleEndian.Uint16(data[8:])
	c.QueueID = binary.LittleEndian.Uint16(data[10:])
	c.CID = binary.LittleEndian.Uint16(data[12:])
	c.Status = binary.LittleEndian.Uint16(data[14:])
}

func (c *CapsuleResponse) Marshal(data []byte) {
	copy(data[0:8], c.FabricResponse[0:8])
	binary.LittleEndian.PutUint16(data[8:], c.SQHD)
	binary.LittleEndian.PutUint16(data[10:], c.QueueID)
	binary.LittleEndian.PutUint16(data[12:], c.CID)
	binary.LittleEndian.PutUint16(data[14:], c.Status)
}

func (c CapsuleResponse) GetStatus() NVMEStatusCode {
	return NVMEStatusCode(c.Status)
}

func (c *CapsuleResponse) SetStatus(code NVMEStatusCode) {
	if code == SCSuccess {
		c.Status = 0
		return
	}
	//c.Status = uint16(code) | 0x4000 // not sure why we were using 0x4000 before
	c.Status = uint16(code) | uint16(SCFlagDoNotRetry)
}

type CapsuleSetFeaturesRequest struct {
	OpCode uint8    `offset:"0"`
	PRP    uint8    `offset:"1"`
	CID    uint16   `offset:"2"`
	FCType uint8    `offset:"4"`
	DPTR   [16]byte `offset:"24" length:"16"` // FIXME what is the offset of DPTR ?
	D10    uint32   `offset:"40"`
	D14    uint32   `offset:"56"`
}
