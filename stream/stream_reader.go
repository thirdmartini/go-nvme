package stream

import (
	"io"

	"github.com/thirdmartini/go-nvme/internal/utilities"
	"github.com/thirdmartini/go-nvme/protocol"
)

// Reader implements the on wire decoding of NVME protocol
type Reader struct {
	reader io.Reader
	CH     protocol.CommonHeader
	header [protocol.MaxHeaderSize]byte
	data   []byte
}

func (r *Reader) Dequeue() (*protocol.CommonHeader, error) {
	err := utilities.MustRead(r.reader, r.header[0:8])
	if err != nil {
		return nil, err
	}

	r.CH.Unmarshal(r.header[0:8])
	return &r.CH, nil
}

func (r *Reader) Receive(h protocol.PDU) error {
	//  Read the remainder of the header
	err := utilities.MustRead(r.reader, r.header[8:r.CH.HeaderLength])
	if err != nil {
		return err
	}

	h.Unmarshal(r.header[8:r.CH.HeaderLength])

	// FIXME: stop allocating things
	// Slurp the padding that may be between the header and the payload
	//   [HEADER][PAD][PAYLOAD]
	slurp := int64(r.CH.DataOffset) - int64(r.CH.HeaderLength)
	if slurp > 0 {
		null := make([]byte, slurp, slurp)
		err = utilities.MustRead(r.reader, null)
	}
	return nil
}

func (r *Reader) Length() uint32 {
	if r.CH.DataOffset != 0 {
		return r.CH.DataLength - uint32(r.CH.DataOffset)
	}
	return 0
}

func (r *Reader) ReceiveData(data []byte) error {
	return utilities.MustRead(r.reader, data)
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: r,
	}
}
