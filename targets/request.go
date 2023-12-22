package targets

import (
	"errors"
	"fmt"
)

type TargetError uint16

const (
	TargetErrorNone          TargetError = 0x0
	TargetErrorUnsupported   TargetError = 0x1
	TargetErrorAborted       TargetError = 0x3
	TargetErrorLbaOutOfRange TargetError = 0x4

	TargetErrorWrite    TargetError = 0x5
	TargetErrorRead     TargetError = 0x6
	TargetErrorInternal TargetError = 0xffff
)

type TargetCommand uint8

const (
	IORequestCmdRead      TargetCommand = 0x1
	IORequestCmdWrite     TargetCommand = 0x2
	IORequestCmdTrim      TargetCommand = 0x3
	IORequestCmdWriteZero TargetCommand = 0x4
	IORequestCmdFlush     TargetCommand = 0x5

	IORequestSnapshot TargetCommand = 0x20
)

// SGE defines a Scatter Gather Element
type SGE struct {
	Data []byte
}

/*
// Completion is a callback to complete the request
type Completion interface {
	Complete(status TargetError)
}*/

type Completer func(status TargetError)
type Executer func(r *IORequest)

type IORequest struct {
	Command         TargetCommand
	Lba             uint64
	Length          uint32
	SGL             [16]SGE
	SGLC            int
	ExecuteRequest  Executer
	CompleteRequest Completer
}

func (r *IORequest) Init(c TargetCommand, lba uint64, length uint32, completion Completer) *IORequest {
	r.Command = c
	r.Lba = lba
	r.Length = length
	r.CompleteRequest = completion
	r.SGLC = 0
	return r
}

func (r *IORequest) Assert() error {
	bufferLen := 0
	for i := 0; i < r.SGLC; i++ {
		bufferLen += len(r.SGL[i].Data)
	}
	if uint32(bufferLen) != r.Length {
		return fmt.Errorf("%d buffer and length mismatch sglc:%d %d:%d", r.Command, r.SGLC, uint32(bufferLen), r.Length)
	}

	if r.CompleteRequest == nil {
		return errors.New("no completion set")
	}
	return nil
}

func (r *IORequest) SetExecuter(e Executer) {
	r.ExecuteRequest = e
}

func (r *IORequest) AddBuffer(buffer []byte) {
	r.SGL[r.SGLC].Data = buffer
	r.SGLC++
}

func (r *IORequest) Buffers() []SGE {
	return r.SGL[0:r.SGLC]
}

func (r *IORequest) Start() {
}

func (r *IORequest) Complete(status TargetError) TargetError {
	r.CompleteRequest(status)
	return TargetErrorNone
}
