package nvme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/targets"
)

const (
	NVMEDiscoverySubsystemName = "nqn.2014-08.org.nvmexpress.discovery"
)

// DiscoverySubsystem implements an NVME over Fabrics Discovery Service
//
//	TODO: this always teturns our 1 target, but the goal here is to
//	      have this ALSO manage our dynamic target create/destroy
type DiscoverySubsystem struct {
	// The discovery subsystem needs a handle to the discovery server to get a list of subsystems
	Server *Server
}

func (s *DiscoverySubsystem) GetNQN() string {
	return NVMEDiscoverySubsystemName
}

// Identify implements Subsystem.Identify
func (s *DiscoverySubsystem) Identify(ctrlID uint16, cns uint8) ([]byte, error) {
	sm := serialize.New(make([]byte, 4096, 4096))

	switch cns {
	case protocol.CNSIdentifyController:
		id := protocol.IdentifyController{
			PCIVendor:         0,
			PCIDevice:         0,
			ControllerId:      ctrlID,
			Version:           protocol.NVMESpecificationVersion,
			OAES:              0x80000000,
			LogPageAttributes: 0x4,
			MaxCMDS:           protocol.NVMECtrlMaxCmds,
			SGLSupport:        1 | 1<<20,
			SubNQN:            NVMEDiscoverySubsystemName,
		}
		sm.Serialize(&id)
	default:
		return nil, fmt.Errorf("identify with unsupported cns:0x%x for subsystem", cns)
	}

	return sm.Get(), nil
}

// GetLogPage implements Subsystem.GetLogPage
// FIXME: log page should ignore the length parameter and we should trim the log in the caller
//
//	as we also have to deal with offsets
func (s *DiscoverySubsystem) GetLogPage(pageId int, offset uint64, length int) ([]byte, error) {
	ss := serialize.New(make([]byte, length, length))

	switch pageId {
	case protocol.LPDiscovery: // GetList of Discovery Targets we serve
		// Todo, something we could do here that may be less convoluted is to just jkeep the marshalled log page in memory

		subsys := make([]Subsystem, 0, len(s.Server.SubSystems))

		for k := range s.Server.SubSystems {
			if k == NVMEDiscoverySubsystemName {
				continue
			}
			subsys = append(subsys, s.Server.SubSystems[k])
		}

		sort.Slice(subsys, func(i, j int) bool {
			return subsys[i].GetNQN() < subsys[j].GetNQN()
		})

		if offset == 0 {
			dlp := protocol.DiscoveryLogPage{
				GenerationCounter: 0,
				RecordFormat:      0,
				NumberOfRecords:   uint64(len(subsys)),
			}

			// Fixme turn this into an api

			s.Server.Lock.Lock()
			ipAndPort := strings.Split(s.Server.Address, ":")

			for _, v := range subsys {
				dlpe := protocol.DiscoveryLogPageEntry{
					TransportType:         0x03, // TCP
					AddressFamily:         0x01, // AF_INET
					SubsystemType:         0x02, // NVMe Device
					TransportRequirements: 0x04,
					PortId:                0x1,
					ControllerId:          0xffff,       // Dynamic Controller Model
					AdminMaxQueueSize:     0x2000,       //
					TransportServiceId:    ipAndPort[1], // port
					SubNQN:                v.GetNQN(),
					TransportAddress:      ipAndPort[0],
				}
				dlp.DiscoveryLofEntries = append(dlp.DiscoveryLofEntries, dlpe)
			}
			s.Server.Lock.Unlock()
			ss.Serialize(&dlp)
		} else {
			dlp := protocol.DiscoveryLogPageData{}
			s.Server.Lock.Lock()

			idxOffset := int(offset/1024) - 1 // need -1 because the first entry is always only 3 indexes

			ipAndPort := strings.Split(s.Server.Address, ":")
			for idx := idxOffset; idx < len(subsys); idx++ {
				v := subsys[idx]

				dlpe := protocol.DiscoveryLogPageEntry{
					TransportType:         0x03, // TCP
					AddressFamily:         0x01, // AF_INET
					SubsystemType:         0x02, // NVMe Device
					TransportRequirements: 0x04,
					PortId:                0x1,
					ControllerId:          0xffff,       // Dynamic Controller Model
					AdminMaxQueueSize:     0x2000,       //
					TransportServiceId:    ipAndPort[1], // port
					SubNQN:                v.GetNQN(),
					TransportAddress:      ipAndPort[0],
				}
				dlp.DiscoveryLofEntries = append(dlp.DiscoveryLofEntries, dlpe)
			}
			s.Server.Lock.Unlock()
			ss.Serialize(&dlp)
		}

	default:
		return nil, fmt.Errorf("log page 0x%x not supported by discover subsystem", pageId)
	}
	return ss.Get(), nil
}

func (s *DiscoverySubsystem) HandleIO(r *targets.IORequest) targets.TargetError {
	return targets.TargetErrorUnsupported
}

func (s *DiscoverySubsystem) QueueIO(r *targets.IORequest) targets.TargetError {
	return targets.TargetErrorUnsupported
}

func (s *DiscoverySubsystem) GetRuntimeDetails() []targets.KV {
	return nil
}
