package nvme

import (
	"fmt"

	"github.com/thirdmartini/go-nvme/internal/serialize"
	"github.com/thirdmartini/go-nvme/protocol"
	"github.com/thirdmartini/go-nvme/targets"
)

const (
	MaximumDataSize = 65536 //524288
	//MaximumDataSize = 524288
	MaximumPDUDataSize = (64 + MaximumDataSize) / 16 // 64Bytes for Header + 262144 for data
	// and the value must be in 16byte units

	NumberOfNamespaces = 1
)

type TargetSubsystem struct {
	NQN             string
	UUID            [16]byte
	ModelName       string
	SerialNumber    string
	FirmwareVersion string
	Target          targets.Target
}

func (s *TargetSubsystem) GetNQN() string {
	return s.NQN
}

func (s *TargetSubsystem) Identify(ctrlID uint16, cns uint8) ([]byte, error) {
	sm := serialize.New(make([]byte, 4096, 4096))

	switch cns {
	case protocol.CNSIdentifyNamespace:
		lbaCount := s.Target.GetSize() / 512

		id := protocol.IdentifyNamespaceData{
			NSZE:     lbaCount,
			NCAP:     lbaCount,
			NUSE:     lbaCount,
			NSFEAT:   0x0, // was 0x2
			NLBAF:    0x0, // 0based (ie +1)
			NMIC:     0x1,
			RESCAP:   0xff, //0x12,
			FPI:      0x80, // was 0c0
			ANAGRPID: 0x1,  // was  0x1
			NVMCAP:   [2]uint64{s.Target.GetSize(), 0},
		}
		id.LBAF[0] = uint32(9 << 16) // 512 bytes
		sm.Serialize(&id)

	case protocol.CNSIdentifyController: //0x01
		// Identify Controller header structure for the controller processing the command
		id := protocol.IdentifyController{
			PCIVendor:           0x144d,
			PCIDevice:           0x144d,
			RAB:                 0x6, //2^6
			OUI:                 [3]byte{0x38, 0x25, 0x00},
			CMIC:                0xb, // More than 1 port, more than 1 controller, and Asymmetric Namespace Access (ANA)
			MDTS:                4,   // this is 2^n * size of CAP.MPSMIN ( 4096*2^4) == 64K will be the biggest transfer from the  host to us
			ControllerId:        ctrlID,
			Version:             protocol.NVMESpecificationVersion,
			OAES:                0x0, // 0x0900,
			CTRATT:              0x0,
			CNTRLTYPE:           0x1,  // CNTRLTYPE is required for NVME 1.4 or newer
			OACS:                0x17, // 0x1 << 7, // support virtualization
			ACL:                 0x7,  // 0x3,
			AERL:                0x3,
			FRWM:                0x16, //0x3,
			LogPageAttributes:   0x3,  // 0x07,
			ErrorLogPageEntries: 0x3f, //0x7f,
			KAS:                 30,   // 10 seconds ( 100x100ms)
			ANATT:               0xa,
			ANACAP:              0x1f,
			ANAGRPMAX:           0x80,
			NANAGRPID:           0x80,
			TNVMCAP0:            s.Target.GetSize(),
			SQES:                0x66,
			CQES:                0x44,
			MaxCMDS:             protocol.NVMECtrlMaxCmds,
			NumberNamespaces:    NumberOfNamespaces,
			ONCS:                0xc,
			//			FUSES:               0x1,
			FNA:   0x5, //0x0,
			AWUN:  0xffff,
			AWUPF: 0x800,
			//			ACWU:                63, // 32K (64 x 512by block)
			VWC:        0x1,
			NWPC:       0x1,
			SGLSupport: 0x1,
			MNAN:       NumberOfNamespaces,
			SubNQN:     s.NQN,
			IOCCSZ:     MaximumPDUDataSize,
			IORCSZ:     0x01, // 16 bytes
			ICDOFF:     0,
			FCATT:      0,
			MSDBD:      1,
			OFCS:       0,
			PSD: [32]byte{0xC4, 0x09, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		}

		serialize.MarshalPaddedString(id.SerialNumber[:], s.SerialNumber)
		serialize.MarshalPaddedString(id.ModelNumber[:], s.ModelName)
		serialize.MarshalPaddedString(id.FirmwareRevision[:], s.FirmwareVersion)
		sm.Serialize(&id)

	case protocol.CNSIdentifyActiveNamespaces:
		// Return a list of namespaces currently enabled
		//   right now we only support 1 namespace
		id := protocol.IdentifyActiveNamespaceListData{}
		id.CNS[0] = 0x1 // just our 1 namespace
		sm.Serialize(&id)

	case protocol.CNSIdentifyNamespaceDescriptorList:
		id := protocol.IdentifyNamespaceDescriptor{}
		id.NIDT = 0x3 // UUID
		id.NIDL = 16  // 16 bytes
		copy(id.NID[:], s.UUID[:])
		sm.Serialize(&id)

	case protocol.CNSIdentifyControlerDataStructures:
		return nil, fmt.Errorf("identify with unsupported cns:0x%x for subsystem:target", cns)

	default:
		return nil, fmt.Errorf("identify with unsupported cns:0x%x for subsystem:target", cns)
	}

	return sm.Get(), nil
}

func (s *TargetSubsystem) GetLogPage(pageId int, offset uint64, length int) ([]byte, error) {
	ss := serialize.New(make([]byte, length, length))
	switch pageId {
	case protocol.LPErrorInformation:
		// no error information

	case protocol.LPHealthInformation:
		// SMART

	case protocol.LPCommandsSupported: // Commands Supported and Effects (Log Identifier 05h)
		lp := protocol.CommandsSupportedLogPage{}

		lp.ACS[protocol.CapsuleCmdDeleteQueue] = 0x01
		lp.ACS[protocol.CapsuleCmdAbort] = 0x01
		lp.ACS[protocol.CapsuleCmdIdentify] = 0x01
		lp.ACS[protocol.CapsuleCmdSetFeatures] = 0x01
		lp.ACS[protocol.CapsuleCmdGetFeatures] = 0x01
		lp.ACS[protocol.CapsuleCmdAsyncEventRequest] = 0x01
		lp.ACS[protocol.CapsuleCmdKeepAlive] = 0x01
		lp.IOCS[protocol.CapsuleCmdFlush] = 0x01
		lp.IOCS[protocol.CapsuleCmdWrite] = 0x01
		lp.IOCS[protocol.CapsuleCmdRead] = 0x01
		lp.IOCS[protocol.CapsuleCmdWriteZeros] = 0x01
		lp.IOCS[protocol.CapsuleCmdDatasetMgmt] = 0x01

		ss.Serialize(&lp)

	case protocol.LPDeviceSelfTest:

	case protocol.LPAsymmetricNamespaceAccess: // Asymmetric Namespace Access (Log Identifier 0Ch)
		lp := protocol.AsymmetricNamespaceAccessLog{
			DescriptorCount: 1,
		}

		lp.ANAGroupDesc.ANAGroupID = 1
		lp.ANAGroupDesc.NSIDCount = 1
		lp.ANAGroupDesc.ChangeCount = 0
		lp.ANAGroupDesc.ANAS = 0x01 // State = ANA Optimized state
		lp.ANAGroupDesc.NID[0] = 0x1
		ss.Serialize(&lp)

	default:
		return nil, fmt.Errorf("log page 0x%x not supported by target subsystem", pageId)
	}
	return ss.Get(), nil
}

func (s *TargetSubsystem) HandleIO(r *targets.IORequest) targets.TargetError {
	return s.Target.Queue(r)
}

func (s *TargetSubsystem) GetRuntimeDetails() []targets.KV {
	return s.Target.GetRuntimeDetails()
}

func (s *TargetSubsystem) QueueIO(r *targets.IORequest) targets.TargetError {
	return s.Target.Queue(r)
}
