package targets

import (
	"errors"
	"fmt"
	"os"
)

func init() {
	defaultFactory.RegisterProvider("file", FILECreateTarget)
}

// FileTarget implements target that write to file image
type FileTarget struct {
	File      *os.File
	imageName string
}

func (t *FileTarget) GetSize() uint64 {
	fi, err := t.File.Stat()
	if err != nil {
		return 0
	}

	return uint64(fi.Size())
}

func (t *FileTarget) Queue(r *IORequest) TargetError {
	switch r.Command {
	case IORequestCmdRead:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			cnt, err := t.File.ReadAt(r.SGL[i].Data, offset)
			if err != nil {
				fmt.Printf("Read Error: %s\n", err.Error())
				return r.Complete(TargetErrorUnsupported)
			}
			if cnt != len(r.SGL[i].Data) {
				panic(false)
			}
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			cnt, err := t.File.WriteAt(r.SGL[i].Data, offset)
			if err != nil {
				fmt.Printf("Write Error: %s\n", err.Error())
				return r.Complete(TargetErrorUnsupported)
			}
			if cnt != len(r.SGL[i].Data) {
				panic(false)
			}
			offset += int64(len(r.SGL[i].Data))
		}

		return r.Complete(TargetErrorNone)

	case IORequestCmdTrim, IORequestCmdWriteZero:
		offset := int64(r.Lba * 512)
		buf := make([]byte, r.Length, r.Length)
		cnt, err := t.File.WriteAt(buf, offset)
		if err != nil || cnt != len(buf) {
			return r.Complete(TargetErrorInternal)
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		return r.Complete(TargetErrorNone)

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *FileTarget) Start() error {
	return nil
}

func (t *FileTarget) Close() error {
	return t.File.Close()
}

func (t *FileTarget) GetRuntimeDetails() []KV {
	return []KV{
		{
			Key:   "Image",
			Value: t.imageName,
		},
	}
}

func FILECreateTarget(options Options) (Target, error) {
	img, ok := options["image"]
	if !ok {
		return nil, errors.New("no image option provided")
	}

	f, err := os.OpenFile(img, os.O_RDWR|os.O_SYNC, 0755)
	if err != nil {
		return nil, err
	}

	t := &FileTarget{
		File:      f,
		imageName: img,
	}

	return NewWorkQueue(options, t), nil
}
