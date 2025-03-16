// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"encoding/binary"
	"os"
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

// WriteCPUIDToFile calls cpuid with the provided eax and ecx values,
// then writes the returned register values to the specified file in binary format.
func WriteCPUIDToFile(filename string, eax, ecx uint32) error {
	// Call the assembly cpuid function.
	a, b, c, d := cpuid(eax, ecx)

	// Create (or truncate) the file.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write each register as a 32-bit little-endian value.
	if err := binary.Write(file, binary.LittleEndian, a); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, b); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, c); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, d); err != nil {
		return err
	}
	return nil
}

// fromFile opens the given file and reads the 4 register values
// in the same order they were written. It returns the values so you can use them
// as if you had run the cpuid command.
func fromFile(filename string) (a, b, c, d uint32, err error) {
	// Open the file for reading.
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	// Read the register values in little-endian order.
	if err = binary.Read(file, binary.LittleEndian, &a); err != nil {
		return
	}
	if err = binary.Read(file, binary.LittleEndian, &b); err != nil {
		return
	}
	if err = binary.Read(file, binary.LittleEndian, &c); err != nil {
		return
	}
	if err = binary.Read(file, binary.LittleEndian, &d); err != nil {
		return
	}
	return
}
