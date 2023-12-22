package tracer

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/thirdmartini/go-nvme/protocol"
)

const (
	TraceController    = uint64(0x1 << 0)
	TraceCommands      = uint64(0x1 << 1)
	TraceCapsule       = uint64(0x1 << 8)
	TraceCapsuleDetail = uint64(0x1 << 9)
	TraceCapsuleData   = uint64(0x1 << 10)
	TraceFabric        = uint64(0x1 << 16)
	TraceData          = uint64(0x1 << 60)
	TraceAll           = uint64(0xFFFFFFFFFFFFFFFF)
)

var AbortOnTodo = false

type Tracer interface {
	HexDump(level uint64, label string, data []byte)
	Trace(level uint64, format string, a ...interface{})
	TraceProtocol(level uint64, a interface{})
	TraceCapsule(isAdmin bool, c *protocol.CapsuleCommand)
	TraceFabric(c *protocol.CapsuleCommand)
	Begin(level uint64, format string, a ...interface{})
	End()

	Todo(f string, a ...interface{})
	Fatal(f string, a ...interface{})
}

func Fatal(f string, a ...interface{}) {
	fmt.Printf(f, "FATAL::")
	fmt.Printf(f, a...)
	fmt.Printf("\n")
	debug.PrintStack()
	os.Exit(-1)
}
