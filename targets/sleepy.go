package targets

import (
	"time"
)

func init() {
	defaultFactory.RegisterProvider("sleepy", SleepyCreateTarget)
}

type SleepyTarget struct {
	Size      uint64
	SleepTime time.Duration
}

func (t *SleepyTarget) GetSize() uint64 {
	return t.Size
}

func (t *SleepyTarget) Queue(r *IORequest) TargetError {
	time.Sleep(t.SleepTime)
	switch r.Command {
	case IORequestCmdRead:
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		return r.Complete(TargetErrorNone)

	case IORequestCmdWriteZero:
		return r.Complete(TargetErrorNone)

	case IORequestCmdTrim:
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		return r.Complete(TargetErrorNone)

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *SleepyTarget) Start() error {
	return nil
}

func (t *SleepyTarget) Close() error {
	return nil
}

func (t *SleepyTarget) GetRuntimeDetails() []KV {
	return nil
}

func SleepyCreateTarget(options Options) (Target, error) {
	s := &SleepyTarget{
		Size:      1024 * 1024 * 1024 * 1024,
		SleepTime: time.Millisecond * 5,
	}
	return NewWorkQueue(options, s), nil
}
