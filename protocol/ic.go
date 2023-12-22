package protocol

import (
	"encoding/binary"
	"fmt"
)

type ICRequest struct {
	PDUFormatVersion uint16 `offset:"0"`
	PDUDataAlignment uint8  `offset:"2"`
	PDUDataDigest    uint8  `offset:"3"`
	PDUMaxR2T        uint32 `offset:"4"`
}

func (c *ICRequest) String() string {
	return fmt.Sprintf("[ICReq ] FW:%d  DA:%d DD:%d MaxR2T:%d",
		c.PDUFormatVersion,
		c.PDUDataAlignment,
		c.PDUDataDigest,
		c.PDUMaxR2T)
}

func (c *ICRequest) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.PDUFormatVersion)
	data[2] = c.PDUDataAlignment
	data[3] = c.PDUDataDigest
	binary.LittleEndian.PutUint32(data[4:], c.PDUMaxR2T)
}

func (c *ICRequest) Unmarshal(data []byte) {
	c.PDUFormatVersion = binary.LittleEndian.Uint16(data[0:])
	c.PDUDataAlignment = data[2]
	c.PDUDataDigest = data[3]
	c.PDUMaxR2T = binary.LittleEndian.Uint32(data[4:])
}

type ICResponse struct {
	PDUFormatVersion uint16 `offset:"0"`
	PDUDataAlignment uint8  `offset:"2"`
	PDUDataDigest    uint8  `offset:"3"`
	MaxH2CDataLength uint32 `offset:"4"`
}

func (c *ICResponse) String() string {
	return fmt.Sprintf("[ICRes ] FW:%d  DA:%d DD:%d MaxDataLen:%d",
		//		c.PDULength,
		c.PDUFormatVersion,
		c.PDUDataAlignment,
		c.PDUDataDigest,
		c.MaxH2CDataLength)
}

func (c *ICResponse) Marshal(data []byte) {
	binary.LittleEndian.PutUint16(data[0:], c.PDUFormatVersion)
	data[2] = c.PDUDataAlignment
	data[3] = c.PDUDataDigest
	binary.LittleEndian.PutUint32(data[4:], c.MaxH2CDataLength)
}

func (c *ICResponse) Unmarshal(data []byte) {
	//	c.PDULength = binary.LittleEndian.Uint32(data[4:])
	c.PDUFormatVersion = binary.LittleEndian.Uint16(data[0:])
	c.PDUDataAlignment = data[2]
	c.PDUDataDigest = data[3]
	c.MaxH2CDataLength = binary.LittleEndian.Uint32(data[4:])
}
