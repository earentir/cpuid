package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

// FeatureSet defines a group of CPU features and how to query them
type FeatureSet struct {
	name      string            // Display name
	leaf      uint32            // CPUID leaf (eax input)
	subleaf   uint32            // CPUID subleaf (ecx input)
	register  int               // Which register to use (0=EAX, 1=EBX, 2=ECX, 3=EDX)
	condition func(uint32) bool // Optional condition function
	group     string            // Group name
	features  map[int]Feature   // Feature map
}

// Feature represents a CPU feature with its description and function
type Feature struct {
	name                  string
	description           string
	function              string
	vendor                string // "amd", "intel", or "common"
	equivalentFeatureName string // equivalent feature name here
	equivalent            int    // equivalent int here
}

var (
	vendorID   = getVendorID()
	isAMD      = strings.Contains(strings.ToUpper(vendorID), "AMD")
	extb       uint32
	maxFunc    uint32
	maxExtFunc uint32
)

// All CPU features are stored here
//var cpuFeaturesList is in vars.go

func cpuid(eax, ecx uint32) (a, b, c, d uint32)

func init() {
	maxFunc, maxExtFunc = getMaxFunctions()
	vendorID = getVendorID()
	isAMD = strings.Contains(strings.ToUpper(vendorID), "AMD")
}

func printBasicInfo() {
	fmt.Printf("Running on %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Detecting CPU features for x86/x64...\n")
	fmt.Printf("CPUID Max Standard Function: %d\n", maxFunc)
	fmt.Printf("CPUID Max Extended Function: 0x%08x\n", maxExtFunc)
	fmt.Printf("CPU Vendor ID: %s\n", vendorID)

	isIntel := strings.Contains(strings.ToUpper(vendorID), "INTEL")

	// Get processor info
	a, b, _, _ := cpuid(1, 0)
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

	fmt.Printf("\nProcessor Details:\n")
	fmt.Printf("  Stepping ID: %d\n", steppingID)
	fmt.Printf("  Model: %d (0x%x)\n", effectiveModel, effectiveModel)
	fmt.Printf("  Family: %d (0x%x)\n", effectiveFamily, effectiveFamily)
	fmt.Printf("  Processor Type: %d\n", processorType)

	// Get brand string
	if maxExtFunc >= 0x80000004 {
		var brand [48]byte
		for i := 0; i < 3; i++ {
			a, b, c, d := cpuid(0x80000002+uint32(i), 0)
			copy(brand[i*16:], int32ToBytes(a))
			copy(brand[i*16+4:], int32ToBytes(b))
			copy(brand[i*16+8:], int32ToBytes(c))
			copy(brand[i*16+12:], int32ToBytes(d))
		}
		fmt.Printf("  Brand String: %s\n", strings.TrimSpace(string(brand[:])))
	}

	fmt.Printf("\nCache Information:\n")

	// AMD Cache Detection
	if isAMD && maxExtFunc >= 0x8000001D {
		for i := uint32(0); ; i++ {
			a, b, c, _ := cpuid(0x8000001D, i)
			cacheType := a & 0x1F
			if cacheType == 0 {
				break
			}
			level := (a >> 5) & 0x7
			lineSize := (b & 0xFFF) + 1
			partitions := ((b >> 12) & 0x3FF) + 1
			associativity := ((b >> 22) & 0x3FF) + 1
			sets := c + 1
			size := lineSize * partitions * associativity * sets

			typeString := getCacheTypeString(cacheType)

			fmt.Printf("  L%d %s Cache:\n", level, typeString)
			fmt.Printf("    Size: %d KB\n", size/1024)
			fmt.Printf("    Ways: %d\n", associativity)
			fmt.Printf("    Line Size: %d bytes\n", lineSize)
			fmt.Printf("    Total Sets: %d\n", sets)

			maxCoresSharing := ((a >> 14) & 0xFFF) + 1
			fmt.Printf("    Max Cores Sharing: %d\n", maxCoresSharing)

			printCacheFlags(a)
		}
		// Intel Cache Detection
	} else if isIntel && maxFunc >= 4 {
		for i := uint32(0); ; i++ {
			a, b, c, _ := cpuid(4, i)
			cacheType := a & 0x1F
			if cacheType == 0 {
				break
			}
			level := (a >> 5) & 0x7
			lineSize := (b & 0xFFF) + 1
			partitions := ((b >> 12) & 0x3FF) + 1
			associativity := ((b >> 22) & 0x3FF) + 1
			sets := c + 1
			size := lineSize * partitions * associativity * sets

			typeString := getCacheTypeString(cacheType)

			fmt.Printf("  L%d %s Cache:\n", level, typeString)
			fmt.Printf("    Size: %d KB\n", size/1024)
			fmt.Printf("    Ways: %d\n", associativity)
			fmt.Printf("    Line Size: %d bytes\n", lineSize)
			fmt.Printf("    Total Sets: %d\n", sets)

			maxCoresSharing := ((a >> 14) & 0xFFF) + 1
			fmt.Printf("    Max Cores Sharing: %d\n", maxCoresSharing)

			printCacheFlags(a)
		}
	}

	fmt.Printf("\nTLB Information:\n")
	if isAMD && maxExtFunc >= 0x80000005 {
		printAMDTLBInfo()
	} else if isIntel {
		printIntelTLBInfo()
	}

	fmt.Printf("\nBasic Features:\n")
	_, b, _, _ = cpuid(1, 0)
	fmt.Printf("  Max Logical Processors: %d\n", (b>>16)&0xFF)
	fmt.Printf("  Initial APIC ID: %d\n", (b>>24)&0xFF)

	// Physical address and linear address bits
	if maxExtFunc >= 0x80000008 {
		a, _, c, _ := cpuid(0x80000008, 0)
		fmt.Printf("  Physical Address Bits: %d\n", a&0xFF)
		fmt.Printf("  Linear Address Bits: %d\n", (a>>8)&0xFF)

		// AMD specific core count info
		if isAMD {
			coreCount := (c & 0xFF) + 1
			fmt.Printf("  Core Count: %d\n", coreCount)
			threadPerCore := ((c >> 8) & 0xFF) + 1
			fmt.Printf("  Threads per Core: %d\n", threadPerCore)
		}
	}

	// Additional Intel-specific information
	if isIntel && maxFunc >= 0x1A {
		a, _, _, _ := cpuid(0x1A, 0)
		if a&1 != 0 {
			fmt.Printf("\nHybrid Information:\n")
			fmt.Printf("  Hybrid CPU: Yes\n")
			fmt.Printf("  Native Model ID: %d\n", (a>>24)&0xFF)
			fmt.Printf("  Core Type: %d\n", (a>>16)&0xFF)
		}
	}

	fmt.Printf("\nFeature Information:\n")
	fmt.Printf("\nFeature Information:\n")
	_, _, c, d := cpuid(1, 0) // Get standard features
	fmt.Printf("\nStandard Features ECX:\n")
	printFeatureFlags(cpuFeaturesList["StandardECX"].features, c)
	fmt.Printf("\nStandard Features EDX:\n")
	printFeatureFlags(cpuFeaturesList["StandardEDX"].features, d)

	if maxFunc >= 7 {
		_, b, c, d := cpuid(7, 0) // Get extended features
		fmt.Printf("\nExtended Features EBX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedEBX"].features, b)
		fmt.Printf("\nExtended Features ECX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedECX"].features, c)
		fmt.Printf("\nExtended Features EDX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedEDX"].features, d)
	}

	if isAMD && maxExtFunc >= 0x80000001 {
		_, _, c, _ := cpuid(0x80000001, 0) // Get AMD extended features
		fmt.Printf("\nAMD Extended Features ECX:\n")
		printFeatureFlags(cpuFeaturesList["AMDExtendedECX"].features, c)
	}
}

func getCacheTypeString(cacheType uint32) string {
	switch cacheType {
	case 1:
		return "Data"
	case 2:
		return "Instruction"
	case 3:
		return "Unified"
	default:
		return "Unknown"
	}
}

func printCacheInfo(level uint32, typeString string, size, associativity, lineSize, sets uint32) {
	fmt.Printf("  L%d %s Cache:\n", level, typeString)
	fmt.Printf("    Size: %d KB\n", size/1024)
	fmt.Printf("    Ways: %d\n", associativity)
	fmt.Printf("    Line Size: %d bytes\n", lineSize)
	fmt.Printf("    Total Sets: %d\n", sets)
}

func printCacheFlags(a uint32) {
	if (a>>9)&1 == 1 {
		fmt.Printf("    Fully Associative: Yes\n")
	}
	if (a>>10)&1 == 1 {
		fmt.Printf("    Write Back: Yes\n")
	}
	if (a>>11)&1 == 1 {
		fmt.Printf("    Inclusive: Yes\n")
	}
}

func printAMDTLBInfo() {
	a, b, _, _ := cpuid(0x80000005, 0)
	fmt.Printf("  L1 Data TLB:\n")
	fmt.Printf("    2MB/4MB Pages: %d entries\n", (a>>16)&0xFF)
	fmt.Printf("    4KB Pages: %d entries\n", (a>>24)&0xFF)
	fmt.Printf("  L1 Instruction TLB:\n")
	fmt.Printf("    2MB/4MB Pages: %d entries\n", (b>>16)&0xFF)
	fmt.Printf("    4KB Pages: %d entries\n", (b>>24)&0xFF)

	if maxExtFunc >= 0x80000006 {
		a, b, _, _ = cpuid(0x80000006, 0)
		fmt.Printf("  L2 Data TLB:\n")
		fmt.Printf("    2MB/4MB Pages: %d entries\n", (a>>16)&0xFFF)
		fmt.Printf("    4KB Pages: %d entries\n", (a>>28)&0xF)
		fmt.Printf("  L2 Instruction TLB:\n")
		fmt.Printf("    2MB/4MB Pages: %d entries\n", (b>>16)&0xFFF)
		fmt.Printf("    4KB Pages: %d entries\n", (b>>28)&0xF)
	}
}

func printIntelTLBInfo() {
	if maxFunc >= 0x2 {
		a, b, c, d := cpuid(0x2, 0)
		fmt.Printf("  Intel TLB information available through descriptor bytes\n")
		fmt.Printf("  Raw descriptor values: EAX=%08x EBX=%08x ECX=%08x EDX=%08x\n", a, b, c, d)
	}
}

func getMaxFunctions() (uint32, uint32) {
	a, _, _, _ := cpuid(0, 0)
	maxFunc := a

	a, _, _, _ = cpuid(0x80000000, 0)
	maxExtFunc := a

	return maxFunc, maxExtFunc
}

func int32ToBytes(i uint32) []byte {
	return []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
}

func printProcessorInfo() {
	// Get initial CPUID values
	a, _, _, _ := cpuid(1, 0)
	_, extb, _, _ = cpuid(7, 0)

	// Print basic processor info
	steppingID := a & 0xF
	modelID := (a >> 4) & 0xF
	familyID := (a >> 8) & 0xF
	processorType := (a >> 12) & 0x3
	extendedModelID := (a >> 16) & 0xF
	extendedFamilyID := (a >> 20) & 0xFF

	fmt.Printf("Processor Info:\n")
	fmt.Printf("  Stepping ID: %d\n", steppingID)
	fmt.Printf("  Model: %d\n", modelID+(extendedModelID<<4))
	fmt.Printf("  Family: %d\n", familyID+(extendedFamilyID<<4))
	fmt.Printf("  Processor Type: %d\n\n", processorType)

	vendorID := getVendorID()
	isAMD = strings.Contains(strings.ToUpper(vendorID), "AMD")

	// Print all feature sets
	for _, set := range cpuFeaturesList {
		// Skip if there's a condition and it's not met
		if set.condition != nil && !set.condition(0) {
			continue
		}

		// Get the register values for this leaf/subleaf
		a, b, c, d := cpuid(set.leaf, set.subleaf)

		// Get the correct register value based on the register index
		var regValue uint32
		switch set.register {
		case 0:
			regValue = a
		case 1:
			regValue = b
		case 2:
			regValue = c
		case 3:
			regValue = d
		}

		fmt.Printf("\n%s:\n", set.name)
		printFeatureFlags(set.features, regValue)
	}
}

// Helper function to get vendor ID
func getVendorID() string {
	_, b, c, d := cpuid(0, 0)
	return fmt.Sprintf("%s%s%s",
		string([]byte{byte(b), byte(b >> 8), byte(b >> 16), byte(b >> 24)}),
		string([]byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}),
		string([]byte{byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)}),
	)
}

// Modified printFeatureFlags to only show names
func printFeatureFlags(features map[int]Feature, reg uint32) []string {
	var recognized []string
	var unrecognized []string

	for i := 0; i < 32; i++ {
		if (reg>>i)&1 == 1 {
			if feature, exists := features[i]; exists {
				recognized = append(recognized, feature.name)
			} else {
				unrecognized = append(unrecognized, fmt.Sprintf("Bit %d", i))
			}
		}
	}

	sort.Strings(recognized)
	fmt.Printf("  %s\n", strings.Join(recognized, ", "))
	return unrecognized
}

func main() {
	fmt.Printf("Running on %s/%s\n", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOARCH {
	case "amd64", "386":
		fmt.Println("Detecting CPU features for x86/x64...")
		printBasicInfo()
		printProcessorInfo()
	default:
		fmt.Println("Unsupported architecture.")
	}
}
