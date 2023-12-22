package tracer

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/thirdmartini/go-nvme/protocol"
)

// LogTracer performs protocol Level traces
type LogTracer struct {
	//	Ctrl  *Controller
	Id    string
	Level uint64
}

func (t *LogTracer) HexDump(level uint64, label string, data []byte) {
	if level&t.Level == 0 {
		return
	}

	fmt.Printf("%s | -- %s",
		t.Id,
		label)

	for idx := range data {
		if idx%32 == 0 {
			fmt.Printf("\n   | 0x%04x:", idx)
		}
		if idx%2 == 0 {
			fmt.Printf(" ")
		}
		fmt.Printf("%02x", data[idx])

	}
	fmt.Printf("\n==\n")
}

func (t *LogTracer) Begin(level uint64, format string, a ...interface{}) {
	if level&t.Level == 0 {
		return
	}
	fmt.Printf("%s | -- %s\n",
		t.Id,
		fmt.Sprintf(format, a...))
}

func (t *LogTracer) Trace(level uint64, format string, a ...interface{}) {
	if level&t.Level == 0 {
		return
	}
	fmt.Printf("%s |", t.Id)
	fmt.Printf(format, a...)
	fmt.Printf("\n")
}

func (t *LogTracer) TraceProtocol(level uint64, a interface{}) {
	if level&t.Level == 0 {
		return
	}
	fmt.Printf("%s |", t.Id)
	fmt.Print(a)
	fmt.Printf("\n")
}

func (t *LogTracer) TraceCapsule(isAdmin bool, c *protocol.CapsuleCommand) {
	if t.Level&TraceCapsule != 0 {
		var opName string
		var ok bool

		if isAdmin {
			opName, ok = protocol.AdminCmdToString[c.OpCode]
			if !ok {
				opName = "Admin/Unknown"
			}
		} else {
			opName, ok = protocol.IOCmdToString[c.OpCode]
			if !ok {
				opName = "Io/Unknown"
			}
		}

		if c.OpCode == protocol.CapsuleCmdFabric {
			fcName := protocol.FCTypeToString(c.FCType)
			fmt.Printf("%s | OpCode:%s (0x%0x) FCType:%s (0x%x)\n", t.Id, opName, c.OpCode, fcName, c.FCType)
		} else {
			fmt.Printf("%s | OpCode:%s (0x%0x)\n", t.Id, opName, c.OpCode)
		}
	}

	if t.Level&TraceCapsuleData != 0 {
		fmt.Printf("%s |   D:10 D:11 D:12 D:13 D:14 D:15\n", t.Id)
		fmt.Printf("%s |   %04x %04x %04x %04x %04x %04x\n", t.Id, c.D10, c.D11, c.D12, c.D13, c.D14, c.D15)
	}
}

func (t *LogTracer) TraceFabric(c *protocol.CapsuleCommand) {
	if t.Level&TraceFabric != 0 {
		opName, ok := protocol.FabricCmdToString[c.FCType]
		if !ok {
			opName = "Fabric/Unknown"
		}
		fmt.Printf("   | %s (0x%0x)\n", opName, c.FCType)
	}
}

func (t *LogTracer) End() {
	if t.Level != 0 {
		fmt.Printf("---------------------------------------------------------\n")
	}
}

func (t *LogTracer) Todo(f string, a ...interface{}) {
	fmt.Printf("TODO::\n")
	fmt.Printf(f, a...)
	fmt.Printf("\n")
	if AbortOnTodo {
		os.Exit(-1)
	}
}

func (t *LogTracer) Fatal(f string, a ...interface{}) {
	fmt.Printf("FATAL::\n")
	fmt.Printf(f, a...)
	fmt.Printf("\n")
	debug.PrintStack()
	os.Exit(-1)
}
