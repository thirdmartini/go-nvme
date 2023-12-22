package protocol

type ControllerLogPage struct {
}

// IdentifyController implements 5.15.2.2 Identify Controller data structure (CNS 01h)
type IdentifyController struct {
	PCIVendor           uint16   `offset:"0"`
	PCIDevice           uint16   `offset:"2"`
	SerialNumber        [20]byte `offset:"4" length:"20"`
	ModelNumber         [40]byte `offset:"24" length:"40"`
	FirmwareRevision    [8]byte  `offset:"64" length:"8"`
	RAB                 uint8    `offset:"72"`
	OUI                 [3]byte  `offset:"73" length:"3"`
	CMIC                uint8    `offset:"76"`
	MDTS                uint8    `offset:"77"`
	ControllerId        uint16   `offset:"78"`
	Version             uint32   `offset:"80"`
	OAES                uint32   `offset:"92"`
	CTRATT              uint32   `offset:"96"`
	RRLS                uint16   `offset:"100"`
	CNTRLTYPE           uint8    `offset:"111"`
	FGUID               [16]byte `offset:"112"`
	OACS                uint16   `offset:"256"`
	ACL                 uint8    `offset:"258"`
	AERL                uint8    `offset:"259"`
	FRWM                uint8    `offset:"260"`
	LogPageAttributes   uint8    `offset:"261"`
	ErrorLogPageEntries uint8    `offset:"262"`
	NPSS                uint8    `offset:"263"`
	AVSCC               uint8    `offset:"264"`
	APSTA               uint8    `offset:"265"`
	WCTEMP              uint16   `offset:"266"`
	CCTEMP              uint16   `offset:"268"`
	TNVMCAP0            uint64   `offset:"280"` // in BYTES!!! ( 128 bits of size )
	TNVMCAP1            uint64   `offset:"288"`
	KAS                 uint16   `offset:"320"`
	ANATT               uint8    `offset:"342"`
	ANACAP              uint8    `offset:"343"`
	ANAGRPMAX           uint32   `offset:"344"`
	NANAGRPID           uint32   `offset:"348"`
	SQES                uint8    `offset:"512"`
	CQES                uint8    `offset:"513"`
	MaxCMDS             uint16   `offset:"514"`
	NumberNamespaces    uint32   `offset:"516"`
	ONCS                uint16   `offset:"520"`
	FUSES               uint16   `offset:"522"`
	FNA                 uint8    `offset:"524"`
	VWC                 uint8    `offset:"525"`
	AWUN                uint16   `offset:"526"`
	AWUPF               uint16   `offset:"528"`
	NWPC                uint8    `offset:"531"` // Namespace Write Protection Capabilities
	ACWU                uint16   `offset:"532"`
	MNAN                uint32   `offset:"540"` //  Maximum Number of Allowed Namespaces
	SGLSupport          uint32   `offset:"536"`
	SubNQN              string   `offset:"768" length:"256"`
	// NVMe over Fabrics required fields
	IOCCSZ uint32 `offset:"1792"`
	IORCSZ uint32 `offset:"1796"`
	ICDOFF uint16 `offset:"1800"`
	FCATT  uint8  `offset:"1802"`
	MSDBD  uint8  `offset:"1803"`
	OFCS   uint16 `offset:"1804"`

	// we only support a single psd
	PSD [32]uint8 `offset:"2048" length:"32"`
}

// IdentifyNamespaceData implements 5.15.2.1 Identify Namespace data structure (CNS 00h)
type IdentifyNamespaceData struct {
	NSZE   uint64    `offset:"0"`
	NCAP   uint64    `offset:"8"`
	NUSE   uint64    `offset:"16"`
	NSFEAT uint8     `offset:"24"`
	NLBAF  uint8     `offset:"25"`
	FLBAS  uint8     `offset:"26"`
	MC     uint8     `offset:"27"`
	DPC    uint8     `offset:"28"`
	DPS    uint8     `offset:"29"`
	NMIC   uint8     `offset:"30"`
	RESCAP uint8     `offset:"31"`
	FPI    uint8     `offset:"32"`
	DLFEAT uint8     `offset:"33"`
	NAWUN  uint16    `offset:"34"`
	NAWUPF uint16    `offset:"36"`
	NACWU  uint16    `offset:"38"`
	NABSN  uint16    `offset:"40"`
	NABO   uint16    `offset:"42"`
	NABSPF uint16    `offset:"44"`
	NOIOB  uint16    `offset:"46"`
	NVMCAP [2]uint64 `offset:"48"` // 128 byte
	NPWG   uint16    `offset:"64"`
	NPWA   uint16    `offset:"66"`
	NPDG   uint16    `offset:"68"`
	NPDA   uint16    `offset:"70"`
	NOWS   uint16    `offset:"72"`
	// 91:74 Reserved
	ANAGRPID uint32 `offset:"92"`
	// 98:96 Reserved
	NSATTR   uint8  `offset:"99"`
	NVMSETID uint16 `offset:"100"`
	ENDGID   uint16 `offset:"102"`

	// This field uses the EUI-64 based 16-byte designator format. Bytes 114:112 contain the
	// 24-bit Organizationally Unique Identifier (OUI) value assigned by the IEEE Registration
	// Authority. Bytes 119:115 contain an extension identifier assigned by the corresponding
	// organization. Bytes 111:104 contain the vendor specific extension identifier assigned by
	// the corresponding organization. Refer to the IEEE EUI-64 guidelines for more
	// information. This field is big endian (refer to section 7.10.5).
	NGUID [16]uint8 `offset:"104"`
	//This field contains a 64-bit IEEE Extended
	//Unique Identifier (EUI-64) that is globally unique and assigned to the namespace when
	//the namespace is created. This field remains fixed throughout the life of the namespace
	//and is preserved across namespace and controller operations
	EUI64 uint64     `offset:"120"`
	LBAF  [16]uint32 `offset:"128" length:"64"`
}

type IdentifyActiveNamespaceListData struct {
	CNS [1024]uint32 `offset:"0" length:"1024"`
}

type IdentifyNamespaceDescriptor struct {
	NIDT uint8 `offset:"0"`
	NIDL uint8 `offset:"1"`
	// FIXME: technically this is broken
	//  this SHOULD be 04 - NIDL+3 ( IE offset 4, length NIDL )
	//  but we're going to use 16 always since we're storing uuids
	NID [16]byte `offset:"4" length:"16"`
}

type GetLogPageCommand struct {
	OpCode         uint8  `offset:"0"`
	Flags          uint8  `offset:"1"`
	CommendId      uint8  `offset:"2"`
	NamespaceId    uint32 `offset:"4"`
	LogPageId      uint8  `offset:"40"` // D10
	LogSp          uint8  `offset:"41"` // D10
	NumDWordsLower uint16 `offset:"42"` // D10 LOW
	NumDWordsUpper uint16 `offset:"44"` // D11

	D12 uint32 `offset:"48"` //  Logpage offset lower
	D13 uint32 `offset:"52"` //  Logpage offset higher
	D14 uint32 `offset:"56"`
	D15 uint32 `offset:"60"`
}

func (p *GetLogPageCommand) GetReturnBufferLength() uint32 {
	// The counter is zero based where 0 -> 1
	return 4 * ((uint32(p.NumDWordsUpper) | uint32(p.NumDWordsLower)) + 1)
}

func (p *GetLogPageCommand) GetReturnOffset() uint64 {
	// The counter is zero based where 0 -> 1
	return uint64(p.D12) | uint64(p.D13)<<32
}

type AsymmetricNamespaceAccessLog struct {
	ChangeCount     uint64 `offset:"0"`
	DescriptorCount uint16 `offset:"8"`
	// FIXME this is a marshalled list of ANAGroupDescriptors which are of variable size
	//  but we're cheating since we only have one
	ANAGroupDesc ANAGroupDescriptor `offset:"16" length:""`
}

type ANAGroupDescriptor struct {
	ANAGroupID  uint32    `offset:"0"`
	NSIDCount   uint32    `offset:"4"`
	ChangeCount uint64    `offset:"8"`
	ANAS        uint8     `offset:"16"`
	NID         [1]uint32 `offset:"32" length:"1"`
}

type DiscoveryLogPageEntry struct {
	TransportType         uint8  `offset:"0"`
	AddressFamily         uint8  `offset:"1"`
	SubsystemType         uint8  `offset:"2"`
	TransportRequirements uint8  `offset:"3"`
	PortId                uint16 `offset:"4"`
	ControllerId          uint16 `offset:"6"`
	AdminMaxQueueSize     uint16 `offset:"8"`
	TransportServiceId    string `offset:"32" length:"32"`
	SubNQN                string `offset:"256" length:"256"`
	TransportAddress      string `offset:"512" length:"256"`
	TSAS                  string `offset:"768" length:"256"`
}

type DiscoveryLogPage struct {
	GenerationCounter   uint64                  `offset:"0"`
	NumberOfRecords     uint64                  `offset:"8"`
	RecordFormat        uint64                  `offset:"16"`
	DiscoveryLofEntries []DiscoveryLogPageEntry `offset:"1024" step:"1024"`
}

// DiscoveryLogPageData is raw data that gets returned to the initiator in a discovery request
//
//	initially we return DiscoveryLogPage when offset of 0 is requested, but then initiator will request follow un data
//	in size increments at a given offset
type DiscoveryLogPageData struct {
	DiscoveryLofEntries []DiscoveryLogPageEntry `offset:"0" step:"1024"`
}

// CommandsSupportedLogPage 5.14.1.5 Commands Supported and Effects (Log Identifier 05h)
type CommandsSupportedLogPage struct {
	ACS  [256]uint32 `offset:"0" length:"256"`
	IOCS [256]uint32 `offset:"256" length:"256"`
}
