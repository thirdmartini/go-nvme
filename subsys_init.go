package nvme

import (
	"fmt"

	"github.com/thirdmartini/go-nvme/targets"
)

// InitSubsystem subsystem that is set on an initial login
//
//	has limited functionality and will throw errors
type InitSubsystem struct {
	RootController *Controller
}

func (s *InitSubsystem) Identify(ctrlID uint16, cns uint8) ([]byte, error) {
	return nil, fmt.Errorf("identify command not supported on root subsystem")
}

func (s *InitSubsystem) GetLogPage(pageId int, offset uint64, length int) ([]byte, error) {
	return nil, fmt.Errorf("log page command not supported on root subsystem")
}

func (s *InitSubsystem) HandleIO(r *targets.IORequest) targets.TargetError {
	return targets.TargetErrorUnsupported
}

func (s *InitSubsystem) QueueIO(r *targets.IORequest) targets.TargetError {
	return targets.TargetErrorUnsupported
}

func (s *InitSubsystem) GetNQN() string {
	return "nqn.2014-08.org.nvmexpress.init"
}

func (s *InitSubsystem) GetRuntimeDetails() []targets.KV {
	return nil
}
