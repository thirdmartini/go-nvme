package protocol

const (
	MaxHeaderSize = 128
	MaxH2CPDUSize = 0x8000
)

// var NVMESpecificationVersion = uint32(1<<16 | 3<<8 | 0)
var NVMESpecificationVersion = uint32(1<<16 | 3<<8 | 0)

// NVME/TCP headers
const (
	ICReq       = 0x0
	ICResp      = 0x1
	H2CTermReq  = 0x2
	C2HTermReq  = 0x3
	CapsuleCmd  = 0x4
	CapsuleResp = 0x5
	C2HData     = 0x7
	R2T         = 0x9
)

var headerTypeToString = map[uint8]string{
	ICReq:       "ICReq",
	ICResp:      "ICResp",
	H2CTermReq:  "H2CTermReq",
	C2HTermReq:  "C2HTermReq",
	CapsuleCmd:  "CapsuleCmd",
	CapsuleResp: "CapsuleResp",
	C2HData:     "C2HData",
	R2T:         "R2T",
}

func HeaderTypeToString(t uint8) string {
	if s, ok := headerTypeToString[t]; ok {
		return s
	}
	return "--"
}

// Controller attributes
const (
	NVMECtrlAttrMaxQueueSize = 64
	NVMECtrlMaxCmds          = NVMECtrlAttrMaxQueueSize
)

// NVM-Express-1_4-2019.06.10-Ratified.pdf
// 5 Admin Command Set
const (
	// Figure 139: Opcodes for Admin Command
	CapsuleCmdDeleteQueue              = 0x00
	CapsuleCmdCreateQueue              = 0x01
	CapsuleCmdGetLogPage               = 0x02
	CapsuleCmdDeleteCompletionQueue    = 0x04
	CapsuleCmdCreateCompletionQueue    = 0x05
	CapsuleCmdIdentify                 = 0x06
	CapsuleCmdAbort                    = 0x08
	CapsuleCmdSetFeatures              = 0x09
	CapsuleCmdGetFeatures              = 0x0A
	CapsuleCmdAsyncEventRequest        = 0x0C
	CapsuleCmdNamespaceManagement      = 0x0D
	CapsuleCmdFirmwareCommand          = 0x10
	CapsuleCmdFirmwareDownload         = 0x11
	CapsuleCmdDeviceSelfTest           = 0x14
	CapsuleCmdNamespaceAttachment      = 0x15
	CapsuleCmdKeepAlive                = 0x18
	CapsuleCmdDirectSent               = 0x19
	CapsuleCmdDirectReceive            = 0x1A
	CapsuleCmdVirtualizationManagement = 0x1C
	CapsuleCmdNVMEMISend               = 0x1D
	CapsuleCmdNVMEMIReceive            = 0x1E
	CapsuleCmdDoorbellBufferConfig     = 0x7C
	CapsuleCmdFabric                   = 0x7F
	CapsuleCmdSecurityRecv             = 0x82
	CapsuleCmdInvalid                  = 0xFF // test command should be treated as invalid by the target
)

var AdminCmdToString = map[uint8]string{
	CapsuleCmdDeleteQueue:              "CapsuleCmdDeleteQueue",
	CapsuleCmdCreateQueue:              "CapsuleCmdCreateQueue",
	CapsuleCmdGetLogPage:               "CapsuleCmdGetLogPage",
	CapsuleCmdDeleteCompletionQueue:    "CapsuleCmdDeleteCompletionQueue",
	CapsuleCmdCreateCompletionQueue:    "CapsuleCmdCreateCompletionQueue",
	CapsuleCmdIdentify:                 "CapsuleCmdIdentify",
	CapsuleCmdAbort:                    "CapsuleCmdAbort",
	CapsuleCmdSetFeatures:              "CapsuleCmdSetFeatures",
	CapsuleCmdGetFeatures:              "CapsuleCmdGetFeatures",
	CapsuleCmdAsyncEventRequest:        "CapsuleCmdAsyncEventRequest",
	CapsuleCmdNamespaceManagement:      "CapsuleCmdNamespaceManagement",
	CapsuleCmdFirmwareCommand:          "CapsuleCmdFirmwareCommand",
	CapsuleCmdFirmwareDownload:         "CapsuleCmdFirmwareDownload",
	CapsuleCmdDeviceSelfTest:           "CapsuleCmdDeviceSelfTest",
	CapsuleCmdNamespaceAttachment:      "CapsuleCmdNamespaceAttachment",
	CapsuleCmdKeepAlive:                "CapsuleCmdKeepAlive",
	CapsuleCmdDirectSent:               "CapsuleCmdDirectSent",
	CapsuleCmdDirectReceive:            "CapsuleCmdDirectReceive",
	CapsuleCmdVirtualizationManagement: "CapsuleCmdVirtualizationManagement",
	CapsuleCmdNVMEMISend:               "CapsuleCmdNVMEMISend",
	CapsuleCmdNVMEMIReceive:            "CapsuleCmdNVMEMIReceive",
	CapsuleCmdDoorbellBufferConfig:     "CapsuleCmdDoorbellBufferConfig",
	CapsuleCmdFabric:                   "CapsuleCmdFabric",
	CapsuleCmdInvalid:                  "CapsuleCmdInvalid-Test",
}

const (
	// 5.21.1 Feature Specific Information
	//  Figure 271: Set Features – Feature Identifiers
	FeatureLbaRangeType     = 0x03
	FeatureNumberOfQueues   = 0x07
	FeatureAsyncEventConfig = 0x0b
	FeatureTimestamp        = 0x0e
	FeatureKeepAliveTimer   = 0x0f
)

// NVME IO Capsule Commands
const (
	CapsuleCmdFlush               = 0x00
	CapsuleCmdWrite               = 0x01
	CapsuleCmdRead                = 0x02
	CapsuleCmdWriteUncorrectable  = 0x04
	CapsuleCmdCompare             = 0x05
	CapsuleCmdWriteZeros          = 0x08
	CapsuleCmdDatasetMgmt         = 0x09
	CapsuleCmdVerify              = 0x0c
	CapsuleCmdReservationRegister = 0x0d
	CapsuleCmdReservationReport   = 0x0e
	CapsuleCmdReservationAcquire  = 0x11
	CapsuleCmdReservationRelease  = 0x15
)

var IOCmdToString = map[uint8]string{
	CapsuleCmdFlush:               "CapsuleCmdFlush",
	CapsuleCmdWrite:               "CapsuleCmdWrite",
	CapsuleCmdRead:                "CapsuleCmdRead",
	CapsuleCmdWriteUncorrectable:  "CapsuleCmdWriteUncorrectable",
	CapsuleCmdCompare:             "CapsuleCmdCompare",
	CapsuleCmdWriteZeros:          "CapsuleCmdWriteZeros",
	CapsuleCmdDatasetMgmt:         "CapsuleCmdDatasetMgmt",
	CapsuleCmdVerify:              "CapsuleCmdVerify",
	CapsuleCmdReservationRegister: "CapsuleCmdReservationRegister",
	CapsuleCmdReservationReport:   "CapsuleCmdReservationReport",
	CapsuleCmdReservationAcquire:  "CapsuleCmdReservationAcquire",
	CapsuleCmdReservationRelease:  "CapsuleCmdReservationRelease",
}

// NVME Fabric Level Command Set
const (
	FabricCmdPropertySet           = 0x0
	FabricCmdAuthenticationSend    = 0x5
	FabricCmdAuthenticationReceive = 0x6
	FabricCmdConnect               = 0x1
	FabricCmdPropertyGet           = 0x4
	FabricCmdDisconnect            = 0x8
	FabricCmdInvalid               = 0xff
)

var FabricCmdToString = map[uint8]string{
	FabricCmdPropertySet:           "FabricCmdPropertySet",
	FabricCmdAuthenticationSend:    "FabricCmdAuthenticationSend",
	FabricCmdAuthenticationReceive: "FabricCmdAuthenticationReceive",
	FabricCmdConnect:               "FabricCmdConnect",
	FabricCmdPropertyGet:           "FabricCmdPropertyGet",
	FabricCmdDisconnect:            "FabricCmdDisconnect",
	FabricCmdInvalid:               "FabricCmdInvalid",
}

func FCTypeToString(op uint8) string {
	if s, ok := FabricCmdToString[op]; ok {
		return s
	}
	return ""
}

const (
	RegisterControllerCapabilities  = 0x0
	RegisterControllerVersion       = 0x8
	RegisterControllerStatus        = 0x1c
	RegisterControllerConfiguration = 0x14
	RegisterControllerReset         = 0x20
)

const (
	CommandBitDeallocateSet = 1 << 25
)

// Identify CNS Pages
const (
	CNSIdentifyNamespace               = 0x00
	CNSIdentifyController              = 0x01
	CNSIdentifyActiveNamespaces        = 0x02
	CNSIdentifyNamespaceDescriptorList = 0x03
	CNSIdentifyControlerDataStructures = 0x06
)

const (
	LPErrorInformation          = 0x01
	LPHealthInformation         = 0x02
	LPFirmwareSlotInformation   = 0x03
	LPCommandsSupported         = 0x05
	LPDeviceSelfTest            = 0x06
	LPAsymmetricNamespaceAccess = 0x0c

	LPDiscovery = 0x70
)

const (
	PropertyControllerCapabilities  = 0x0
	PropertyVersion                 = 0x08
	PropertyControllerConfiguration = 0x14
	PropertyControllerStatus        = 0x1c
	PropertySubsystemReset          = 0x20
)
