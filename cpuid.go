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

// GetTLBInfo returns TLB information for the CPU
func GetTLBInfo(maxFunc, maxExtFunc uint32, vendorID string) TLBInfo {
	if isAMD() {
		return GetAMDTLBInfo(maxExtFunc)
	}

	if isIntel() {
		return GetIntelTLBInfo(maxFunc)
	}

	fmt.Println("Unknown CPU vendor")
	return TLBInfo{}
}

// GetCacheInfo returns cache information for the CPU
func GetCacheInfo(maxFunc, maxExtFunc uint32, vendorID string) []CPUCacheInfo {
	isIntel := strings.Contains(strings.ToUpper(vendorID), "INTEL")
	isAMD := strings.Contains(strings.ToUpper(vendorID), "AMD")

	if isAMD {
		return GetAMDCache(maxExtFunc)
	}

	if isIntel {
		return GetIntelCache(maxFunc)
	}

	fmt.Println("Unknown CPU vendor")
	return nil
}

// GetAMDCache returns cache information for AMD processors
func GetAMDCache(maxExtFunc uint32) []CPUCacheInfo {
	if maxExtFunc < 0x8000001D {
		return nil
	}

	var caches []CPUCacheInfo
	for i := uint32(0); ; i++ {
		info := GetCPUCacheDetails(0x8000001D, i)
		if info.Type == getCacheTypeString(0) {
			break
		}
		caches = append(caches, info)
	}
	return caches
}

// GetIntelCache returns cache information for Intel processors
func GetIntelCache(maxFunc uint32) []CPUCacheInfo {
	if maxFunc < 4 {
		return nil
	}

	var caches []CPUCacheInfo
	for i := uint32(0); ; i++ {
		info := GetCPUCacheDetails(4, i)
		if info.Type == getCacheTypeString(0) {
			break
		}
		caches = append(caches, info)
	}
	return caches
}

// GetCPUCacheDetails returns detailed information about the CPU cache.
func GetCPUCacheDetails(leaf, subLeaf uint32) CPUCacheInfo {
	a, b, c, _ := cpuid(leaf, subLeaf)
	cacheType := a & 0x1F
	level := (a >> 5) & 0x7
	lineSize := (b & 0xFFF) + 1
	partitions := ((b >> 12) & 0x3FF) + 1
	associativity := ((b >> 22) & 0x3FF) + 1
	sets := c + 1
	size := lineSize * partitions * associativity * sets
	selfInit := (a>>8)&1 != 0
	fullyAssoc := (a>>9)&1 != 0
	maxProcIDs := ((a >> 26) & 0x3F) + 1
	typeString := getCacheTypeString(cacheType)
	maxCoresSharing := ((a >> 14) & 0xFFF) + 1

	writePolicy := ""
	switch (a >> 10) & 0x3 {
	case 0:
		writePolicy = "Write Back"
	case 1:
		writePolicy = "Write Through"
	case 2:
		writePolicy = "Write Protected"
	default:
		writePolicy = "Unknown"
	}

	return CPUCacheInfo{
		Level:            level,
		Type:             typeString,
		SizeKB:           size / 1024,
		Ways:             associativity,
		LineSizeBytes:    lineSize,
		TotalSets:        sets,
		MaxCoresSharing:  maxCoresSharing,
		SelfInitializing: selfInit,
		FullyAssociative: fullyAssoc,
		MaxProcessorIDs:  maxProcIDs,
		WritePolicy:      writePolicy,
	}
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

// IntelHypridInfo Stores information about hybrid CPUs for Intel processors
type IntelHypridInfo struct {
	HybridCPU     bool
	NativeModelID uint32
	CoreType      uint32
}

// GetIntelHybrid returns information about hybrid CPUs for Intel processors
func GetIntelHybrid() IntelHypridInfo {
	a, _, _, _ := cpuid(0x1A, 0)
	if a&1 != 0 {
		return IntelHypridInfo{
			HybridCPU:     true,
			NativeModelID: (a >> 24) & 0xFF,
			CoreType:      (a >> 16) & 0xFF,
		}
	}
	return IntelHypridInfo{
		HybridCPU: false,
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

// GetAMDTLBInfo retrieves TLB information for AMD processors
func GetAMDTLBInfo(maxExtFunc uint32) TLBInfo {
	info := TLBInfo{
		Vendor: "AMD",
	}

	// L1 TLB info from 0x80000005
	a, b, _, _ := cpuid(0x80000005, 0)

	// L1 Data TLB
	info.L1.Data = append(info.L1.Data, TLBEntry{
		PageSize:      "2MB/4MB",
		Entries:       int((a >> 16) & 0xFF),
		Associativity: getAssociativity((a >> 8) & 0xFF),
	})
	info.L1.Data = append(info.L1.Data, TLBEntry{
		PageSize:      "4KB",
		Entries:       int((a >> 24) & 0xFF),
		Associativity: getAssociativity((a >> 8) & 0xFF),
	})

	// L1 Instruction TLB
	info.L1.Instruction = append(info.L1.Instruction, TLBEntry{
		PageSize:      "2MB/4MB",
		Entries:       int((b >> 16) & 0xFF),
		Associativity: getAssociativity((b >> 8) & 0xFF),
	})
	info.L1.Instruction = append(info.L1.Instruction, TLBEntry{
		PageSize:      "4KB",
		Entries:       int((b >> 24) & 0xFF),
		Associativity: getAssociativity((b >> 8) & 0xFF),
	})

	// L2 TLB info from 0x80000006 if available
	if maxExtFunc >= 0x80000006 {
		a, b, _, _ = cpuid(0x80000006, 0)

		// L2 Data TLB
		info.L2.Data = append(info.L2.Data, TLBEntry{
			PageSize:      "2MB/4MB",
			Entries:       int((a >> 16) & 0xFFF),
			Associativity: getAssociativity((a >> 12) & 0xF),
		})
		info.L2.Data = append(info.L2.Data, TLBEntry{
			PageSize:      "4KB",
			Entries:       int((a >> 28) & 0xF),
			Associativity: getAssociativity((a >> 12) & 0xF),
		})

		// L2 Instruction TLB
		info.L2.Instruction = append(info.L2.Instruction, TLBEntry{
			PageSize:      "2MB/4MB",
			Entries:       int((b >> 16) & 0xFFF),
			Associativity: getAssociativity((b >> 12) & 0xF),
		})
		info.L2.Instruction = append(info.L2.Instruction, TLBEntry{
			PageSize:      "4KB",
			Entries:       int((b >> 28) & 0xF),
			Associativity: getAssociativity((b >> 12) & 0xF),
		})

		// L3 TLB info if supported
		if maxExtFunc >= 0x80000019 {
			a, _, _, _ = cpuid(0x80000019, 0)

			info.L3.Data = append(info.L3.Data, TLBEntry{
				PageSize:      "1GB",
				Entries:       int((a >> 16) & 0xFFF),
				Associativity: getAssociativity((a >> 12) & 0xF),
			})
		}
	}

	return info
}

// GetIntelTLBInfo retrieves TLB information for Intel processors
func GetIntelTLBInfo(maxFunc uint32) TLBInfo {
	info := TLBInfo{
		Vendor: "Intel",
	}

	if maxFunc < 0x2 {
		return info
	}

	// Process traditional descriptors (leaf 0x2)
	a, b, c, d := cpuid(0x2, 0)
	processIntelDescriptors(&info, a>>8, b, c, d)

	// Process structured TLB information (leaf 0x18)
	if maxFunc >= 0x18 {
		subleaf := uint32(0)
		for {
			_, b, c, d = cpuid(0x18, subleaf)

			if (d & 0x1F) != 1 { // 1 indicates TLB entry
				break
			}

			entry := TLBEntry{
				PageSize:      getTLBPageSize(b),
				Entries:       int((b>>16)&0xFFF) + 1,
				Associativity: getIntelAssociativity(b >> 8),
			}

			level := (c >> 5) & 0x7
			tlbType := getTLBType((c >> 8) & 0x3)

			// Add entry to appropriate level and type
			switch level {
			case 1:
				addIntelTLBEntry(&info.L1, tlbType, entry)
			case 2:
				addIntelTLBEntry(&info.L2, tlbType, entry)
			case 3:
				addIntelTLBEntry(&info.L3, tlbType, entry)
			}

			subleaf++
		}
	}

	return info
}

// PrintTLBInfo prints the TLB information in a formatted way
func PrintTLBInfo(info TLBInfo) {
	// fmt.Printf("%s TLB Information:\n\n", info.Vendor)

	// Helper function to print TLB entries
	printEntries := func(label string, entries []TLBEntry) {
		if len(entries) > 0 {
			fmt.Printf("%s:\n", label)
			for _, entry := range entries {
				fmt.Printf("    %s Pages: %d entries, %s\n",
					entry.PageSize,
					entry.Entries,
					entry.Associativity)
			}
		}
	}

	// Print L1 TLB
	if len(info.L1.Data) > 0 || len(info.L1.Instruction) > 0 || len(info.L1.Unified) > 0 {
		fmt.Println("L1 TLB:")
		printEntries("  Data", info.L1.Data)
		printEntries("  Instruction", info.L1.Instruction)
		printEntries("  Unified", info.L1.Unified)
	}

	// Print L2 TLB
	if len(info.L2.Data) > 0 || len(info.L2.Instruction) > 0 || len(info.L2.Unified) > 0 {
		fmt.Println("\nL2 TLB:")
		printEntries("  Data", info.L2.Data)
		printEntries("  Instruction", info.L2.Instruction)
		printEntries("  Unified", info.L2.Unified)
	}

	// Print L3 TLB
	if len(info.L3.Data) > 0 || len(info.L3.Instruction) > 0 || len(info.L3.Unified) > 0 {
		fmt.Println("\nL3 TLB:")
		printEntries("  Data", info.L3.Data)
		printEntries("  Instruction", info.L3.Instruction)
		printEntries("  Unified", info.L3.Unified)
	}
}

// Helper function to add Intel TLB entry to appropriate slice
func addIntelTLBEntry(level *TLBLevel, tlbType string, entry TLBEntry) {
	switch tlbType {
	case "Data":
		level.Data = append(level.Data, entry)
	case "Instruction":
		level.Instruction = append(level.Instruction, entry)
	case "Unified":
		level.Unified = append(level.Unified, entry)
	}
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

// PrintProcessorInfo prints detailed information about the CPU.
func PrintProcessorInfo() {
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

// Helpers

// PrintCacheTable prints the cache information in a table format
func PrintCacheTable(caches []CPUCacheInfo) {
	maxKeyLength := 0
	keys := []string{
		"L%d %s Cache:",
		"Ways:",
		"Line Size:",
		"Total Sets:",
		"Max Cores Sharing:",
		"Self Initializing:",
		"Fully Associative:",
		"Max Processor IDs:",
		"Write Policy:",
	}
	for _, key := range keys {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
	}

	for _, cache := range caches {
		fmt.Printf("  %-*s %d KB\n", maxKeyLength, fmt.Sprintf("L%d %s Cache:", cache.Level, cache.Type), cache.SizeKB)
		fmt.Printf("  %-*s %d\n", maxKeyLength, "Ways:", cache.Ways)
		fmt.Printf("  %-*s %d bytes\n", maxKeyLength, "Line Size:", cache.LineSizeBytes)
		fmt.Printf("  %-*s %d\n", maxKeyLength, "Total Sets:", cache.TotalSets)
		fmt.Printf("  %-*s %d\n", maxKeyLength, "Max Cores Sharing:", cache.MaxCoresSharing)
		fmt.Printf("  %-*s %v\n", maxKeyLength, "Self Initializing:", cache.SelfInitializing)
		fmt.Printf("  %-*s %v\n", maxKeyLength, "Fully Associative:", cache.FullyAssociative)
		fmt.Printf("  %-*s %d\n", maxKeyLength, "Max Processor IDs:", cache.MaxProcessorIDs)
		fmt.Printf("  %-*s %s\n", maxKeyLength, "Write Policy:", cache.WritePolicy)
		fmt.Println()
	}
}

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

// getTLBPageSize converts Intel's page size value to a string description
func getTLBPageSize(value uint32) string {
	switch value & 0xF {
	case 1:
		return "4KB"
	case 2:
		return "2MB"
	case 3:
		return "4MB"
	case 4:
		return "1GB"
	case 5:
		return "256MB"
	case 0xF:
		return "Reserved"
	default:
		return "Unknown"
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

// getTLBType converts Intel's TLB type value to a string description
func getTLBType(value uint32) string {
	switch value {
	case 0:
		return "Invalid"
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
