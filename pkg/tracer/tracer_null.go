package tracer

import (
	"fmt"
	"os"

	"github.com/thirdmartini/go-nvme/protocol"
)

// NullTracer is used for keeping traces quiet
type NullTracer struct{}

func (t *NullTracer) HexDump(level uint64, label string, data []byte)       {}
func (t *NullTracer) Trace(level uint64, format string, a ...interface{})   {}
func (t *NullTracer) TraceProtocol(level uint64, a interface{})             {}
func (t *NullTracer) TraceCapsule(isAdmin bool, c *protocol.CapsuleCommand) {}
func (t *NullTracer) TraceFabric(c *protocol.CapsuleCommand)                {}
func (t *NullTracer) Begin(level uint64, format string, a ...interface{})   {}
func (t *NullTracer) End()                                                  {}

func (t *NullTracer) Todo(f string, a ...interface{}) {
	fmt.Printf("TODO::\n")
	fmt.Printf(f, a...)
	fmt.Printf("\n")
	if AbortOnTodo {
		os.Exit(-1)
	}
}

func (t *NullTracer) Fatal(f string, a ...interface{}) {
	fmt.Printf("FATAL::\n")
	fmt.Printf(f, a...)
	fmt.Printf("\n")
	os.Exit(-1)
}
