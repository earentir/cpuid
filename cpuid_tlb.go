// Package cpuid provides information about the CPU running the current program.
package cpuid

import (
	"fmt"
	"strings"
)

// GetTLBInfo returns TLB information for the CPU
func GetTLBInfo(maxFunc, maxExtFunc uint32, offline bool, filename string) (TLBInfo, error) {
	if isAMD(offline, filename) {
		return GetAMDTLBInfo(maxExtFunc, offline, filename), nil
	}

	if isIntel(offline, filename) {
		return GetIntelTLBInfo(maxFunc, offline, filename), nil
	}

	return TLBInfo{}, fmt.Errorf("Unknown/Unsupported CPU vendor")
}

// GetAMDTLBInfo retrieves TLB information for AMD processors
func GetAMDTLBInfo(maxExtFunc uint32, offline bool, filename string) TLBInfo {
	info := TLBInfo{
		Vendor: "AMD",
	}

	// L1 TLB info from 0x80000005
	a, b, _, _ := CPUIDWithMode(0x80000005, 0, offline, filename)

	// L1 Data TLB
	info.L1.Data = append(info.L1.Data, TLBEntry{
		PageSize:      "2MB/4MB",
		Entries:       int((a >> 16) & 0xFF),
		Associativity: getAMDAssociativity((a >> 8) & 0xFF),
	})
	info.L1.Data = append(info.L1.Data, TLBEntry{
		PageSize:      "4KB",
		Entries:       int((a >> 24) & 0xFF),
		Associativity: getAMDAssociativity((a >> 8) & 0xFF),
	})

	// L1 Instruction TLB
	info.L1.Instruction = append(info.L1.Instruction, TLBEntry{
		PageSize:      "2MB/4MB",
		Entries:       int((b >> 16) & 0xFF),
		Associativity: getAMDAssociativity((b >> 8) & 0xFF),
	})
	info.L1.Instruction = append(info.L1.Instruction, TLBEntry{
		PageSize:      "4KB",
		Entries:       int((b >> 24) & 0xFF),
		Associativity: getAMDAssociativity((b >> 8) & 0xFF),
	})

	// L2 TLB info from 0x80000006 if available
	if maxExtFunc >= 0x80000006 {
		a, b, _, _ = CPUIDWithMode(0x80000006, 0, offline, filename)

		// L2 Data TLB
		info.L2.Data = append(info.L2.Data, TLBEntry{
			PageSize:      "2MB/4MB",
			Entries:       int((a >> 16) & 0xFFF),
			Associativity: getAMDAssociativity((a >> 12) & 0xF),
		})
		info.L2.Data = append(info.L2.Data, TLBEntry{
			PageSize:      "4KB",
			Entries:       int((a >> 28) & 0xF),
			Associativity: getAMDAssociativity((a >> 12) & 0xF),
		})

		// L2 Instruction TLB
		info.L2.Instruction = append(info.L2.Instruction, TLBEntry{
			PageSize:      "2MB/4MB",
			Entries:       int((b >> 16) & 0xFFF),
			Associativity: getAMDAssociativity((b >> 12) & 0xF),
		})
		info.L2.Instruction = append(info.L2.Instruction, TLBEntry{
			PageSize:      "4KB",
			Entries:       int((b >> 28) & 0xF),
			Associativity: getAMDAssociativity((b >> 12) & 0xF),
		})

		// L3 TLB info if supported
		if maxExtFunc >= 0x80000019 {
			a, _, _, _ = CPUIDWithMode(0x80000019, 0, offline, filename)

			info.L3.Data = append(info.L3.Data, TLBEntry{
				PageSize:      "1GB",
				Entries:       int((a >> 16) & 0xFFF),
				Associativity: getAMDAssociativity((a >> 12) & 0xF),
			})
		}
	}

	return info
}

// GetIntelTLBInfo retrieves TLB information for Intel processors
func GetIntelTLBInfo(maxFunc uint32, offline bool, filename string) TLBInfo {
	info := TLBInfo{
		Vendor: "Intel",
	}

	if maxFunc < 0x2 {
		return info
	}

	// Process traditional descriptors (leaf 0x2)
	a, b, c, d := CPUIDWithMode(0x2, 0, offline, filename)
	processIntelDescriptors(&info, a>>8, b, c, d)

	// Process structured TLB information (leaf 0x18)
	if maxFunc >= 0x18 {
		subleaf := uint32(0)
		for {
			_, b, c, d = CPUIDWithMode(0x18, subleaf, offline, filename)

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

// getAMDAssociativity converts AMD's associativity value to a string description
func getAMDAssociativity(value uint32) string {
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
