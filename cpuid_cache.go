// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"fmt"
	"strings"
)

// GetCacheInfo returns cache information for the CPU
func GetCacheInfo(maxFunc, maxExtFunc uint32, vendorID string, offline bool, filename string) ([]CPUCacheInfo, error) {
	isIntel := strings.Contains(strings.ToUpper(vendorID), "INTEL")
	isAMD := strings.Contains(strings.ToUpper(vendorID), "AMD")

	if isAMD {
		return GetAMDCache(maxExtFunc, offline, filename), nil
	}

	if isIntel {
		return GetIntelCache(maxFunc, offline, filename), nil
	}

	return []CPUCacheInfo{}, fmt.Errorf("Unknown/Unsupported CPU vendor")
}

// GetAMDCache returns cache information for AMD processors
func GetAMDCache(maxExtFunc uint32, offline bool, filename string) []CPUCacheInfo {
	if maxExtFunc < 0x8000001D {
		return nil
	}

	var caches []CPUCacheInfo
	for i := uint32(0); ; i++ {
		info := GetCPUCacheDetails(0x8000001D, i, offline, filename)
		if info.Type == getCacheTypeString(0) {
			break
		}
		caches = append(caches, info)
	}
	return caches
}

// GetIntelCache returns cache information for Intel processors
func GetIntelCache(maxFunc uint32, offline bool, filename string) []CPUCacheInfo {
	if maxFunc < 4 {
		return nil
	}

	var caches []CPUCacheInfo
	for i := uint32(0); ; i++ {
		info := GetCPUCacheDetails(4, i, offline, filename)
		if info.Type == getCacheTypeString(0) {
			break
		}
		caches = append(caches, info)
	}
	return caches
}

// GetCPUCacheDetails returns detailed information about the CPU cache.
func GetCPUCacheDetails(leaf, subLeaf uint32, offline bool, filename string) CPUCacheInfo {
	a, b, c, _ := CPUIDWithMode(leaf, subLeaf, offline, filename)
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
