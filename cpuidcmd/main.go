// cmd/cpuinfo/main.go
package main

import (
	"fmt"
	"strings"

	"github.com/earentir/cpuid"
)

func main() {
	// Get basic CPU information
	info := cpuid.GetCPUInfo()

	fmt.Printf("CPU Information:\n")
	fmt.Printf("Vendor: %s\n", info.VendorID)
	fmt.Printf("Brand String: %s\n", info.BrandString)
	fmt.Printf("Family: %d\n", info.Family)
	fmt.Printf("Model: %d\n", info.Model)
	fmt.Printf("Stepping: %d\n", info.Stepping)
	fmt.Printf("Max Standard Function: %d\n", info.MaxStandard)
	fmt.Printf("Max Extended Function: 0x%08x\n\n", info.MaxExtended)

	// Get cache information
	fmt.Printf("Cache Information:\n")
	for _, cache := range cpuid.GetCacheInfo() {
		fmt.Printf("L%d %s Cache:\n", cache.Level, cache.Type)
		fmt.Printf("  Size: %d KB\n", cache.Size/1024)
		fmt.Printf("  Ways: %d\n", cache.Ways)
		fmt.Printf("  Line Size: %d bytes\n", cache.LineSize)
		fmt.Printf("  Sets: %d\n", cache.Sets)
		fmt.Printf("  Cores Sharing: %d\n\n", cache.SharedCores)
	}

	// Get CPU features
	fmt.Printf("CPU Features:\n")
	for setName, features := range cpuid.GetFeatures() {
		if len(features) > 0 {
			fmt.Printf("\n%s:\n", setName)
			fmt.Printf("  %s\n", strings.Join(features, ", "))
		}
	}
}
