package targets

import (
	"errors"
	"strconv"
)

func init() {
	defaultFactory.RegisterProvider("mem", MEMCreateTarget)
}

// MemTarget implements target that write to file image
type MemTarget struct {
	Buffer []byte
}

func (t *MemTarget) GetSize() uint64 {
	return uint64(len(t.Buffer))
}

func (t *MemTarget) Queue(r *IORequest) TargetError {
	// fmt.Printf("-->> Handle: 0x%0x Len:%d\n", r.Command, len(r.SGL[0].Data))

	switch r.Command {
	case IORequestCmdRead:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			copy(r.SGL[i].Data, t.Buffer[offset:offset+int64(r.Length)])
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			copy(t.Buffer[offset:offset+int64(r.Length)], r.SGL[i].Data)
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		return r.Complete(TargetErrorNone)

	case IORequestCmdWriteZero, IORequestCmdTrim:
		offset := int64(r.Lba * 512)
		for i := int64(0); i < int64(r.Length); i++ {
			t.Buffer[offset+i] = 0
		}
		return r.Complete(TargetErrorNone)

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *MemTarget) Start() error {
	return nil
}

func (t *MemTarget) Close() error {
	return nil
}

func (t *MemTarget) GetRuntimeDetails() []KV {
	return []KV{}
}

func MEMCreateTarget(options Options) (Target, error) {
	szParm, ok := options["size"]
	if !ok {
		return nil, errors.New("no size provided")
	}

	sz, err := strconv.ParseInt(szParm, 10, 32)
	if err != nil {
		return nil, err
	}

	return &MemTarget{
		Buffer: make([]byte, sz, sz),
	}, nil
}
