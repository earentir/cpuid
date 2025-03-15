// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"fmt"
	"sort"
	"strings"
)

func cpuid(eax, ecx uint32) (a, b, c, d uint32)

// GetProcessorInfo returns detailed information about the CPU.
func GetProcessorInfo(maxFunc, maxExtFunc uint32) ProcessorInfo {
	// Get processor info
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

	// Get brand string
	var brandString string
	if maxExtFunc >= 0x80000004 {
		var brand [48]byte
		for i := 0; i < 3; i++ {
			a, b, c, d := cpuid(0x80000002+uint32(i), 0)
			copy(brand[i*16:], int32ToBytes(a))
			copy(brand[i*16+4:], int32ToBytes(b))
			copy(brand[i*16+8:], int32ToBytes(c))
			copy(brand[i*16+12:], int32ToBytes(d))
		}
		brandString = strings.TrimSpace(string(brand[:]))
	}

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
		fmt.Sprintf("%d", steppingID),
		fmt.Sprintf("%d", modelID),
		fmt.Sprintf("%d", familyID),
		fmt.Sprintf("%d", processorType),
		fmt.Sprintf("%d", extendedModelID),
		fmt.Sprintf("%d", extendedFamilyID),
		effectiveModel,
		effectiveFamily,
		GetVendorID(),
		maxFunc,
		maxExtFunc,
		brandString,
		maxLogicalProcessors,
		initialAPICID,
		physicalAddressBits,
		linearAddressBits,
		coreCount,
		threadPerCore,
	}
}

func isAMD() bool {
	return strings.Contains(strings.ToUpper(GetVendorID()), "AMD")
}

func isIntel() bool {
	return strings.Contains(strings.ToUpper(GetVendorID()), "INTEL")
}

// GetAllFeatureCategories reports all categories
func GetAllFeatureCategories() []string {
	categories := make([]string, 0, len(cpuFeaturesList))
	for category := range cpuFeaturesList {
		categories = append(categories, category)
	}
	//sort categories
	sort.Strings(categories)

	return categories
}

// GetAllFeatureCategoriesDetailed returns all categories and their features with details.
func GetAllFeatureCategoriesDetailed() map[string][]map[string]string {
	details := make(map[string][]map[string]string)

	for _, fs := range cpuFeaturesList {
		categoryDetails := []map[string]string{}
		for _, feat := range fs.features {
			vendor := feat.vendor
			if vendor == "common" {
				vendor = "both"
			}

			entry := map[string]string{
				"name":        feat.name,
				"description": feat.description,
				"vendor":      vendor,
			}

			if feat.equivalentFeatureName != "" {
				entry["equivalent"] = feat.equivalentFeatureName
			}

			categoryDetails = append(categoryDetails, entry)
		}
		details[fs.name] = categoryDetails
	}

	return details
}

// GetAllKnownFeatures reports all known features
func GetAllKnownFeatures(category string) []string {
	fs, exists := cpuFeaturesList[category]
	if !exists {
		return nil
	}

	features := make([]string, 0, len(fs.features))
	for _, f := range fs.features {
		features = append(features, f.name)
	}
	return features
}

// GetSupportedFeatures reports all supported features
func GetSupportedFeatures(category string) []string {
	fs, exists := cpuFeaturesList[category]
	if !exists {
		return nil
	}

	// If there's a condition to check (some featuresets may only be valid if condition is met)
	if fs.condition != nil && !fs.condition(0) {
		return nil
	}

	a, b, c, d := cpuid(fs.leaf, fs.subleaf)
	var regValue uint32
	switch fs.register {
	case 0:
		regValue = a
	case 1:
		regValue = b
	case 2:
		regValue = c
	case 3:
		regValue = d
	}

	supported := []string{}
	for bit, f := range fs.features {
		if (regValue>>bit)&1 == 1 {
			supported = append(supported, f.name)
		}
	}
	return supported
}

// IsFeatureSupported reports if a feature is supported
func IsFeatureSupported(featureName string) bool {
	for _, fs := range cpuFeaturesList {
		// Check condition if present
		if fs.condition != nil && !fs.condition(0) {
			continue
		}

		var bitPos *int
		for bit, f := range fs.features {
			if f.name == featureName {
				bitPos = &bit
				break
			}
		}

		if bitPos == nil {
			continue // feature not in this category
		}

		a, b, c, d := cpuid(fs.leaf, fs.subleaf)
		var regValue uint32
		switch fs.register {
		case 0:
			regValue = a
		case 1:
			regValue = b
		case 2:
			regValue = c
		case 3:
			regValue = d
		}

		if (regValue>>(*bitPos))&1 == 1 {
			return true
		} else {
			return false
		}
	}
	return false
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

// Helper function to process Intel descriptors and add them to TLBInfo
func processIntelDescriptors(info *TLBInfo, bytes ...uint32) {
	for _, val := range bytes {
		if val == 0 {
			continue
		}

		for i := 0; i < 4; i++ {
			descriptor := (val >> (i * 8)) & 0xFF
			if entry := parseIntelDescriptor(descriptor); entry != nil {
				// Add entry to appropriate level and type based on descriptor
				// This is a simplified version - you might want to add more complex parsing
				if strings.Contains(entry.PageSize, "4KB") || strings.Contains(entry.PageSize, "4MB") {
					info.L1.Data = append(info.L1.Data, *entry)
				}
			}
		}
	}
}

// Helper function to parse Intel descriptor into TLBEntry
func parseIntelDescriptor(descriptor uint32) *TLBEntry {
	// This is a simplified version - you would want to expand this map
	descriptors := map[uint32]TLBEntry{
		0x01: {PageSize: "4KB", Entries: 32, Associativity: "4-way"},
		0x02: {PageSize: "4MB", Entries: 2, Associativity: "4-way"},
		0x03: {PageSize: "4KB", Entries: 64, Associativity: "4-way"},
		0x04: {PageSize: "4MB", Entries: 8, Associativity: "4-way"},
		// Add more descriptors as needed
	}

	if entry, ok := descriptors[descriptor]; ok {
		return &entry
	}
	return nil
}

// GetMaxFunctions returns the maximum standard and extended function values supported by the CPU.
func GetMaxFunctions() (uint32, uint32) {
	a, _, _, _ := cpuid(0, 0)
	maxFunc := a

	a, _, _, _ = cpuid(0x80000000, 0)
	maxExtFunc := a

	return maxFunc, maxExtFunc
}

func int32ToBytes(i uint32) []byte {
	return []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
}

// GetVendorID returns the vendor ID of the CPU.
func GetVendorID() string {
	_, b, c, d := cpuid(0, 0)
	return fmt.Sprintf("%s%s%s",
		string([]byte{byte(b), byte(b >> 8), byte(b >> 16), byte(b >> 24)}),
		string([]byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}),
		string([]byte{byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)}),
	)
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

// Helpers

// getAssociativity converts AMD's associativity value to a string description
func getAssociativity(value uint32) string {
	switch value {
	case 0:
		return "Reserved"
	case 1:
		return "1-way (direct mapped)"
	case 2:
		return "2-way"
	case 4:
		return "4-way"
	case 6:
		return "6-way"
	case 8:
		return "8-way"
	case 0xF:
		return "Fully associative"
	default:
		return fmt.Sprintf("%d-way", value)
	}
}

// getIntelAssociativity converts Intel's associativity value to a string description
func getIntelAssociativity(value uint32) string {
	switch value {
	case 0:
		return "Reserved"
	case 1:
		return "Direct mapped"
	case 2:
		return "2-way"
	case 3:
		return "3-way"
	case 4:
		return "4-way"
	case 5:
		return "6-way"
	case 6:
		return "8-way"
	case 7:
		return "12-way"
	case 8:
		return "16-way"
	case 9:
		return "32-way"
	case 10:
		return "48-way"
	case 11:
		return "64-way"
	case 12:
		return "96-way"
	case 13:
		return "128-way"
	case 14:
		return "Fully associative"
	case 15:
		return "Reserved"
	default:
		return fmt.Sprintf("Unknown (%d)", value)
	}
}
