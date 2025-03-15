// main.go
package main

import (
	"fmt"

	"github.com/earentir/cpuid"
)

var (
	maxFunc    uint32
	maxExtFunc uint32
	vendorID   string
)

func init() {
	maxFunc, maxExtFunc = cpuid.GetMaxFunctions()
	vendorID = cpuid.GetVendorID()
}

func main() {
	fmt.Println("Detecting CPU features for x86/x64")
	fmt.Println("==================================")
	fmt.Println("Check if SSE4.2 is supported on this CPU")
	checkFeature("SSE4.2")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("All Available CPU Features")
	fmt.Println("==========================")
	getAllFeatureCategories()
	fmt.Println()
	fmt.Println("All Available CPU Features with Details")
	fmt.Println("========================================")
	getAllFeatureCategoriesWithDetails()
	fmt.Println()
	fmt.Println("All Known Features in StandardECX")
	fmt.Println("==================================")
	getAllKnownFeaturesCategory("StandardECX")
	fmt.Println()
	fmt.Println("All Supported Features in StandardECX")
	fmt.Println("=====================================")
	getAllSupportedFeaturesCategory("StandardECX")
	fmt.Println()
	fmt.Println("Basic Info")
	fmt.Println("==========")
	printBasicInfo()
	fmt.Println()
	fmt.Println("Cache Info")
	fmt.Println("==========")
	printCacheInfo()
	fmt.Println()
	fmt.Println("Translation Lookaside Buffer Info")
	fmt.Println("=================================")
	printTLBInfo()
	fmt.Println()
	fmt.Println("Intel Hybric Core Info")
	fmt.Println("======================")
	printIntelHybridInfo()
}

func getAllSupportedFeaturesCategory(category string) {
	supportedFeatures := cpuid.GetSupportedFeatures(category)
	fmt.Println("\nSupported Features in StandardECX:")
	for _, f := range supportedFeatures {
		fmt.Println(" -", f)
	}
}

func getAllKnownFeaturesCategory(category string) {
	knownFeatures := cpuid.GetAllKnownFeatures(category)
	fmt.Println("\nKnown Features in StandardECX:")
	for _, f := range knownFeatures {
		fmt.Println(" -", f)
	}
}

func getAllFeatureCategoriesWithDetails() {
	detailedCategories := cpuid.GetAllFeatureCategoriesDetailed()
	for catName, features := range detailedCategories {
		fmt.Println("Category:", catName)
		for _, f := range features {
			line := fmt.Sprintf("  - %s: %s (Vendor: %s)", f["name"], f["description"], f["vendor"])
			if eq, ok := f["equivalent"]; ok {
				line += fmt.Sprintf(" [Equivalent: %s]", eq)
			}
			fmt.Println(line)
		}
	}
}

func getAllFeatureCategories() {
	categories := cpuid.GetAllFeatureCategories()
	fmt.Println("All Feature Categories:")
	for _, cat := range categories {
		fmt.Println(" -", cat)
	}
}

func checkFeature(featureName string) {
	if cpuid.IsFeatureSupported(featureName) {
		fmt.Printf("\n%s is supported on this CPU.\n", featureName)
	} else {
		fmt.Printf("\n%s is NOT supported on this CPU.\n", featureName)
	}
}

func printBasicInfo() {
	processorInfo := cpuid.GetProcessorInfo(maxFunc, maxExtFunc)
	fmt.Printf("  CPUID Max Standard Function: %d\n", processorInfo.MaxFunc)
	fmt.Printf("  CPUID Max Extended Function: 0x%08x\n", processorInfo.MaxExtFunc)
	fmt.Printf("  CPU Vendor ID:               %s\n", processorInfo.VendorID)
	fmt.Println()
	fmt.Println("Processor Details")
	fmt.Println("=================")
	fmt.Printf("  Brand String:   %s\n", processorInfo.BrandString)
	fmt.Printf("  Family:         %d (0x%x)\n", processorInfo.EffectiveFamily, processorInfo.EffectiveFamily)
	fmt.Printf("  Model:          %d (0x%x)\n", processorInfo.EffectiveModel, processorInfo.EffectiveModel)
	fmt.Printf("  Stepping ID:    %s\n", processorInfo.SteppingID)
	fmt.Printf("  Processor Type: %s\n", processorInfo.ProcessorType)
	fmt.Println()
	fmt.Printf("  Max Logical Processors: %d\n", processorInfo.MaxLogicalProcessors)
	fmt.Printf("  Initial APIC ID: %d\n", processorInfo.InitialAPICID)
	fmt.Printf("  Physical Address Bits: %d\n", processorInfo.PhysicalAddressBits)
	fmt.Printf("  Linear Address Bits: %d\n", processorInfo.LinearAddressBits)
	fmt.Printf("  Cores: %d\n", processorInfo.CoreCount)
	fmt.Printf("  Threads Per Core: %d\n", processorInfo.ThreadPerCore)
}

func printCacheInfo() {
	//Fetch the cache information
	caches, err := cpuid.GetCacheInfo(maxFunc, maxExtFunc, vendorID)
	if err != nil {
		fmt.Println("Failed to fetch cache information:", err)
		return
	}

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

func printTLBInfo() {
	tlbs, err := cpuid.GetTLBInfo(maxFunc, maxExtFunc)
	if err != nil {
		fmt.Println("Failed to fetch TLB information:", err)
		return
	}

	// Helper function to print TLB entries
	printEntries := func(label string, entries []cpuid.TLBEntry) {
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
	if len(tlbs.L1.Data) > 0 || len(tlbs.L1.Instruction) > 0 || len(tlbs.L1.Unified) > 0 {
		fmt.Println("L1 TLB:")
		printEntries("  Data", tlbs.L1.Data)
		printEntries("  Instruction", tlbs.L1.Instruction)
		printEntries("  Unified", tlbs.L1.Unified)
	}

	// Print L2 TLB
	if len(tlbs.L2.Data) > 0 || len(tlbs.L2.Instruction) > 0 || len(tlbs.L2.Unified) > 0 {
		fmt.Println("\nL2 TLB:")
		printEntries("  Data", tlbs.L2.Data)
		printEntries("  Instruction", tlbs.L2.Instruction)
		printEntries("  Unified", tlbs.L2.Unified)
	}

	// Print L3 TLB
	if len(tlbs.L3.Data) > 0 || len(tlbs.L3.Instruction) > 0 || len(tlbs.L3.Unified) > 0 {
		fmt.Println("\nL3 TLB:")
		printEntries("  Data", tlbs.L3.Data)
		printEntries("  Instruction", tlbs.L3.Instruction)
		printEntries("  Unified", tlbs.L3.Unified)
	}
}

func printIntelHybridInfo() {
	hybridInfo := cpuid.GetIntelHybrid()
	fmt.Printf("  Hybrid CPU: %t\n", hybridInfo.HybridCPU)
	if hybridInfo.HybridCPU {
		fmt.Printf("  Native Model ID: %d\n", hybridInfo.NativeModelID)
		fmt.Printf("  Core Type ID: %d\n", hybridInfo.CoreType)
		fmt.Printf("  Core Type: %s\n", hybridInfo.CoreTypeName)
	}
}
