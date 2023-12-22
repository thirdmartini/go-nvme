package protocol

import (
	"encoding/binary"
	"fmt"
)

// CommonHeader (CH) defines the 8 byte preamble of every command header
type CommonHeader struct {
	Type         uint8  `offset:"0"`
	Flags        uint8  `offset:"1"`
	HeaderLength uint8  `offset:"2"`
	DataOffset   uint8  `offset:"3"`
	DataLength   uint32 `offset:"4"`
}

func (r CommonHeader) String() string {
	return fmt.Sprintf("[HDR   ] Type: 0x%x (%s) HL:%d DO:%d DL:%d",
		r.Type,
		HeaderTypeToString(r.Type),
		r.HeaderLength,
		r.DataOffset,
		r.DataLength)
}

func (r *CommonHeader) Marshal(data []byte) {
	data[0] = r.Type
	data[1] = r.Flags
	data[2] = r.HeaderLength
	data[3] = r.DataOffset
	binary.LittleEndian.PutUint32(data[4:], r.DataLength)
}

func (r *CommonHeader) Unmarshal(data []byte) {
	r.Type = data[0]
	r.Flags = data[1]
	r.HeaderLength = data[2]
	r.DataOffset = data[3]
	r.DataLength = binary.LittleEndian.Uint32(data[4:])
}
