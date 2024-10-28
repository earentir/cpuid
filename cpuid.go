// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"fmt"
	"sort"
	"strings"
)

func cpuid(eax, ecx uint32) (a, b, c, d uint32)

// GetCPUInfo returns the basic CPU information
func GetCPUInfo() *CPUInfo {
	info := &CPUInfo{
		Features: make(map[string][]string),
	}

	info.MaxStandard, info.MaxExtended = getMaxFunctions()
	info.VendorID = getVendorID()

	// Get processor info
	a, _, _, _ := cpuid(1, 0)
	info.Stepping = a & 0xF
	modelID := (a >> 4) & 0xF
	familyID := (a >> 8) & 0xF
	info.ProcessorType = (a >> 12) & 0x3
	extendedModelID := (a >> 16) & 0xF
	extendedFamilyID := (a >> 20) & 0xFF

	// Calculate effective values
	info.Model = modelID
	if familyID == 0xF || familyID == 0x6 {
		info.Model += extendedModelID << 4
	}

	info.Family = familyID
	if familyID == 0xF {
		info.Family += extendedFamilyID
	}

	// Get brand string
	if info.MaxExtended >= 0x80000004 {
		info.BrandString = getBrandString()
	}

	return info
}

func getVendorID() string {
	_, b, c, d := cpuid(0, 0)
	return fmt.Sprintf("%s%s%s",
		string([]byte{byte(b), byte(b >> 8), byte(b >> 16), byte(b >> 24)}),
		string([]byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}),
		string([]byte{byte(c), byte(c >> 8), byte(c >> 16), byte(c >> 24)}),
	)
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

// GetFeatures returns all CPU features grouped by feature set
func GetFeatures() map[string][]string {
	features := make(map[string][]string)

	for name, set := range cpuFeaturesList {
		if set.condition != nil && !set.condition(0) {
			continue
		}

		a, b, c, d := cpuid(set.leaf, set.subleaf)
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

		features[name] = getFeatureFlags(set.features, regValue)
	}

	return features
}

// GetCacheInfo returns detailed cache information
func GetCacheInfo() []CacheInfo {
	var caches []CacheInfo

	if isAMD && maxExtFunc >= 0x8000001D {
		caches = getAMDCacheInfo()
	} else if maxFunc >= 4 {
		caches = getIntelCacheInfo()
	}

	return caches
}

// Helper functions
func getFeatureFlags(features map[int]Feature, reg uint32) []string {
	var recognized []string

	for i := 0; i < 32; i++ {
		if (reg>>i)&1 == 1 {
			if feature, exists := features[i]; exists {
				recognized = append(recognized, feature.name)
			}
		}
	}

	sort.Strings(recognized)
	return recognized
}

func getBrandString() string {
	var brand [48]byte
	for i := 0; i < 3; i++ {
		a, b, c, d := cpuid(0x80000002+uint32(i), 0)
		copy(brand[i*16:], int32ToBytes(a))
		copy(brand[i*16+4:], int32ToBytes(b))
		copy(brand[i*16+8:], int32ToBytes(c))
		copy(brand[i*16+12:], int32ToBytes(d))
	}
	return strings.TrimSpace(string(brand[:]))
}

func getAMDCacheInfo() []CacheInfo {
	var caches []CacheInfo
	for i := uint32(0); ; i++ {
		a, b, c, _ := cpuid(0x8000001D, i)
		cacheType := a & 0x1F
		if cacheType == 0 {
			break
		}

		cache := CacheInfo{
			Level:       (a >> 5) & 0x7,
			Type:        getCacheTypeString(cacheType),
			LineSize:    (b & 0xFFF) + 1,
			Ways:        ((b >> 22) & 0x3FF) + 1,
			Sets:        c + 1,
			SharedCores: ((a >> 14) & 0xFFF) + 1,
		}

		cache.Size = cache.LineSize * ((b>>12)&0x3FF + 1) * cache.Ways * cache.Sets
		caches = append(caches, cache)
	}
	return caches
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

func getIntelCacheInfo() []CacheInfo {
	var caches []CacheInfo
	for i := uint32(0); ; i++ {
		a, b, c, _ := cpuid(4, i)
		cacheType := a & 0x1F
		if cacheType == 0 {
			break
		}

		cache := CacheInfo{
			Level:       (a >> 5) & 0x7,
			Type:        getCacheTypeString(cacheType),
			LineSize:    (b & 0xFFF) + 1,
			Ways:        ((b >> 22) & 0x3FF) + 1,
			Sets:        c + 1,
			SharedCores: ((a >> 14) & 0xFFF) + 1,
		}

		cache.Size = cache.LineSize * ((b>>12)&0x3FF + 1) * cache.Ways * cache.Sets
		caches = append(caches, cache)
	}
	return caches
}
