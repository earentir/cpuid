// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"strings"
)

func cpuid(eax, ecx uint32) (a, b, c, d uint32)

// GetMaxFunctions returns the maximum standard and extended function values supported by the CPU.
func GetMaxFunctions() (uint32, uint32) {
	a, _, _, _ := cpuid(0, 0)
	maxFunc := a

	a, _, _, _ = cpuid(0x80000000, 0)
	maxExtFunc := a

	return maxFunc, maxExtFunc
}

func isAMD() bool {
	return strings.Contains(strings.ToUpper(GetVendorID()), "AMD")
}

func isIntel() bool {
	return strings.Contains(strings.ToUpper(GetVendorID()), "INTEL")
}

// GetIntelHybrid returns information about hybrid CPUs for Intel processors.
func GetIntelHybrid() IntelHybridInfo {
	a, _, _, _ := cpuid(0x1A, 0)

	if (a & 1) == 0 {
		// Not hybrid
		return IntelHybridInfo{HybridCPU: false}
	}

	hybridInfo := IntelHybridInfo{
		HybridCPU:     true,
		NativeModelID: (a >> 24) & 0xFF,
		CoreType:      (a >> 16) & 0xFF,
	}

	// Determine a human-readable core type
	switch hybridInfo.CoreType {
	case 1:
		hybridInfo.CoreTypeName = "Performance core (P-core)"
	case 2:
		hybridInfo.CoreTypeName = "Efficient core (E-core)"
	default:
		hybridInfo.CoreTypeName = "Unknown core type"
	}

	return hybridInfo
}

func int32ToBytes(i uint32) []byte {
	return []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
}

// Modified printFeatureFlags to only show names
// func printFeatureFlags(features map[int]Feature, reg uint32) []string {
// 	var recognized []string
// 	var unrecognized []string

// 	for i := 0; i < 32; i++ {
// 		if (reg>>i)&1 == 1 {
// 			if feature, exists := features[i]; exists {
// 				recognized = append(recognized, feature.name)
// 			} else {
// 				unrecognized = append(unrecognized, fmt.Sprintf("Bit %d", i))
// 			}
// 		}
// 	}

// 	sort.Strings(recognized)
// 	fmt.Printf("  %s\n", strings.Join(recognized, ", "))
// 	return unrecognized
// }
