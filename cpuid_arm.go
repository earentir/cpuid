//go:build arm || (arm64 && linux) || darwin
// +build arm arm64,linux darwin

package cpuid

import (
	"fmt"
	"strings"
)

// cpuid is implemented in assembly for ARM/ARM64 to read system registers.
// It returns four uint32 values corresponding to the CPU registers.
func cpuid(eax, ecx uint32) (a, b, c, d uint32)

// GetMaxFunctions returns dummy values since ARM does not support CPUID enumeration.
func GetMaxFunctions() (uint32, uint32) {
	return 1, 0
}

// GetVendorID returns a formatted string with CPU identification info extracted from MIDR.
// For ARMv7, the MIDR layout is:
//
//	Bits [31:24] Implementer
//	Bits [23:20] Variant
//	Bits [19:16] Architecture
//	Bits [15:4]  Part number
//	Bits [3:0]   Revision
func GetVendorID() string {
	a, _, _, _ := cpuid(0, 0)
	implementer := (a >> 24) & 0xff
	variant := (a >> 20) & 0xf
	arch := (a >> 16) & 0xf
	part := (a >> 4) & 0xfff
	revision := a & 0xf
	return fmt.Sprintf("Implementer: 0x%X, Variant: 0x%X, Arch: 0x%X, Part: 0x%X, Rev: 0x%X",
		implementer, variant, arch, part, revision)
}

// GetVendorName returns the vendor name based on the implementer field from MIDR.
// It also includes a check to detect Apple silicon by inspecting the part number.
// For example, many Apple M1 cores return a MIDR where:
//
//	Implementer == 0x41 (ARM Ltd.)
//	Part number    == 0xD03
//
// In unknown cases, the function returns a string with the raw register values.
func GetVendorName() string {
	a, _, _, _ := cpuid(0, 0)
	implementer := (a >> 24) & 0xff
	variant := (a >> 20) & 0xf
	architecture := (a >> 16) & 0xf
	part := (a >> 4) & 0xfff
	revision := a & 0xf

	// Check for Apple silicon.
	if implementer == 0x41 && part == 0xD03 {
		return "Apple"
	}

	switch implementer {
	case 0x41:
		return "ARM Ltd."
	case 0x42:
		return "Broadcom"
	case 0x43:
		return "Cavium"
	case 0x44:
		return "DEC"
	case 0x4E:
		return "NVIDIA"
	case 0x50:
		return "APM"
	case 0x51:
		return "Qualcomm"
	case 0x56:
		return "Marvell"
	default:
		return fmt.Sprintf("Unknown (Implementer: 0x%X, Variant: 0x%X, Arch: 0x%X, Part: 0x%X, Rev: 0x%X)",
			implementer, variant, architecture, part, revision)
	}
}

// isARM returns true if the vendor name contains "ARM" or "Apple".
func isARM() bool {
	name := strings.ToUpper(GetVendorName())
	return strings.Contains(name, "ARM") || strings.Contains(name, "APPLE")
}
