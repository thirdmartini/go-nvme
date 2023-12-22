package targets

func init() {
	defaultFactory.RegisterProvider("null", func(options Options) (Target, error) {
		return NewNullTarget(options), nil
	})
}

type NullTarget struct {
	Size uint64
}

func (t *NullTarget) GetSize() uint64 {
	return t.Size
}

func (t *NullTarget) Queue(r *IORequest) TargetError {
	switch r.Command {
	case IORequestCmdRead:
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		return r.Complete(TargetErrorNone)

	case IORequestCmdWriteZero:
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		return r.Complete(TargetErrorNone)

	case IORequestCmdTrim:
		return r.Complete(TargetErrorNone)

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *NullTarget) Start() error {
	return nil
}

func (t *NullTarget) Close() error {
	return nil
}

func (t *NullTarget) GetRuntimeDetails() []KV {
	return nil
}

func NewNullTarget(options Options) *NullTarget {
	return &NullTarget{
		Size: 1024 * 1024 * 1024 * 1024,
	}
}
