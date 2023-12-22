package protocol

import (
	"encoding/binary"
	"fmt"

	"github.com/thirdmartini/go-nvme/internal/serialize"
)

type ConnectData struct {
	HostIdentifier [16]byte `offset:"0" length:"16"`
	CNTLID         uint16   `offset:"16"`
	SubNQN         string   `offset:"256" length:"256"`
	HostNQN        string   `offset:"512" length:"256"`
}

type ConnectCommand struct {
	OpCode uint8  `offset:"0"`
	PRP    uint8  `offset:"1"`
	CID    uint16 `offset:"2"`
	FCType uint8  `offset:"4"`
	//SGLDescriptor 24:39 `offset:"24"`
	RECFM     uint16 `offset:"40"`
	QueueID   uint16 `offset:"42"`
	QueueSize uint16 `offset:"44"`
	CATTR     uint8  `offset:"46"`
	KATO      uint32 `offset:"48"`
	data      [1024]byte
}

func (c *ConnectCommand) Unmarshal(data []byte) {
	c.OpCode = data[0]
	c.PRP = data[1]
	c.CID = binary.LittleEndian.Uint16(data[2:])
	c.FCType = data[4]
	c.RECFM = binary.LittleEndian.Uint16(data[40:])
	c.QueueID = binary.LittleEndian.Uint16(data[42:])
	c.QueueSize = binary.LittleEndian.Uint16(data[44:])
	c.CATTR = data[46]
	c.KATO = binary.LittleEndian.Uint32(data[48:])
}

func (c *ConnectCommand) Marshal(data []byte) {
	data[0] = c.OpCode
	data[1] = c.PRP
	binary.LittleEndian.PutUint16(data[2:], c.CID)
	data[4] = c.FCType

	binary.LittleEndian.PutUint16(data[40:], c.RECFM)
	binary.LittleEndian.PutUint16(data[42:], c.QueueID)
	binary.LittleEndian.PutUint16(data[44:], c.QueueSize)
	data[46] = c.CATTR
	binary.LittleEndian.PutUint32(data[48:], c.KATO)
}

func (c *ConnectCommand) String() string {
	return fmt.Sprintf("[CapReq] FabricConnect CID:%d QeueueId:%d QueueSize:%d",
		c.CID,
		c.QueueID,
		c.QueueSize)
}

func (c *ConnectCommand) SetCID(id uint16) {
	c.CID = id
}

func (c *ConnectCommand) GetConnectData() *ConnectData {
	ds := serialize.NewDeserializer(c.data[0:])

	fcd := ConnectData{}
	if ds.Deserialize(&fcd) != nil {
		return nil
	}

	return &fcd
}

func (c *ConnectCommand) SetConnectData(cd *ConnectData) error {
	s := serialize.New(c.data[0:])
	return s.Serialize(cd)
}

func (c *ConnectCommand) Data() []byte {
	return c.data[0:]
}
