// Package cpuid provides information about the CPU running the current program.
package cpuid

func cpuid(eax, ecx uint32) (eaxr, ebxr, ecxr, edxr uint32)

// cpuidoffline simulates the cpuid instruction using the data from the JSON file.
func cpuidoffline(eax, ecx uint32, filename string) (a, b, c, d uint32) {
	data, err := DataFromFile(filename)
	if err != nil {
		// If unable to load the data, return zeros.
		return 0, 0, 0, 0
	}

	// Search for a matching leaf and subleaf.
	for _, entry := range data.Entries {
		if entry.Leaf == eax && entry.Subleaf == ecx {
			return entry.EAX, entry.EBX, entry.ECX, entry.EDX
		}
	}
	// If not found, return zeros.
	return 0, 0, 0, 0
}

// CPUIDWithMode returns the result of the cpuid instruction for the given eax and ecx values.
func CPUIDWithMode(eax, ecx uint32, offline bool, filename string) (a, b, c, d uint32) {
	if !offline {
		// Call the live assembly implementation.
		return cpuid(eax, ecx)
	}

	// Use default filename if none provided.
	if filename == "" {
		filename = "cpuid_data.json"
	}

	return cpuidoffline(eax, ecx, filename)
}

// GetMaxFunctions returns the maximum standard and extended function values supported by the CPU.
func GetMaxFunctions(offline bool, filename string) (uint32, uint32) {
	a, _, _, _ := CPUIDWithMode(0, 0, offline, filename)
	maxFunc := a

	a, _, _, _ = CPUIDWithMode(0x80000000, 0, offline, filename)
	maxExtFunc := a

	return maxFunc, maxExtFunc
}

// GetIntelHybrid returns information about hybrid CPUs for Intel processors.
func GetIntelHybrid(offline bool, filename string) IntelHybridInfo {
	a, _, _, _ := CPUIDWithMode(0x1A, 0, offline, filename)

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
