package protocol

import (
	"errors"
	"fmt"
)

type NVMEStatusCode uint16

// Controller status types masks
const (
	SCTypeGeneric uint16 = 0x0
	SCTypeCommand uint16 = 0x100
	SCTypeMedia   uint16 = 0x200
	SCTypePort    uint16 = 0x300
)

const (
	// Generic Status Codes
	SCSuccess                          NVMEStatusCode = 0x0
	SCInvalidCommandOpcode             NVMEStatusCode = 0x1
	SCInvalidFieldInCommand            NVMEStatusCode = 0x2
	SCCommandIDConflict                NVMEStatusCode = 0x3
	SCDataTransferError                NVMEStatusCode = 0x4
	SCCommandAbortedPowerLoss          NVMEStatusCode = 0x5
	SCInternalError                    NVMEStatusCode = 0x6
	SCAbortedRequest                   NVMEStatusCode = 0x7
	SCAbortedQueue                     NVMEStatusCode = 0x8
	SCFuseFailed                       NVMEStatusCode = 0x9
	SCFusesMissing                     NVMEStatusCode = 0xa
	SCInvalidNamespace                 NVMEStatusCode = 0xb
	SCCommandSequenceError             NVMEStatusCode = 0xc
	SCInvalidSGLDescriptor             NVMEStatusCode = 0xd
	SCInvalidSGLCount                  NVMEStatusCode = 0xe
	SCInvalidSGLData                   NVMEStatusCode = 0xf
	SCInvalidSGLMetadata               NVMEStatusCode = 0x10
	SCInvalidSGLType                   NVMEStatusCode = 0x11
	SCInvalidCMBUsage                  NVMEStatusCode = 0x12
	SCInvalidPRPOffset                 NVMEStatusCode = 0x13
	SCAtomicWriteUnitExceeded          NVMEStatusCode = 0x14
	SCOperationDenied                  NVMEStatusCode = 0x15
	SCInvalidSGLOffset                 NVMEStatusCode = 0x16
	SCHostIdentifierInconsistentFormat NVMEStatusCode = 0x18
	SCKeepAliveTimerExpired            NVMEStatusCode = 0x19
	SCKeepAliveTimeoutInvalid          NVMEStatusCode = 0x1A
	SCCommandAbortDueToPreempt         NVMEStatusCode = 0x1B
	SCSanitizeFailed                   NVMEStatusCode = 0x1C
	SCSanitizeInProgress               NVMEStatusCode = 0x1D
	SCInvalidSGLDataBlockGranularity   NVMEStatusCode = 0x1E
	SCCommandNotSupportedForQueue      NVMEStatusCode = 0x1F
	SCNamespaceWriteProtected          NVMEStatusCode = 0x20
	SCCommandInterrupted               NVMEStatusCode = 0x21
	SCTransientTransportError          NVMEStatusCode = 0x22
	SCLBAOutOfRange                    NVMEStatusCode = 0x80
	SCCapacityExceeded                 NVMEStatusCode = 0x81
	SCNamespaceNotReady                NVMEStatusCode = 0x82
	SCReservationConflict              NVMEStatusCode = 0x83
	SCFormatInProgress                 NVMEStatusCode = 0x84

	SCCmdFeatureNotChangeable NVMEStatusCode = 0x10e

	SCMediaWriteFault                  NVMEStatusCode = 0x280
	SCMediaUncorrectableReadError      NVMEStatusCode = 0x281
	SCMediaE2EGuardCheckError          NVMEStatusCode = 0x282
	SCMediaE2EApplicationTagCheckError NVMEStatusCode = 0x283
	SCMediaE2EReferenceTagCheckError   NVMEStatusCode = 0x284
	SCMediaCompareFailure              NVMEStatusCode = 0x285
	SCMediaAccessDenied                NVMEStatusCode = 0x286
	SCMediaDeallocatedLogicalBlock     NVMEStatusCode = 0x287

	SCInternalPathError NVMEStatusCode = 0x300
	SCANAPersistentLoss NVMEStatusCode = 0x301
	SCANAInaccessible   NVMEStatusCode = 0x302
	SCANATransition     NVMEStatusCode = 0x303

	SCControllerPathError NVMEStatusCode = 0x360
	SCHostPathError       NVMEStatusCode = 0x370
	SCHostAbortedCommand  NVMEStatusCode = 0x371

	// Vendor Errors
	SCConnectionFailure NVMEStatusCode = 0x701
	SCInvalidQueueId    NVMEStatusCode = 0x702

	// Flag bits
	SCFlagDoNotRetry NVMEStatusCode = 0x8000
)

var genericStatusToString = map[NVMEStatusCode]string{
	// Generic Command Status Definition (Figure 126,127)
	SCSuccess:                          "sucessfull completion",
	SCInvalidCommandOpcode:             "invalid command opcode",
	SCInvalidFieldInCommand:            "invalid field in command",
	SCCommandIDConflict:                "command ID conflice",
	SCDataTransferError:                "data transfer error",
	SCCommandAbortedPowerLoss:          "command aborted due to power loss notification",
	SCInternalError:                    "internal error",
	SCAbortedRequest:                   "command abort requested",
	SCAbortedQueue:                     "command aborted due to SQ deleteion",
	SCFuseFailed:                       "command aboted due to failed fused commanf",
	SCFusesMissing:                     "command aborted due to missing fused commanf",
	SCInvalidNamespace:                 "invalid namespace format",
	SCCommandSequenceError:             "command sequence error",
	SCInvalidSGLDescriptor:             "invalid SGL segment descriptor",
	SCInvalidSGLCount:                  "invalid number of SGL descriptors",
	SCInvalidSGLData:                   "Data SGL length invalid",
	SCInvalidSGLMetadata:               "MetadataSGL length invalid",
	SCInvalidSGLType:                   "SGL descriptor invalid type",
	SCInvalidCMBUsage:                  "invalid use of controller memory buffer",
	SCInvalidPRPOffset:                 "PRP offset invalid",
	SCAtomicWriteUnitExceeded:          "atomic write unit exceeded",
	SCOperationDenied:                  "operation denied", // security violation
	SCInvalidSGLOffset:                 "SGL offset invalid",
	NVMEStatusCode(0x17):               "reserved",
	SCHostIdentifierInconsistentFormat: "host identifier inconsistent format",
	SCKeepAliveTimerExpired:            "keep alive timer expired",
	SCKeepAliveTimeoutInvalid:          "keep alive timeout invalid",
	SCCommandAbortDueToPreempt:         "command aborted due to preempt and abort",
	SCSanitizeFailed:                   "sanitize failed",
	SCSanitizeInProgress:               "sanitize in progress",
	SCInvalidSGLDataBlockGranularity:   "SDL data block granularity invalid",
	SCCommandNotSupportedForQueue:      "command not supported for queue in CMB",
	SCNamespaceWriteProtected:          "namespace is write protected",
	SCCommandInterrupted:               "command interrupted",
	SCTransientTransportError:          "transient transport error",
	SCLBAOutOfRange:                    "lba out of range",
	SCCapacityExceeded:                 "capacity exceeded",
	SCNamespaceNotReady:                "namespace not ready",
	SCReservationConflict:              "reservation conflict",
	SCFormatInProgress:                 "format in progress",

	// Command Specific Status Definition (Figure 128,129)
	SCCmdFeatureNotChangeable: "feature not changeable",

	// Media Specific Status Definition (Figure 130,131)
	SCMediaWriteFault:                  "write fault",
	SCMediaUncorrectableReadError:      "unrecoverable read error",
	SCMediaE2EGuardCheckError:          "end to end guard check error",
	SCMediaE2EApplicationTagCheckError: "end to end application tag check error",
	SCMediaE2EReferenceTagCheckError:   "end to end reference tag check error",
	SCMediaCompareFailure:              "compare failure",
	SCMediaAccessDenied:                "access denied",
	SCMediaDeallocatedLogicalBlock:     "deallocated or unwritten logical block",

	// Path Related Status Definitions (Figure 123)
	SCInternalPathError:   "internal path error",
	SCANAPersistentLoss:   "asymmetric access persistent loss",
	SCANAInaccessible:     "asymmetric access inaccessible",
	SCANATransition:       "asymmetric access transition",
	SCControllerPathError: "controller pathing error",
	SCHostPathError:       "host pathing error",
	SCHostAbortedCommand:  "command aborted by host",

	// Vendor Specific Errors ( These are errors reported by this target/client code)
	SCConnectionFailure: "client fabric connection failure",
	SCInvalidQueueId:    "client invalid queue id",
}

func (nsc NVMEStatusCode) String() string {
	v := nsc & 0x3FFF

	s, ok := genericStatusToString[v]
	if ok {
		return s
	}
	return fmt.Sprintf("nvme error: %xh", uint16(nsc))
}

func (nsc NVMEStatusCode) IsError() bool {
	return nsc != SCSuccess
}

func (nsc NVMEStatusCode) Code() NVMEStatusCode {
	return nsc & 0x3FFF
}

func (nsc NVMEStatusCode) AsError() error {
	if nsc == SCSuccess {
		return nil
	}
	return errors.New(nsc.String())
}
