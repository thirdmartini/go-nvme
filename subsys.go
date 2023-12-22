package nvme

import (
	"github.com/thirdmartini/go-nvme/targets"
)

// Subsystem defines the interface to an NVME Subsystem
type Subsystem interface {
	Identify(ctrlID uint16, cns uint8) ([]byte, error)
	GetLogPage(pageId int, offset uint64, length int) ([]byte, error)
	HandleIO(r *targets.IORequest) targets.TargetError
	QueueIO(r *targets.IORequest) targets.TargetError
	GetNQN() string
	GetRuntimeDetails() []targets.KV
}
