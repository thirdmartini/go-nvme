package targets

import (
	"sync/atomic"
	"time"
)

const (
	defaultTestableTargetSize = 1024 * 1024 * 1024
)

func init() {
	defaultFactory.RegisterProvider("testable", func(options Options) (Target, error) {
		return NewTestableTarget(options), nil
	})
}

type TestableMediaRegion struct {
	Start     uint64
	End       uint64
	FailRead  bool
	FailWrite bool
}

type TestableTarget struct {
	SleepTime time.Duration

	ReadCount  uint64
	WriteCount uint64
	TrimCount  uint64
	ZeroCount  uint64
	FlushCount uint64

	Buffer  []byte
	Regions []TestableMediaRegion
}

func (t *TestableTarget) GetSize() uint64 {
	return uint64(len(t.Buffer))
}

func (t *TestableTarget) Queue(r *IORequest) TargetError {
	maxLba := t.GetSize() / 512

	reqLen := uint64(r.Length / 512)
	if r.Lba+reqLen >= maxLba {
		return r.Complete(TargetErrorLbaOutOfRange)
	}

	tmr := TestableMediaRegion{}
	for idx := range t.Regions {
		//fmt.Println(t.Regions[idx].Start, r.Lba, t.Regions[idx].End)

		if r.Lba >= t.Regions[idx].Start && r.Lba <= t.Regions[idx].End {
			tmr = t.Regions[idx]
			break
		}
	}

	time.Sleep(t.SleepTime)
	switch r.Command {
	case IORequestCmdRead:
		atomic.AddUint64(&t.ReadCount, 1)
		if tmr.FailRead {
			return r.Complete(TargetErrorRead)
		}

		offset := int64(r.Lba * 512)
		for i := range r.Buffers() {
			copy(r.SGL[i].Data, t.Buffer[offset:offset+int64(r.Length)])
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		atomic.AddUint64(&t.WriteCount, 1)
		if tmr.FailWrite {
			return r.Complete(TargetErrorWrite)
		}

		offset := int64(r.Lba * 512)
		for i := range r.Buffers() {
			copy(t.Buffer[offset:offset+int64(r.Length)], r.SGL[i].Data)
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdWriteZero:
		atomic.AddUint64(&t.ZeroCount, 1)
		offset := int64(r.Lba * 512)
		for i := int64(0); i < int64(r.Length); i++ {
			t.Buffer[offset+i] = 0
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdTrim:
		atomic.AddUint64(&t.TrimCount, 1)
		offset := int64(r.Lba * 512)
		for i := int64(0); i < int64(r.Length); i++ {
			t.Buffer[offset+i] = 0
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		atomic.AddUint64(&t.FlushCount, 1)
		return r.Complete(TargetErrorNone)

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *TestableTarget) Start() error {
	return nil
}

func (t *TestableTarget) Close() error {
	return nil
}

func (t *TestableTarget) GetRuntimeDetails() []KV {
	return nil
}

func NewTestableTarget(options Options) *TestableTarget {
	wait := options.Uint64("sleep", 5)

	return &TestableTarget{
		SleepTime: time.Millisecond * time.Duration(wait),
		Buffer:    make([]byte, defaultTestableTargetSize, defaultTestableTargetSize),
		Regions: []TestableMediaRegion{
			{
				Start:     1000,
				End:       1063,
				FailWrite: true,
			},
			{
				Start:    1064,
				End:      1127,
				FailRead: true,
			},
			{
				Start:     1128,
				End:       1256,
				FailWrite: true,
				FailRead:  true,
			},
		},
	}
}
