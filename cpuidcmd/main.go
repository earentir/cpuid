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
	checkFeature()
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

	// cpuid.PrintProcessorInfo()
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

func checkFeature() {
	featureName := "SSE4.2"
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
	caches := cpuid.GetCacheInfo(maxFunc, maxExtFunc, vendorID)
	cpuid.PrintCacheTable(caches)
}

func printTLBInfo() {
	tlbs := cpuid.GetTLBInfo(maxFunc, maxExtFunc, vendorID)
	cpuid.PrintTLBInfo(tlbs)
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
