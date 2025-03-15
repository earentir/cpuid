// Package cpuid provides information about the CPU running the current program.
package cpuid

import "fmt"

// GetTLBInfo returns TLB information for the CPU
func GetTLBInfo(maxFunc, maxExtFunc uint32) (TLBInfo, error) {
	if isAMD() {
		return GetAMDTLBInfo(maxExtFunc), nil
	}

	if isIntel() {
		return GetIntelTLBInfo(maxFunc), nil
	}

	return TLBInfo{}, fmt.Errorf("Unknown/Unsupported CPU vendor")
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
