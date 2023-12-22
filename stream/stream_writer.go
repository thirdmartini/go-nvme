package stream

import (
	"io"
	"log"

	"github.com/thirdmartini/go-nvme/internal/utilities"
	"github.com/thirdmartini/go-nvme/protocol"
)

type Writer struct {
	writer io.Writer
	CH     protocol.CommonHeader
	header [protocol.MaxHeaderSize]byte
	data   []byte
}

func (s *Writer) Flush() error {
	s.CH.Marshal(s.header[0:8])

	/*
		// buffered code net.Buffers code should allow us to use read.v ... but alas it performs memory allocations which make it actually
		// slower
		if len(s.data) != 0 {
			v := make(net.Buffers, 0, 3)
			v = append(v, s.header[0:s.CH.HeaderLength])
			v = append(v, s.data)
			v.WriteTo(s.writer)
		} else {
			err := utilities.MustWrite(s.writer, s.header[0:s.CH.HeaderLength])
			if err != nil {
				return err
			}
		}*/

	err := utilities.MustWrite(s.writer, s.header[0:s.CH.HeaderLength])
	if err != nil {
		return err
	}

	if len(s.data) != 0 {
		return utilities.MustWrite(s.writer, s.data)
	}
	return nil
}

func (s *Writer) Send(t uint8, hdr protocol.PDU, hlen uint8) error {
	s.CH.Type = t
	s.CH.DataOffset = 0
	s.data = nil

	if hlen > 120 {
		log.Fatal("Come on man!\n") // FIXME: we should never ever have this issue
	}

	if hlen == 0 {
		log.Fatal("Come on man!\n") // FIXME: we should never ever have this issue
		//s.CH.HeaderLength = protocol.MaxHeaderSize
		//s.CH.DataLength = protocol.MaxHeaderSize
	} else {
		s.CH.HeaderLength = hlen + 8
		s.CH.DataLength = uint32(s.CH.HeaderLength)
	}

	hdr.Marshal(s.header[8:])
	return nil
}

func (s *Writer) MarshalWithData2(t uint8, flags uint8, hdr protocol.PDU, hlen uint8, data []byte) error {
	s.CH.Type = t
	s.CH.Flags = flags
	s.CH.HeaderLength = hlen + 8
	s.CH.DataLength = uint32(s.CH.HeaderLength) + uint32(len(data))
	s.CH.DataOffset = 0
	if data != nil {
		s.CH.DataOffset = s.CH.HeaderLength
	}
	s.data = data

	hdr.Marshal(s.header[8:])
	return nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}
