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
	cpuid.PrintBasicInfo()
	fmt.Println()
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("Detecting CPU features for x86/x64")
	fmt.Println("==================================")
	printBasicInfo()
	// fmt.Println()
	// fmt.Println("Cache Info")
	// fmt.Println("==========")
	// printCacheInfo()
	// fmt.Println()
	// fmt.Println("Translation Lookaside Buffer Info")
	// fmt.Println("=================================")
	// printTLBInfo()
	fmt.Println()
	fmt.Println("Intel Hybric Core Info")
	fmt.Println("======================")
	printIntelHybridInfo()

	// cpuid.PrintProcessorInfo()
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
	fmt.Printf("  Hybrid Core: %t\n", hybridInfo.HybridCPU)
	fmt.Printf("  Native Model ID: %d\n", hybridInfo.NativeModelID)
	fmt.Printf("  Hybrid Model ID: %d\n", hybridInfo.CoreType)
}
