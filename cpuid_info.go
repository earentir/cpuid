// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"fmt"
	"strings"
)

// GetVendorID returns the vendor ID of the CPU.
func GetVendorID() string {
	_, b, c, d := cpuid(0, 0)
	return fmt.Sprintf("%s%s%s",
		string([]byte{byte(b), byte(b >> 8), byte(b >> 16), byte(b >> 24)}),
		string([]byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}),
		string([]byte{byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)}),
	)
}

// GetVendorName returns the vendor name of the CPU.
func GetVendorName() string {
	vendorID := GetVendorID()
	switch vendorID {
	case "GenuineIntel":
		return "Intel"
	case "AuthenticAMD":
		return "AMD"
	default:
		return "Unknown"
	}
}

// GetBrandString returns the brand string of the CPU.
func GetBrandString(maxExtFunc uint32) string {
	if maxExtFunc >= 0x80000004 {
		var brand [48]byte
		for i := 0; i < 3; i++ {
			a, b, c, d := cpuid(0x80000002+uint32(i), 0)
			copy(brand[i*16:], int32ToBytes(a))
			copy(brand[i*16+4:], int32ToBytes(b))
			copy(brand[i*16+8:], int32ToBytes(c))
			copy(brand[i*16+12:], int32ToBytes(d))
		}
		return strings.TrimSpace(string(brand[:]))
	}
	return ""
}

// GetModelData contains information about the processor model.
func GetModelData() ProcessorModel {
	// Get Model Data
	a, _, _, _ := cpuid(1, 0)
	steppingID := a & 0xF
	modelID := (a >> 4) & 0xF
	familyID := (a >> 8) & 0xF
	processorType := (a >> 12) & 0x3
	extendedModelID := (a >> 16) & 0xF
	extendedFamilyID := (a >> 20) & 0xFF

	// Calculate effective values
	effectiveModel := modelID
	if familyID == 0xF || familyID == 0x6 {
		effectiveModel += extendedModelID << 4
	}

	effectiveFamily := familyID
	if familyID == 0xF {
		effectiveFamily += extendedFamilyID
	}

	return ProcessorModel{
		fmt.Sprintf("%d", steppingID),
		fmt.Sprintf("%d", modelID),
		fmt.Sprintf("%d", familyID),
		fmt.Sprintf("%d", processorType),
		fmt.Sprintf("%d", extendedModelID),
		fmt.Sprintf("%d", extendedFamilyID),
		effectiveModel,
		effectiveFamily,
	}

}

// GetProcessorInfo returns detailed information about the CPU.
func GetProcessorInfo(maxFunc, maxExtFunc uint32) ProcessorInfo {
	//Basic Features
	_, b, _, _ := cpuid(1, 0)
	maxLogicalProcessors := (b >> 16) & 0xFF
	initialAPICID := (b >> 24) & 0xFF

	// Physical address and linear address bits
	var physicalAddressBits, linearAddressBits, coreCount, threadPerCore uint32
	if maxExtFunc >= 0x80000008 {
		a, _, _, _ := cpuid(0x80000008, 0)
		physicalAddressBits = a & 0xFF
		linearAddressBits = (a >> 8) & 0xFF
	}

	// Core and thread count detection
	if isAMD() {
		// For AMD CPUs using Extended Function 0x8000001E
		if maxExtFunc >= 0x8000001E {
			_, b, _, _ := cpuid(0x8000001E, 0)
			// Get threads per core
			threadPerCore = ((b >> 8) & 0xFF) + 1
			// Get total number of cores
			if maxExtFunc >= 0x80000008 {
				_, _, c, _ := cpuid(0x80000008, 0)
				coreCount = (c & 0xFF) + 1
			}
		} else if maxFunc >= 1 {
			// Fallback to basic CPUID information
			coreCount = ((maxLogicalProcessors + 1) / 2) // Assuming SMT is enabled
			threadPerCore = 2                            // Most modern AMD CPUs support 2 threads per core when SMT is enabled
		}
	} else if isIntel() {
		if maxFunc >= 0xB {
			// Use leaf 0xB for modern Intel CPUs
			var threadsPerCore, totalLogical uint32
			for subleaf := uint32(0); ; subleaf++ {
				_, b, c, _ := cpuid(0xB, subleaf)
				levelType := (c >> 8) & 0xFF
				if levelType == 0 {
					break
				}

				levelProcessors := b & 0xFFFF
				if levelType == 1 { // Thread level
					threadsPerCore = levelProcessors
				} else if levelType == 2 { // Core level
					totalLogical = levelProcessors
				}
			}

			if totalLogical > 0 && threadsPerCore > 0 {
				coreCount = totalLogical / threadsPerCore
				threadPerCore = threadsPerCore
			}
		}

		// Fallback for older Intel CPUs or if leaf 0xB didn't give valid results
		if coreCount == 0 {
			if maxFunc >= 4 {
				a, _, _, _ := cpuid(4, 0)
				coreCount = ((a >> 26) & 0x3F) + 1
				// Check if Hyper-Threading is enabled
				_, d, _, _ := cpuid(1, 0)
				if (d & (1 << 28)) != 0 { // HTT flag
					threadPerCore = 2
				} else {
					threadPerCore = 1
				}
			} else if maxFunc >= 1 {
				coreCount = 1
				// Check if Hyper-Threading is enabled
				_, d, _, _ := cpuid(1, 0)
				if (d & (1 << 28)) != 0 { // HTT flag
					threadPerCore = 2
				} else {
					threadPerCore = 1
				}
			}
		}
	}

	return ProcessorInfo{
		maxLogicalProcessors,
		initialAPICID,
		physicalAddressBits,
		linearAddressBits,
		coreCount,
		threadPerCore,
	}
}

// GotEnoughCores returns true if the CPU has enough cores to run the program.
func GotEnoughCores(coreCount uint32, realcores ...bool) bool {
	processorinfo := GetProcessorInfo(GetMaxFunctions())

	useRealCores := false
	if len(realcores) > 0 {
		useRealCores = realcores[0]
	}

	if useRealCores {
		return processorinfo.CoreCount >= coreCount
	}
	return processorinfo.MaxLogicalProcessors >= coreCount
}
