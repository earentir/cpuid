package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

// FeatureSet defines a group of CPU features and how to query them
type FeatureSet struct {
	name      string            // Display name
	leaf      uint32            // CPUID leaf (eax input)
	subleaf   uint32            // CPUID subleaf (ecx input)
	register  int               // Which register to use (0=EAX, 1=EBX, 2=ECX, 3=EDX)
	condition func(uint32) bool // Optional condition function
	group     string            // Group name
	features  map[int]Feature   // Feature map
}

// Feature represents a CPU feature with its description and function
type Feature struct {
	name                  string
	description           string
	function              string
	vendor                string // "amd", "intel", or "common"
	equivalentFeatureName string // equivalent feature name here
	equivalent            int    // equivalent int here
}

var (
	vendorID   = getVendorID()
	isAMD      = strings.Contains(strings.ToUpper(vendorID), "AMD")
	extb       uint32
	maxFunc    uint32
	maxExtFunc uint32
)

// All CPU features are stored here
var cpuFeaturesList = map[string]FeatureSet{
	"StandardECX": {
		name:     "Standard Features ECX",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "Basic CPU",
		features: map[int]Feature{
			0:  {"SSE3", "Streaming SIMD Extensions 3", "CPUID.1:ECX.SSE3[bit 0]", "common", "", -1},
			1:  {"PCLMULQDQ", "Carryless Multiplication", "CPUID.1:ECX.PCLMULQDQ[bit 1]", "common", "", -1},
			2:  {"DTES64", "64-bit DS Area", "CPUID.1:ECX.DTES64[bit 2]", "intel", "", -1},
			3:  {"MONITOR", "MONITOR/MWAIT Instructions", "CPUID.1:ECX.MONITOR[bit 3]", "common", "", -1},
			4:  {"DS-CPL", "CPL Qualified Debug Store", "CPUID.1:ECX.DS-CPL[bit 4]", "intel", "", -1},
			5:  {"VMX", "Virtual Machine Extensions", "CPUID.1:ECX.VMX[bit 5]", "intel", "AMDExtendedECX", 2}, // equivalent to SVM (2)
			6:  {"SMX", "Safer Mode Extensions", "CPUID.1:ECX.SMX[bit 6]", "intel", "", -1},
			7:  {"EIST", "Enhanced Intel SpeedStep Technology", "CPUID.1:ECX.EIST[bit 7]", "intel", "AMDExtendedECX", 1}, // equivalent to AMD Cool'n'Quiet (1)
			8:  {"TM2", "Thermal Monitor 2", "CPUID.1:ECX.TM2[bit 8]", "intel", "", -1},
			9:  {"SSSE3", "Supplemental Streaming SIMD Extensions 3", "CPUID.1:ECX.SSSE3[bit 9]", "common", "", -1},
			10: {"CNXT-ID", "L1 Context ID", "CPUID.1:ECX.CNXT-ID[bit 10]", "intel", "", -1},
			11: {"SDBG", "Silicon Debug Interface", "CPUID.1:ECX.SDBG[bit 11]", "intel", "", -1},
			12: {"FMA", "Fused Multiply Add", "CPUID.1:ECX.FMA[bit 12]", "common", "", -1},
			13: {"CMPXCHG16B", "CMPXCHG16B Instruction", "CPUID.1:ECX.CMPXCHG16B[bit 13]", "common", "", -1},
			14: {"xTPR", "xTPR Update Control", "CPUID.1:ECX.xTPR[bit 14]", "intel", "", -1},
			15: {"PDCM", "Perfmon and Debug Capability", "CPUID.1:ECX.PDCM[bit 15]", "intel", "", -1},
			16: {"PCID", "Process Context Identifiers", "CPUID.1:ECX.PCID[bit 16]", "intel", "", -1},
			17: {"DCA", "Direct Cache Access", "CPUID.1:ECX.DCA[bit 17]", "intel", "", -1},
			18: {"SSE4.1", "Streaming SIMD Extensions 4.1", "CPUID.1:ECX.SSE4.1[bit 18]", "intel", "AMDExtendedECX", 6}, // equivalent to SSE4a (6)
			19: {"SSE4.2", "Streaming SIMD Extensions 4.2", "CPUID.1:ECX.SSE4.2[bit 19]", "common", "", -1},
			20: {"x2APIC", "x2APIC Support", "CPUID.1:ECX.x2APIC[bit 20]", "common", "", -1},
			21: {"MOVBE", "MOVBE Instruction", "CPUID.1:ECX.MOVBE[bit 21]", "common", "", -1},
			22: {"POPCNT", "POPCNT Instruction", "CPUID.1:ECX.POPCNT[bit 22]", "common", "", -1},
			23: {"TSC-DEADLINE", "Local APIC supports TSC Deadline", "CPUID.1:ECX.TSC-DEADLINE[bit 23]", "common", "", -1},
			24: {"AES", "AES Instruction Set", "CPUID.1:ECX.AES[bit 24]", "common", "", -1},
			25: {"XSAVE", "XSAVE/XRSTOR States", "CPUID.1:ECX.XSAVE[bit 25]", "common", "", -1},
			26: {"OSXSAVE", "OS has enabled XSETBV/XGETBV", "CPUID.1:ECX.OSXSAVE[bit 26]", "common", "", -1},
			27: {"AVX", "Advanced Vector Extensions", "CPUID.1:ECX.AVX[bit 27]", "common", "", -1},
			28: {"F16C", "16-bit FP conversion", "CPUID.1:ECX.F16C[bit 28]", "common", "", -1},
			29: {"RDRAND", "RDRAND instruction", "CPUID.1:ECX.RDRAND[bit 29]", "intel", "", -1},
			30: {"HYPERVISOR", "Running on a hypervisor", "CPUID.1:ECX.HYPERVISOR[bit 30]", "common", "", -1},
			// AMD specific features in ECX
			31: {"RAZ", "Return All Zeros", "CPUID.1:ECX.RAZ[bit 31]", "amd", "", -1},
		},
	}, "StandardEDX": {
		name:     "Standard Features EDX",
		leaf:     1,
		subleaf:  0,
		register: 3,
		group:    "Basic CPU",
		features: map[int]Feature{
			0:  {"FPU", "Floating Point Unit", "CPUID.1:EDX.FPU[bit 0]", "common", "", -1},
			1:  {"VME", "Virtual 8086 Mode Extensions", "CPUID.1:EDX.VME[bit 1]", "common", "", -1},
			2:  {"DE", "Debugging Extensions", "CPUID.1:EDX.DE[bit 2]", "common", "", -1},
			3:  {"PSE", "Page Size Extension", "CPUID.1:EDX.PSE[bit 3]", "common", "", -1},
			4:  {"TSC", "Time Stamp Counter", "CPUID.1:EDX.TSC[bit 4]", "common", "", -1},
			5:  {"MSR", "Model Specific Registers", "CPUID.1:EDX.MSR[bit 5]", "common", "", -1},
			6:  {"PAE", "Physical Address Extension", "CPUID.1:EDX.PAE[bit 6]", "common", "", -1},
			7:  {"MCE", "Machine Check Exception", "CPUID.1:EDX.MCE[bit 7]", "common", "", -1},
			8:  {"CX8", "CMPXCHG8 Instruction", "CPUID.1:EDX.CX8[bit 8]", "common", "", -1},
			9:  {"APIC", "APIC On-Chip", "CPUID.1:EDX.APIC[bit 9]", "common", "", -1},
			11: {"SEP", "SYSENTER/SYSEXIT instructions", "CPUID.1:EDX.SEP[bit 11]", "common", "", -1},
			12: {"MTRR", "Memory Type Range Registers", "CPUID.1:EDX.MTRR[bit 12]", "common", "", -1},
			13: {"PGE", "Page Global Enable", "CPUID.1:EDX.PGE[bit 13]", "common", "", -1},
			14: {"MCA", "Machine Check Architecture", "CPUID.1:EDX.MCA[bit 14]", "common", "", -1},
			15: {"CMOV", "Conditional Move Instructions", "CPUID.1:EDX.CMOV[bit 15]", "common", "", -1},
			16: {"PAT", "Page Attribute Table", "CPUID.1:EDX.PAT[bit 16]", "common", "", -1},
			17: {"PSE-36", "36-bit Page Size Extension", "CPUID.1:EDX.PSE-36[bit 17]", "common", "", -1},
			18: {"PSN", "Processor Serial Number", "CPUID.1:EDX.PSN[bit 18]", "intel", "", -1},
			19: {"CLFSH", "CLFLUSH instruction", "CPUID.1:EDX.CLFSH[bit 19]", "common", "", -1},
			20: {"NX", "Execute Disable Bit", "CPUID.1:EDX.NX[bit 20]", "amd", "", -1},                                  // AMD specific
			21: {"ACPI", "Thermal Monitor and Clock Control", "CPUID.1:EDX.ACPI[bit 21]", "intel", "AMDExtendedECX", 1}, // equivalent to Cool'n'Quiet
			22: {"MMX", "Intel MMX Technology", "CPUID.1:EDX.MMX[bit 22]", "common", "", -1},
			23: {"FXSR", "FXSAVE and FXRSTOR Instructions", "CPUID.1:EDX.FXSR[bit 23]", "common", "", -1},
			24: {"SSE", "Streaming SIMD Extensions", "CPUID.1:EDX.SSE[bit 24]", "common", "", -1},
			25: {"SSE2", "Streaming SIMD Extensions 2", "CPUID.1:EDX.SSE2[bit 25]", "common", "", -1},
			26: {"SS", "Self Snoop", "CPUID.1:EDX.SS[bit 26]", "intel", "", -1},
			27: {"HTT", "Multi-threading", "CPUID.1:EDX.HTT[bit 27]", "common", "", -1},
			28: {"TM", "Thermal Monitor", "CPUID.1:EDX.TM[bit 28]", "intel", "", -1},
			29: {"PBE", "Pending Break Enable", "CPUID.1:EDX.PBE[bit 29]", "intel", "", -1},
			30: {"MCA_OVERFLOW", "MCA Overflow Recovery", "CPUID.1:EDX.MCA_OVERFLOW[bit 30]", "amd", "", -1},  // AMD specific
			31: {"PBE_EXTERNAL", "Pending Break External", "CPUID.1:EDX.PBE_EXTERNAL[bit 31]", "amd", "", -1}, // AMD specific
		},
	}, "ExtendedEBX": {
		name:     "Extended Features EBX",
		leaf:     7,
		subleaf:  0,
		register: 1,
		group:    "Basic CPU",
		features: map[int]Feature{
			0:  {"FSGSBASE", "Access to base of %fs and %gs", "CPUID.7.0:EBX.FSGSBASE[bit 0]", "common", "", -1},
			1:  {"TSC_ADJUST", "TSC adjustment MSR", "CPUID.7.0:EBX.TSC_ADJUST[bit 1]", "common", "", -1},
			2:  {"SGX", "Software Guard Extensions", "CPUID.7.0:EBX.SGX[bit 2]", "intel", "", -1},
			3:  {"BMI1", "Bit Manipulation Instruction Set 1", "CPUID.7.0:EBX.BMI1[bit 3]", "common", "", -1},
			4:  {"HLE", "Hardware Lock Elision", "CPUID.7.0:EBX.HLE[bit 4]", "intel", "", -1},
			5:  {"AVX2", "Advanced Vector Extensions 2", "CPUID.7.0:EBX.AVX2[bit 5]", "common", "", -1},
			6:  {"FDP_EXCPTN_ONLY", "FPU DP only updated on exceptions", "CPUID.7.0:EBX.FDP_EXCPTN_ONLY[bit 6]", "intel", "", -1},
			7:  {"SMEP", "Supervisor Mode Execution Prevention", "CPUID.7.0:EBX.SMEP[bit 7]", "common", "", -1},
			8:  {"BMI2", "Bit Manipulation Instruction Set 2", "CPUID.7.0:EBX.BMI2[bit 8]", "common", "", -1},
			9:  {"ERMS", "Enhanced REP MOVSB/STOSB", "CPUID.7.0:EBX.ERMS[bit 9]", "common", "", -1},
			10: {"INVPCID", "INVPCID instruction", "CPUID.7.0:EBX.INVPCID[bit 10]", "common", "", -1},
			11: {"RTM", "Restricted Transactional Memory", "CPUID.7.0:EBX.RTM[bit 11]", "intel", "", -1},
			12: {"RDT_M", "Resource Director Technology Monitoring", "CPUID.7.0:EBX.RDT_M[bit 12]", "intel", "", -1},
			13: {"DEP_FPU_CS_DS", "Deprecates FPU CS and DS", "CPUID.7.0:EBX.DEP_FPU_CS_DS[bit 13]", "intel", "", -1},
			14: {"MPX", "Memory Protection Extensions", "CPUID.7.0:EBX.MPX[bit 14]", "intel", "", -1},
			15: {"RDT_A", "Resource Director Technology Allocation", "CPUID.7.0:EBX.RDT_A[bit 15]", "intel", "", -1},
			16: {"AVX512F", "AVX-512 Foundation", "CPUID.7.0:EBX.AVX512F[bit 16]", "intel", "", -1},
			17: {"AVX512DQ", "AVX-512 Doubleword and Quadword", "CPUID.7.0:EBX.AVX512DQ[bit 17]", "intel", "", -1},
			18: {"RDSEED", "RDSEED instruction", "CPUID.7.0:EBX.RDSEED[bit 18]", "common", "", -1},
			19: {"ADX", "Multi-Precision Add-Carry Instruction", "CPUID.7.0:EBX.ADX[bit 19]", "common", "", -1},
			20: {"SMAP", "Supervisor Mode Access Prevention", "CPUID.7.0:EBX.SMAP[bit 20]", "common", "", -1},
			21: {"AVX512_IFMA", "AVX-512 Integer Fused Multiply-Add", "CPUID.7.0:EBX.AVX512_IFMA[bit 21]", "intel", "", -1},
			22: {"PCOMMIT", "PCOMMIT instruction", "CPUID.7.0:EBX.PCOMMIT[bit 22]", "intel", "", -1},
			23: {"CLFLUSHOPT", "CLFLUSHOPT instruction", "CPUID.7.0:EBX.CLFLUSHOPT[bit 23]", "common", "", -1},
			24: {"CLWB", "CLWB instruction", "CPUID.7.0:EBX.CLWB[bit 24]", "common", "", -1},
			25: {"INTEL_PT", "Intel Processor Trace", "CPUID.7.0:EBX.INTEL_PT[bit 25]", "intel", "", -1},
			26: {"AVX512PF", "AVX-512 Prefetch", "CPUID.7.0:EBX.AVX512PF[bit 26]", "intel", "", -1},
			27: {"AVX512ER", "AVX-512 Exponential and Reciprocal", "CPUID.7.0:EBX.AVX512ER[bit 27]", "intel", "", -1},
			28: {"AVX512CD", "AVX-512 Conflict Detection", "CPUID.7.0:EBX.AVX512CD[bit 28]", "intel", "", -1},
			29: {"SHA", "SHA Extensions", "CPUID.7.0:EBX.SHA[bit 29]", "common", "", -1},
			30: {"AVX512BW", "AVX-512 Byte and Word", "CPUID.7.0:EBX.AVX512BW[bit 30]", "intel", "", -1},
			31: {"AVX512VL", "AVX-512 Vector Length Extensions", "CPUID.7.0:EBX.AVX512VL[bit 31]", "intel", "", -1},
		},
	}, "ExtendedECX": {
		name:     "Extended Features ECX",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Basic CPU",
		features: map[int]Feature{
			0:  {"PREFETCHWT1", "PREFETCHWT1 instruction", "CPUID.7.0:ECX.PREFETCHWT1[bit 0]", "common", "", -1},
			1:  {"AVX512_VBMI", "AVX-512 Vector Bit Manipulation Instructions", "CPUID.7.0:ECX.AVX512_VBMI[bit 1]", "intel", "", -1},
			2:  {"UMIP", "User Mode Instruction Prevention", "CPUID.7.0:ECX.UMIP[bit 2]", "common", "", -1},
			3:  {"PKU", "Memory Protection Keys for User-mode", "CPUID.7.0:ECX.PKU[bit 3]", "intel", "", -1},
			4:  {"OSPKE", "OS Protection Keys Enable", "CPUID.7.0:ECX.OSPKE[bit 4]", "intel", "", -1},
			5:  {"WAITPKG", "TPAUSE, UMONITOR, UMWAIT", "CPUID.7.0:ECX.WAITPKG[bit 5]", "intel", "AMDExtendedECX", 26}, // equivalent to MWAITX
			6:  {"AVX512_VBMI2", "AVX-512 Vector Bit Manipulation Instructions 2", "CPUID.7.0:ECX.AVX512_VBMI2[bit 6]", "intel", "", -1},
			7:  {"CET_SS", "Control Flow Enforcement Shadow Stack", "CPUID.7.0:ECX.CET_SS[bit 7]", "common", "", -1},
			8:  {"GFNI", "Galois Field instructions", "CPUID.7.0:ECX.GFNI[bit 8]", "common", "", -1},
			9:  {"VAES", "Vector AES instructions", "CPUID.7.0:ECX.VAES[bit 9]", "common", "", -1},
			10: {"VPCLMULQDQ", "Vector CLMUL instruction", "CPUID.7.0:ECX.VPCLMULQDQ[bit 10]", "common", "", -1},
			11: {"AVX512_VNNI", "AVX-512 Vector Neural Network Instructions", "CPUID.7.0:ECX.AVX512_VNNI[bit 11]", "intel", "", -1},
			12: {"AVX512_BITALG", "AVX-512 BITALG instructions", "CPUID.7.0:ECX.AVX512_BITALG[bit 12]", "intel", "", -1},
			13: {"TME", "Total Memory Encryption", "CPUID.7.0:ECX.TME[bit 13]", "intel", "AMDExtendedECX", 7}, // equivalent to SME
			14: {"AVX512_VPOPCNTDQ", "AVX-512 Vector Population Count D/Q", "CPUID.7.0:ECX.AVX512_VPOPCNTDQ[bit 14]", "intel", "", -1},
			15: {"LA57", "5-level page tables", "CPUID.7.0:ECX.LA57[bit 15]", "common", "", -1},
			16: {"RDPID", "Read Processor ID", "CPUID.7.0:ECX.RDPID[bit 16]", "common", "", -1},
			17: {"KL", "Key Locker", "CPUID.7.0:ECX.KL[bit 17]", "intel", "", -1},
			18: {"CLDEMOTE", "Cache Line Demote", "CPUID.7.0:ECX.CLDEMOTE[bit 18]", "common", "", -1},
			19: {"MOVDIRI", "MOVDIRI instruction", "CPUID.7.0:ECX.MOVDIRI[bit 19]", "intel", "", -1},
			20: {"MOVDIR64B", "MOVDIR64B instruction", "CPUID.7.0:ECX.MOVDIR64B[bit 20]", "intel", "", -1},
			21: {"ENQCMD", "Enqueue Command", "CPUID.7.0:ECX.ENQCMD[bit 21]", "intel", "", -1},
			22: {"UINTR", "User Interrupts", "CPUID.7.0:ECX.UINTR[bit 22]", "intel", "", -1},
			23: {"TILE", "Tile computation on matrix", "CPUID.7.0:ECX.TILE[bit 23]", "intel", "", -1},
			24: {"AMX_BF16", "AMX bfloat16 Support", "CPUID.7.0:ECX.AMX_BF16[bit 24]", "intel", "", -1},
			25: {"SPEC_CTRL", "Speculation Control", "CPUID.7.0:ECX.SPEC_CTRL[bit 25]", "common", "", -1},
			26: {"STIBP", "Single Thread Indirect Branch Predictors", "CPUID.7.0:ECX.STIBP[bit 26]", "common", "AMDExtendedECX", 17}, // equivalent to AMD's implementation of STIBP
			27: {"L1D_FLUSH", "L1 Data Cache Flush", "CPUID.7.0:ECX.L1D_FLUSH[bit 27]", "common", "", -1},
			28: {"IA32_ARCH_CAPS", "IA32_ARCH_CAPABILITIES MSR", "CPUID.7.0:ECX.IA32_ARCH_CAPS[bit 28]", "common", "", -1},
			29: {"IA32_CORE_CAPS", "IA32_CORE_CAPABILITIES MSR", "CPUID.7.0:ECX.IA32_CORE_CAPS[bit 29]", "common", "", -1},
			30: {"SSBD", "Speculative Store Bypass Disable", "CPUID.7.0:ECX.SSBD[bit 30]", "common", "AMDExtendedECX", 24}, // equivalent to AMD's SSBD
			31: {"IBRS_IBPB", "Indirect Branch Restricted Speculation", "CPUID.7.0:ECX.IBRS_IBPB[bit 31]", "amd", "", -1},  // AMD specific
		},
	}, "AMDExtendedECX": {
		name:      "AMD Extended Features ECX",
		leaf:      0x80000001,
		subleaf:   0,
		register:  2,
		group:     "AMD",
		condition: func(f uint32) bool { return isAMD },
		features: map[int]Feature{
			0:  {"LAHF_LM", "LAHF/SAHF in long mode", "CPUID.80000001H:ECX.LAHF_LM[bit 0]", "amd", "", -1},
			1:  {"CMP_LEGACY", "Core multi-processing legacy mode", "CPUID.80000001H:ECX.CMP_LEGACY[bit 1]", "amd", "", -1},
			2:  {"SVM", "Secure Virtual Machine", "CPUID.80000001H:ECX.SVM[bit 2]", "amd", "StandardECX", 5}, // equivalent to Intel VMX
			3:  {"EXTAPIC", "Extended APIC space", "CPUID.80000001H:ECX.EXTAPIC[bit 3]", "amd", "", -1},
			4:  {"CR8_LEGACY", "CR8 in 32-bit mode", "CPUID.80000001H:ECX.CR8_LEGACY[bit 4]", "amd", "", -1},
			5:  {"ABM", "Advanced bit manipulation", "CPUID.80000001H:ECX.ABM[bit 5]", "amd", "", -1},
			6:  {"SSE4A", "SSE4a", "CPUID.80000001H:ECX.SSE4A[bit 6]", "amd", "StandardECX", 18}, // partial equivalent to Intel SSE4.1
			7:  {"MISALIGNSSE", "Misaligned SSE mode", "CPUID.80000001H:ECX.MISALIGNSSE[bit 7]", "amd", "", -1},
			8:  {"3DNOWPREFETCH", "3DNow prefetch instructions", "CPUID.80000001H:ECX.3DNOWPREFETCH[bit 8]", "amd", "", -1},
			9:  {"OSVW", "OS Visible Workaround", "CPUID.80000001H:ECX.OSVW[bit 9]", "amd", "", -1},
			10: {"IBS", "Instruction Based Sampling", "CPUID.80000001H:ECX.IBS[bit 10]", "amd", "", -1},
			11: {"XOP", "Extended Operations", "CPUID.80000001H:ECX.XOP[bit 11]", "amd", "ExtendedEBX", 11}, // similar concept to Intel RTM
			12: {"SKINIT", "SKINIT/STGI instructions", "CPUID.80000001H:ECX.SKINIT[bit 12]", "amd", "", -1},
			13: {"WDT", "Watchdog timer", "CPUID.80000001H:ECX.WDT[bit 13]", "amd", "", -1},
			14: {"LWP", "Light Weight Profiling", "CPUID.80000001H:ECX.LWP[bit 14]", "amd", "", -1},
			15: {"FMA4", "4-operand FMA instructions", "CPUID.80000001H:ECX.FMA4[bit 15]", "amd", "", -1},
			16: {"TCE", "Translation Cache Extension", "CPUID.80000001H:ECX.TCE[bit 16]", "amd", "", -1},
			17: {"NODEID_MSR", "NodeId MSR", "CPUID.80000001H:ECX.NODEID_MSR[bit 17]", "amd", "", -1},
			19: {"TBM", "Trailing Bit Manipulation", "CPUID.80000001H:ECX.TBM[bit 19]", "amd", "", -1},
			20: {"TOPOEXT", "Topology Extensions", "CPUID.80000001H:ECX.TOPOEXT[bit 20]", "amd", "", -1},
			21: {"PERFCTR_CORE", "Core performance counter extensions", "CPUID.80000001H:ECX.PERFCTR_CORE[bit 21]", "amd", "", -1},
			22: {"PERFCTR_NB", "NB performance counter extensions", "CPUID.80000001H:ECX.PERFCTR_NB[bit 22]", "amd", "", -1},
			23: {"DBX", "Data breakpoint extensions", "CPUID.80000001H:ECX.DBX[bit 23]", "amd", "", -1},
			24: {"PERFTSC", "Performance time-stamp counter", "CPUID.80000001H:ECX.PERFTSC[bit 24]", "amd", "", -1},
			25: {"PCX_L2I", "L2I performance counter extensions", "CPUID.80000001H:ECX.PCX_L2I[bit 25]", "amd", "", -1},
			26: {"MWAITX", "MONITORX/MWAITX instructions", "CPUID.80000001H:ECX.MWAITX[bit 26]", "amd", "ExtendedECX", 5}, // equivalent to Intel WAITPKG
			27: {"ADDR_MASK_EXT", "Address mask extension for instruction breakpoint", "CPUID.80000001H:ECX.ADDR_MASK_EXT[bit 27]", "amd", "", -1},
			28: {"MONITORX", "MONITORX/MWAITX instructions", "CPUID.80000001H:ECX.MONITORX[bit 28]", "amd", "", -1},
			29: {"PSFD", "Predictive Store Forward Disable", "CPUID.80000001H:ECX.PSFD[bit 29]", "amd", "", -1},
			30: {"IBPB", "Indirect Branch Prediction Barrier", "CPUID.80000001H:ECX.IBPB[bit 30]", "amd", "", -1},
			31: {"IBRS", "Indirect Branch Restricted Speculation", "CPUID.80000001H:ECX.IBRS[bit 31]", "amd", "", -1},
		},
	}, "PowerManagement": {
		name:     "Power Management Features",
		leaf:     6,
		subleaf:  0,
		register: 0,
		group:    "Power Management",
		features: map[int]Feature{
			0:  {"DTHERM", "Digital Thermal Sensor", "CPUID.6:EAX.DTHERM[bit 0]", "common", "", -1},
			1:  {"IDA", "Intel Dynamic Acceleration", "CPUID.6:EAX.IDA[bit 1]", "intel", "AMDExtendedECX", 1}, // equivalent to AMD CMP_LEGACY
			2:  {"ARAT", "Always Running APIC Timer", "CPUID.6:EAX.ARAT[bit 2]", "common", "", -1},
			3:  {"PLN", "Power Limit Notification", "CPUID.6:EAX.PLN[bit 3]", "intel", "", -1},
			4:  {"ECMD", "Extended Clock Modulation Duty", "CPUID.6:EAX.ECMD[bit 4]", "intel", "", -1},
			5:  {"PTM", "Package Thermal Management", "CPUID.6:EAX.PTM[bit 5]", "intel", "", -1},
			6:  {"HWP", "Hardware P-states", "CPUID.6:EAX.HWP[bit 6]", "intel", "AMDExtendedECX", 14}, // equivalent to AMD LWP
			7:  {"HWP_NOTIFY", "HWP Notification", "CPUID.6:EAX.HWP_NOTIFY[bit 7]", "intel", "", -1},
			8:  {"HWP_ACTIVITY", "HWP Activity Window", "CPUID.6:EAX.HWP_ACTIVITY[bit 8]", "intel", "", -1},
			9:  {"HWP_EPP", "HWP Energy Performance Preference", "CPUID.6:EAX.HWP_EPP[bit 9]", "intel", "", -1},
			10: {"HWP_PLR", "HWP Package Level Request", "CPUID.6:EAX.HWP_PLR[bit 10]", "intel", "", -1},
			11: {"HDC", "Hardware Duty Cycling", "CPUID.6:EAX.HDC[bit 11]", "intel", "", -1},
			12: {"TURBO3", "Intel Turbo Boost Max Technology 3.0", "CPUID.6:EAX.TURBO3[bit 12]", "intel", "", -1},
			13: {"HWP_CAP", "HWP Capabilities", "CPUID.6:EAX.HWP_CAP[bit 13]", "intel", "", -1},
			14: {"HWP_PECI", "HWP PECI override", "CPUID.6:EAX.HWP_PECI[bit 14]", "intel", "", -1},
			15: {"HWP_FLEX", "Flexible HWP", "CPUID.6:EAX.HWP_FLEX[bit 15]", "intel", "", -1},
			16: {"HWP_FAST", "Fast access mode for HWP", "CPUID.6:EAX.HWP_FAST[bit 16]", "intel", "", -1},
			17: {"HWFB", "HW Feedback Structure", "CPUID.6:EAX.HWFB[bit 17]", "intel", "", -1},
			18: {"HWP_REQUEST", "Ignoring Idle Logical Processor HWP request", "CPUID.6:EAX.HWP_REQUEST[bit 18]", "intel", "", -1},
			// AMD specific power features in same leaf
			19: {"CPB", "Core Performance Boost", "CPUID.6:EAX.CPB[bit 19]", "amd", "", -1},
			20: {"EFRO", "Energy Frequency Optimized Core", "CPUID.6:EAX.EFRO[bit 20]", "amd", "", -1},
			21: {"PFE", "Preferred Frequency Enable", "CPUID.6:EAX.PFE[bit 21]", "amd", "", -1},
			22: {"RAPL", "Running Average Power Limit", "CPUID.6:EAX.RAPL[bit 22]", "common", "", -1},
		},
	}, "SGX": {
		name:      "Software Guard Extensions",
		leaf:      0x12,
		subleaf:   0,
		register:  0,
		group:     "Security",
		condition: func(f uint32) bool { return (extb>>2)&1 == 1 }, // Checks if SGX is supported via CPUID.7:EBX[2]
		features: map[int]Feature{
			0: {"SGX1", "SGX1 instruction set", "CPUID.12H:EAX.SGX1[bit 0]", "intel", "", -1},
			1: {"SGX2", "SGX2 instruction set", "CPUID.12H:EAX.SGX2[bit 1]", "intel", "", -1},
			2: {"ENCLV", "ENCLV instruction leaves", "CPUID.12H:EAX.ENCLV[bit 2]", "intel", "", -1},
			3: {"ENCLS", "ENCLS instruction leaves", "CPUID.12H:EAX.ENCLS[bit 3]", "intel", "", -1},
			4: {"ENCLU", "ENCLU instruction leaves", "CPUID.12H:EAX.ENCLU[bit 4]", "intel", "", -1},
			5: {"EUPDATERET", "EUPDATERET leaf function", "CPUID.12H:EAX.EUPDATERET[bit 5]", "intel", "", -1},
			6: {"EDECCSSA", "EDECCSSA leaf function", "CPUID.12H:EAX.EDECCSSA[bit 6]", "intel", "", -1},
			7: {"PROVISIONKEY", "Provision Key leaf function", "CPUID.12H:EAX.PROVISIONKEY[bit 7]", "intel", "AMDExtendedECX", 12}, // Equivalent to AMD SKINIT
			8: {"TOKENKEY", "Token Key leaf function", "CPUID.12H:EAX.TOKENKEY[bit 8]", "intel", "", -1},
			9: {"EINITTOKEN", "EINIT Token functionality", "CPUID.12H:EAX.EINITTOKEN[bit 9]", "intel", "", -1},
			// AMD Memory Encryption features that correspond to SGX functionality
			10: {"SEV", "Secure Encrypted Virtualization", "CPUID.12H:EAX.SEV[bit 10]", "amd", "", -1},
			11: {"SEV_ES", "SEV Encrypted State", "CPUID.12H:EAX.SEV_ES[bit 11]", "amd", "", -1},
			12: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.12H:EAX.SEV_SNP[bit 12]", "amd", "", -1},
			13: {"VMPL", "Virtual Machine Privilege Levels", "CPUID.12H:EAX.VMPL[bit 13]", "amd", "", -1},
		},
	}, "PT": {
		name:      "Processor Trace Features",
		leaf:      0x14,
		subleaf:   0,
		register:  1,
		group:     "Debugging",
		condition: func(f uint32) bool { return (extb>>25)&1 == 1 }, // Checks Intel PT support via CPUID.7:EBX[25]
		features: map[int]Feature{
			0: {"PT_CR3_FILTERING", "CR3 filtering support", "CPUID.14H:EBX.CR3_FILTERING[bit 0]", "intel", "", -1},
			1: {"PT_CONFIGURABLE_PSB", "Configurable PSB support", "CPUID.14H:EBX.CONFIGURABLE_PSB[bit 1]", "intel", "", -1},
			2: {"PT_IP_FILTERING", "IP filtering support", "CPUID.14H:EBX.IP_FILTERING[bit 2]", "intel", "", -1},
			3: {"PT_MTC", "MTC support", "CPUID.14H:EBX.MTC[bit 3]", "intel", "", -1},
			4: {"PT_PTWRITE", "PTWRITE support", "CPUID.14H:EBX.PTWRITE[bit 4]", "intel", "", -1},
			5: {"PT_POWER_EVENT_TRACE", "Power Event Trace support", "CPUID.14H:EBX.POWER_EVENT_TRACE[bit 5]", "intel", "", -1},
			6: {"PT_PSB_PMI", "PSB and PMI preservation", "CPUID.14H:EBX.PSB_PMI[bit 6]", "intel", "", -1},
			7: {"PT_EVENT_TRACE", "Event Trace support", "CPUID.14H:EBX.EVENT_TRACE[bit 7]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			8: {"PT_TNT_DISABLE", "TNT disable support", "CPUID.14H:EBX.TNT_DISABLE[bit 8]", "intel", "", -1},
		},
	}, "MTRR": { // Corrected name from MTTR to MTRR (Memory Type Range Registers)
		name:     "Memory Type Range Register Features",
		leaf:     0x0B,
		subleaf:  0,
		register: 0,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"MTRR_VCNT", "Variable range registers count", "CPUID.MTRRcap:EAX.VCNT[bits 7-0]", "common", "", -1},
			1: {"MTRR_FIX", "Fixed range registers", "CPUID.MTRRcap:EAX.FIX[bit 8]", "common", "", -1},
			2: {"MTRR_WC", "Write-combining memory type", "CPUID.MTRRcap:EAX.WC[bit 9]", "common", "", -1},
			3: {"MTRR_SMRR", "System management range registers", "CPUID.MTRRcap:EAX.SMRR[bit 11]", "common", "", -1},
			4: {"MTRR_UC", "Uncacheable memory type support", "CPUID.MTRRcap:EAX.UC[bit 10]", "common", "", -1},
			5: {"MTRR_ENH_EXT", "Enhanced MTRR extension", "CPUID.MTRRcap:EAX.ENH_EXT[bit 12]", "common", "", -1},
			6: {"MTRR_MEM_CLEAR", "Memory clear feature", "CPUID.MTRRcap:EAX.MEM_CLEAR[bit 13]", "common", "", -1},
		},
	}, "CacheFeatures": {
		name:     "Cache Properties",
		leaf:     4,
		subleaf:  0,
		register: 3,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0:  {"CACHE_SELF_SNOOP", "Self-snooping support", "CPUID.4:EDX.SELF_SNOOP[bit 0]", "common", "", -1},
			1:  {"CACHE_INCLUSIVENESS", "Cache inclusiveness", "CPUID.4:EDX.INCLUSIVE[bit 1]", "common", "", -1},
			2:  {"CACHE_COMPLEX_INDEX", "Complex cache indexing", "CPUID.4:EDX.COMPLEX_INDEX[bit 2]", "common", "", -1},
			3:  {"CACHE_LEVEL", "Cache level", "CPUID.4:EAX.CACHE_LEVEL[bits 7-5]", "common", "", -1},
			4:  {"CACHE_TYPE", "Cache type", "CPUID.4:EAX.CACHE_TYPE[bits 4-0]", "common", "", -1},
			5:  {"CACHE_SIZE", "Cache size in bytes", "CPUID.4:EBX", "common", "", -1},
			6:  {"CACHE_PREFETCH", "Hardware prefetch", "CPUID.4:EDX.PREFETCH[bit 4]", "common", "", -1},
			7:  {"L3_CACHE_WAYS", "L3 Cache ways of associativity", "CPUID.4:EAX.L3_WAYS[bits 31-22]", "common", "", -1},
			8:  {"L3_CACHE_PARTITIONING", "L3 Cache partitioning support", "CPUID.4:EDX.L3_PART[bit 3]", "intel", "AMDExtendedECX", 20}, // equivalent to AMD TOPOEXT
			9:  {"CACHE_SHARING", "Cache line sharing", "CPUID.4:EAX.CACHE_SHARING[bits 25-14]", "common", "", -1},
			10: {"WBINVD", "WBINVD/WBNOINVD support", "CPUID.4:EDX.WBINVD[bit 5]", "common", "", -1},
			// AMD specific cache features
			11: {"L3_NOT_USED", "L3 cache not used", "CPUID.8000001DH:EAX.L3_NOT_USED[bit 6]", "amd", "", -1},
			12: {"COMPUTE_UNIT_ID", "AMD Compute Unit ID", "CPUID.8000001DH:EAX.COMPUTE_UNIT_ID[bits 15-8]", "amd", "", -1},
		},
	}, "XSave": {
		name:     "Extended State Features (XSAVE)",
		leaf:     0xD,
		subleaf:  0,
		register: 0,
		features: map[int]Feature{
			0: {"XSAVEOPT", "XSAVEOPT instruction", "CPUID.0DH:EAX.XSAVEOPT[bit 0]", "common", "", -1},
			1: {"XSAVEC", "XSAVEC instruction", "CPUID.0DH:EAX.XSAVEC[bit 1]", "common", "", -1},
			2: {"XGETBV_ECX1", "XGETBV with ECX=1", "CPUID.0DH:EAX.XGETBV_ECX1[bit 2]", "common", "", -1},
			3: {"XSAVES", "XSAVES/XRSTORS instructions", "CPUID.0DH:EAX.XSAVES[bit 3]", "common", "", -1},
			4: {"XFD", "Extended Feature Disable", "CPUID.0DH:EAX.XFD[bit 4]", "common", "", -1},
			// Additional state components
			5:  {"XSAVE_YMM", "YMM state support", "CPUID.0DH:EAX.YMM[bit 5]", "common", "", -1},
			6:  {"XSAVE_BNDREGS", "MPX bound register state", "CPUID.0DH:EAX.BNDREGS[bit 6]", "intel", "", -1},
			7:  {"XSAVE_BNDCSR", "MPX bound config state", "CPUID.0DH:EAX.BNDCSR[bit 7]", "intel", "", -1},
			8:  {"XSAVE_OPMASK", "AVX-512 opmask state", "CPUID.0DH:EAX.OPMASK[bit 8]", "intel", "", -1},
			9:  {"XSAVE_ZMM_HI256", "AVX-512 ZMM_Hi256 state", "CPUID.0DH:EAX.ZMM_HI256[bit 9]", "intel", "", -1},
			10: {"XSAVE_HI16_ZMM", "AVX-512 Hi16_ZMM state", "CPUID.0DH:EAX.HI16_ZMM[bit 10]", "intel", "", -1},
			11: {"XSAVE_PKRU", "PKRU state", "CPUID.0DH:EAX.PKRU[bit 11]", "intel", "", -1},
			// AMD specific XSAVE features
			12: {"XSAVE_CET_U", "CET user state", "CPUID.0DH:EAX.CET_U[bit 12]", "common", "", -1},
			13: {"XSAVE_CET_S", "CET supervisor state", "CPUID.0DH:EAX.CET_S[bit 13]", "common", "", -1},
			14: {"XSAVE_HDC", "HDC state", "CPUID.0DH:EAX.HDC[bit 14]", "intel", "", -1},
			15: {"XSAVE_UINTR", "User interrupt state", "CPUID.0DH:EAX.UINTR[bit 15]", "intel", "", -1},
			16: {"XSAVE_LBR", "LBR state", "CPUID.0DH:EAX.LBR[bit 16]", "intel", "", -1},
			17: {"XSAVE_HWP", "HWP state", "CPUID.0DH:EAX.HWP[bit 17]", "intel", "", -1},
		},
	}, "PerformanceMonitor": {
		name:     "Performance Monitor",
		leaf:     0xA,
		subleaf:  0,
		register: 0,
		group:    "Performance Monitoring",
		features: map[int]Feature{
			// EAX Features
			0: {"PMC_VERSION", "Version ID of performance monitoring", "CPUID.0AH:EAX[bits 7-0]", "common", "", -1},
			1: {"PMC_GP_COUNTER", "Number of general-purpose counters", "CPUID.0AH:EAX[bits 15-8]", "common", "", -1},
			2: {"PMC_GP_COUNTER_WIDTH", "Bit width of general-purpose counters", "CPUID.0AH:EAX[bits 23-16]", "common", "", -1},
			3: {"PMC_EBX_LENGTH", "Length of EBX bit vector", "CPUID.0AH:EAX[bits 31-24]", "common", "", -1},
			// EDX Features
			4:  {"PMC_FIXED_COUNTER", "Number of fixed-function counters", "CPUID.0AH:EDX[bits 4-0]", "common", "", -1},
			5:  {"PMC_FIXED_COUNTER_WIDTH", "Bit width of fixed-function counters", "CPUID.0AH:EDX[bits 12-5]", "common", "", -1},
			6:  {"PMC_PEBS", "Precise Event Based Sampling", "CPUID.0AH:EDX[bit 13]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			7:  {"PMC_PERF_TSC", "Read Performance TSC", "CPUID.0AH:EDX[bit 14]", "intel", "AMDExtendedECX", 24},     // equivalent to AMD PERFTSC
			8:  {"PMC_LBR_FMT", "LBR Format Support", "CPUID.0AH:EDX[bit 15]", "intel", "", -1},
			9:  {"PMC_PERFCTR_CORE_CYCLES", "Core Cycles Event Available", "CPUID.0AH:EDX[bit 16]", "common", "", -1},
			10: {"PMC_PERFCTR_INST_RET", "Instructions Retired Event", "CPUID.0AH:EDX[bit 17]", "common", "", -1},
			11: {"PMC_PERFCTR_REF_CYCLES", "Reference Cycles Event", "CPUID.0AH:EDX[bit 18]", "common", "", -1},
			12: {"PMC_CACHE_MISSES", "Last Level Cache References/Misses", "CPUID.0AH:EDX[bit 19]", "common", "", -1},
			13: {"PMC_BRANCH_INST_RET", "Branch Instructions Retired", "CPUID.0AH:EDX[bit 20]", "common", "", -1},
			14: {"PMC_BRANCH_MISPREDICT_RET", "Branch Mispredict Retired", "CPUID.0AH:EDX[bit 21]", "common", "", -1},
			// AMD specific features
			15: {"PMC_NB_EVENTS", "North Bridge Performance Events", "CPUID.0AH:EDX[bit 22]", "amd", "", -1},
			16: {"PMC_CORE_EVENTS", "Core Performance Events", "CPUID.0AH:EDX[bit 23]", "amd", "", -1},
		},
	}, "Virtualization": {
		name:     "Virtualization",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "Virtualization",
		features: map[int]Feature{
			0:  {"VMX_ROOT", "VMX Root Operations", "CPUID.1:ECX.VMX[bit 0]", "intel", "AMDExtendedECX", 2}, // equivalent to AMD SVM
			1:  {"VMX_EPT", "Extended Page Tables", "CPUID.1:ECX.EPT[bit 1]", "intel", "AMDExtendedECX", 3}, // equivalent to AMD NESTED_PAGING
			2:  {"VMX_VPID", "Virtual Processor IDs", "CPUID.1:ECX.VPID[bit 2]", "intel", "", -1},
			3:  {"VMX_UNRESTRICTED", "Unrestricted Guest", "CPUID.1:ECX.UG[bit 3]", "intel", "", -1},
			4:  {"VMX_PREEMPTION", "VMX Preemption Timer", "CPUID.1:ECX.VMXPT[bit 4]", "intel", "", -1},
			5:  {"VMX_POSTED_INTR", "Posted Interrupts", "CPUID.1:ECX.VMXPI[bit 5]", "intel", "", -1},
			6:  {"VMX_VNMI", "Virtual NMIs", "CPUID.1:ECX.VNMI[bit 6]", "intel", "", -1},
			7:  {"VMX_TRUE_MSR", "True MSR Interface", "CPUID.1:ECX.VMXTMSR[bit 7]", "intel", "", -1},
			8:  {"VMX_EPT_A_ONLY", "EPT Access-only", "CPUID.1:ECX.EPTAO[bit 8]", "intel", "", -1},
			9:  {"VMX_VINTR_PENDING", "Virtual Interrupt Pending", "CPUID.1:ECX.VIP[bit 9]", "intel", "", -1},
			10: {"VMX_EPT_MC", "EPT MC Bits Control", "CPUID.1:ECX.EPTMC[bit 10]", "intel", "", -1},
			11: {"VMX_RDRAND_EXITING", "RDRAND Exiting", "CPUID.1:ECX.VMXRDRAND[bit 11]", "intel", "", -1},
			12: {"VMX_INVPCID", "INVPCID Support", "CPUID.1:ECX.VMXINVPCID[bit 12]", "intel", "", -1},
			13: {"VMX_VMFUNC", "VM Functions", "CPUID.1:ECX.VMFUNC[bit 13]", "intel", "", -1},
			14: {"VMX_SHADOW_VMCS", "Shadow VMCS", "CPUID.1:ECX.SVMCS[bit 14]", "intel", "", -1},
			15: {"VMX_ENCLS", "ENCLS Exiting", "CPUID.1:ECX.VMXENCLS[bit 15]", "intel", "", -1},
			// AMD specific features
			16: {"SVM_NESTED_PAGING", "SVM Nested Paging", "CPUID.8000000AH:EDX[bit 0]", "amd", "", -1},
			17: {"SVM_LBR_VIRT", "SVM LBR Virtualization", "CPUID.8000000AH:EDX[bit 1]", "amd", "", -1},
			18: {"SVM_NRIPS", "SVM Next RIP Save", "CPUID.8000000AH:EDX[bit 3]", "amd", "", -1},
			19: {"SVM_VMCB_CLEAN", "SVM VMCB Clean Bits", "CPUID.8000000AH:EDX[bit 4]", "amd", "", -1},
		},
	}, "ExtendedSecurity": {
		name:     "Extended Security",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Security",
		features: map[int]Feature{
			0:  {"SMAP", "Supervisor Mode Access Prevention", "CPUID.7:EBX.SMAP[bit 0]", "common", "", -1},
			1:  {"SMEP", "Supervisor Mode Execution Prevention", "CPUID.7:EBX.SMEP[bit 1]", "common", "", -1},
			2:  {"UMIP", "User Mode Instruction Prevention", "CPUID.7:ECX.UMIP[bit 2]", "common", "", -1},
			3:  {"PKU", "Protection Keys for User-Mode Pages", "CPUID.7:ECX.PKU[bit 3]", "intel", "", -1},
			4:  {"IBT", "Indirect Branch Tracking", "CPUID.7:EDX.IBT[bit 4]", "common", "", -1},
			5:  {"SHSTK", "Shadow Stack", "CPUID.7:ECX.SHSTK[bit 5]", "common", "", -1},
			6:  {"SRBDS_CTRL", "SRBDS Mitigation MSR", "CPUID.7:EDX.SRBDS_CTRL[bit 6]", "intel", "", -1},
			7:  {"MD_CLEAR", "VERW Clear CPU Buffers", "CPUID.7:EDX.MD_CLEAR[bit 7]", "common", "", -1},
			8:  {"TSX_FORCE_ABORT", "TSX Force Abort", "CPUID.7:EDX.TSX_FORCE_ABORT[bit 8]", "intel", "", -1},
			9:  {"SERIALIZE", "Serialize Instruction", "CPUID.7:EDX.SERIALIZE[bit 9]", "common", "", -1},
			10: {"HYBRID", "Hybrid CPU", "CPUID.7:EDX.HYBRID[bit 10]", "intel", "", -1},
			11: {"TSXLDTRK", "TSX Load Address Tracking", "CPUID.7:EDX.TSXLDTRK[bit 11]", "intel", "", -1},
			12: {"PCONFIG", "Platform Configuration", "CPUID.7:EDX.PCONFIG[bit 12]", "intel", "", -1},
			13: {"CET_IBT", "Control Flow Enforcement - IBT", "CPUID.7:EDX.CET_IBT[bit 13]", "common", "", -1},
			14: {"CET_SSS", "Control Flow Enforcement - Shadow Stack", "CPUID.7:EDX.CET_SSS[bit 14]", "common", "", -1},
			15: {"KEY_LOCKER", "Key Locker", "CPUID.7:EDX.KEY_LOCKER[bit 15]", "intel", "", -1},
			// AMD specific security features
			16: {"SEV", "Secure Encrypted Virtualization", "CPUID.8000001FH:EAX[bit 1]", "amd", "", -1},
			17: {"SEV_ES", "SEV Encrypted State", "CPUID.8000001FH:EAX[bit 2]", "amd", "", -1},
			18: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000001FH:EAX[bit 3]", "amd", "", -1},
			19: {"VMPL", "VM Permission Levels", "CPUID.8000001FH:EAX[bit 4]", "amd", "", -1},
		},
	}, "PlatformSecurity": {
		name:     "Platform Security",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Security",
		features: map[int]Feature{
			0:  {"TPM", "Trusted Platform Module", "CPUID.7:ECX.TPM[bit 0]", "common", "", -1},
			1:  {"SKINIT", "Secure Init and Jump", "CPUID.8000_0001:ECX.SKINIT[bit 1]", "amd", "", -1},
			2:  {"SEV", "Secure Encrypted Virtualization", "CPUID.8000_0001:ECX.SEV[bit 2]", "amd", "SGX", 0}, // Equivalent to Intel SGX
			3:  {"SEV_ES", "SEV Encrypted State", "CPUID.8000_0001:ECX.SEV_ES[bit 3]", "amd", "", -1},
			4:  {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000_0001:ECX.SEV_SNP[bit 4]", "amd", "", -1},
			5:  {"VMPL", "Virtual Machine Privilege Levels", "CPUID.8000_0001:ECX.VMPL[bit 5]", "amd", "", -1},
			6:  {"VMSA_REGPROT", "VMSA Register Protection", "CPUID.8000_0001:ECX.VMSA_REGPROT[bit 6]", "amd", "", -1},
			7:  {"SME", "Secure Memory Encryption", "CPUID.8000_0001:ECX.SME[bit 7]", "amd", "ExtendedECX", 13}, // Equivalent to Intel TME
			8:  {"SME_COHERENT", "SME Coherent Memory", "CPUID.8000_0001:ECX.SME_COHERENT[bit 8]", "amd", "", -1},
			9:  {"TSC_SCALE", "TSC Scaling", "CPUID.8000_0001:ECX.TSC_SCALE[bit 9]", "amd", "", -1},
			10: {"SVME_ADDR_CHECK", "SVME Address Check", "CPUID.8000_0001:ECX.SVME_ADDR_CHECK[bit 10]", "amd", "", -1},
			11: {"SECURE_TSC", "Secure TSC", "CPUID.8000_0001:ECX.SECURE_TSC[bit 11]", "amd", "", -1},
			// Intel-specific platform security features
			12: {"SGX_LC", "SGX Launch Control", "CPUID.7:ECX.SGX_LC[bit 30]", "intel", "", -1},
			13: {"SGX_KEYS", "SGX Key Generation", "CPUID.7:ECX.SGX_KEYS[bit 31]", "intel", "", -1},
		},
	}, "ExtendedDebug": {
		name:     "Extended Debug",
		leaf:     0x15,
		subleaf:  0,
		register: 0,
		group:    "Debugging",
		features: map[int]Feature{
			0:  {"LBR", "Last Branch Record", "CPUID.0x15:EAX.LBR[bit 0]", "common", "", -1},
			1:  {"PEBS", "Precise Event Based Sampling", "CPUID.0x15:EAX.PEBS[bit 1]", "intel", "AMDExtendedECX", 10}, // Equivalent to AMD IBS
			2:  {"PEBS_ARCH", "Architectural PEBS", "CPUID.0x15:EAX.PEBS_ARCH[bit 2]", "intel", "", -1},
			3:  {"PEBS_TRAP", "PEBS Trap Flag", "CPUID.0x15:EAX.PEBS_TRAP[bit 3]", "intel", "", -1},
			4:  {"IPT", "Intel Processor Trace", "CPUID.0x15:EAX.IPT[bit 4]", "intel", "", -1},
			5:  {"BTS", "Branch Trace Store", "CPUID.0x15:EAX.BTS[bit 5]", "intel", "", -1},
			6:  {"PEA", "Precise Event Address", "CPUID.0x15:EAX.PEA[bit 6]", "intel", "", -1},
			7:  {"DS", "Debug Store", "CPUID.0x15:EAX.DS[bit 7]", "intel", "", -1},
			8:  {"PTW", "PTWrite Event", "CPUID.0x15:EAX.PTW[bit 8]", "intel", "", -1},
			9:  {"PSB", "PSB and PAUSE Filtering", "CPUID.0x15:EAX.PSB[bit 9]", "intel", "", -1},
			10: {"IPRED_TRACE", "Indirect Prediction Tracing", "CPUID.0x15:EAX.IPRED_TRACE[bit 10]", "intel", "", -1},
			11: {"MTF", "Monitor Trap Flag", "CPUID.0x15:EAX.MTF[bit 11]", "common", "", -1},
			// AMD specific debug features
			12: {"IBS_FETCH", "IBS Fetch Sampling", "CPUID.8000001BH:EAX[bit 0]", "amd", "", -1},
			13: {"IBS_OP", "IBS Op Sampling", "CPUID.8000001BH:EAX[bit 1]", "amd", "", -1},
			14: {"IBS_RIP", "IBS RIP Invalid", "CPUID.8000001BH:EAX[bit 2]", "amd", "", -1},
			15: {"IBS_BRANCH", "IBS Branch Target Address", "CPUID.8000001BH:EAX[bit 3]", "amd", "", -1},
		},
	}, "ExtendedTopologyEnumeration": {
		name:     "Extended Topology Enumeration",
		leaf:     0x1A,
		subleaf:  0,
		register: 0,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"CORE_TYPE", "Core Type", "CPUID.1AH:EAX.CORE_TYPE[bit 0]", "common", "", -1},
			1: {"CORE_ID", "Core ID", "CPUID.1AH:EAX.CORE_ID[bit 1]", "common", "", -1},
			2: {"SMT_ID", "SMT ID", "CPUID.1AH:EAX.SMT_ID[bit 2]", "common", "", -1},
			3: {"EXTENDED_APIC", "Extended APIC ID", "CPUID.1AH:EAX.EXTENDED_APIC[bit 3]", "common", "", -1},
			4: {"DIE_ID", "Die ID", "CPUID.1AH:EAX.DIE_ID[bit 4]", "common", "", -1},
			5: {"CLUSTER_ID", "Cluster ID", "CPUID.1AH:EAX.CLUSTER_ID[bit 5]", "intel", "AMDExtendedECX", 20}, // Equivalent to AMD TOPOEXT
			6: {"HYBRID_CPU", "Hybrid CPU Support", "CPUID.1AH:EAX.HYBRID_CPU[bit 6]", "intel", "", -1},
			// AMD specific topology features
			7: {"CCD_ID", "CCD ID", "CPUID.8000001EH:EBX[bits 7-0]", "amd", "", -1},
			8: {"CCX_ID", "CCX ID", "CPUID.8000001EH:ECX[bits 7-0]", "amd", "", -1},
			9: {"THREADS_PER_CORE", "Threads per Core", "CPUID.8000001EH:EBX[bits 15-8]", "amd", "", -1},
		},
	}, "MemoryCache": {
		name:     "Memory Cache",
		leaf:     4,
		subleaf:  0,
		register: 0,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"L1_CACHE", "L1 Cache Present", "CPUID.4:EAX.L1_CACHE[bit 0]", "common", "", -1},
			1: {"L2_CACHE", "L2 Cache Present", "CPUID.4:EAX.L2_CACHE[bit 1]", "common", "", -1},
			2: {"L3_CACHE", "L3 Cache Present", "CPUID.4:EAX.L3_CACHE[bit 2]", "common", "", -1},
			3: {"CACHE_INCLUSIVE", "Cache Inclusiveness", "CPUID.4:EAX.CACHE_INCLUSIVE[bit 3]", "common", "", -1},
			4: {"WBINVD", "WBINVD/WBNOINVD Support", "CPUID.4:EAX.WBINVD[bit 4]", "common", "", -1},
			5: {"CACHE_QOS", "Cache QoS Support", "CPUID.4:EAX.CACHE_QOS[bit 5]", "intel", "AMDExtendedECX", 21}, // Equivalent to AMD PERFCTR_CORE
			6: {"SPLIT_LOCK_DETECT", "Split Lock Detection", "CPUID.4:EAX.SPLIT_LOCK_DETECT[bit 6]", "common", "", -1},
			7: {"BUS_LOCK_DETECT", "Bus Lock Detection", "CPUID.4:EAX.BUS_LOCK_DETECT[bit 7]", "common", "", -1},
			// AMD specific cache features
			8:  {"CACHE_TYPE", "Cache Type", "CPUID.8000001DH:EAX[bits 3-0]", "amd", "", -1},
			9:  {"CACHE_LEVEL", "Cache Level", "CPUID.8000001DH:EAX[bits 7-4]", "amd", "", -1},
			10: {"CACHE_SIZE", "Cache Size", "CPUID.8000001DH:EBX", "amd", "", -1},
			11: {"CACHE_WAYS", "Cache Ways", "CPUID.8000001DH:EBX[bits 31-22]", "amd", "", -1},
			12: {"CACHE_PARTITIONING", "Cache Partitioning", "CPUID.8000001DH:EDX[bit 0]", "amd", "", -1},
			// Additional Intel Cache features
			13: {"RDT_M", "Resource Director Technology Monitoring", "CPUID.4:EAX.RDT_M[bit 8]", "intel", "", -1},
			14: {"RDT_A", "Resource Director Technology Allocation", "CPUID.4:EAX.RDT_A[bit 9]", "intel", "", -1},
		},
	}, "ExtendedStateSaveRestore": {
		name:     "Extended State Save/Restore",
		leaf:     0xD,
		subleaf:  1,
		register: 0,
		group:    "State & Register",
		features: map[int]Feature{
			0:  {"MPX_STATE", "MPX State", "CPUID.0DH:EAX.MPX_STATE[bit 0]", "intel", "", -1},
			1:  {"AVX_STATE", "AVX State", "CPUID.0DH:EAX.AVX_STATE[bit 1]", "common", "", -1},
			2:  {"AVX512_STATE", "AVX-512 State", "CPUID.0DH:EAX.AVX512_STATE[bit 2]", "intel", "", -1},
			3:  {"PKRU_STATE", "PKRU State", "CPUID.0DH:EAX.PKRU_STATE[bit 3]", "intel", "", -1},
			4:  {"CET_U_STATE", "CET User State", "CPUID.0DH:EAX.CET_U_STATE[bit 4]", "common", "", -1},
			5:  {"CET_S_STATE", "CET Supervisor State", "CPUID.0DH:EAX.CET_S_STATE[bit 5]", "common", "", -1},
			6:  {"HDC_STATE", "HDC State", "CPUID.0DH:EAX.HDC_STATE[bit 6]", "intel", "", -1},
			7:  {"UINTR_STATE", "User Interrupts State", "CPUID.0DH:EAX.UINTR_STATE[bit 7]", "intel", "", -1},
			8:  {"LBR_STATE", "LBR State", "CPUID.0DH:EAX.LBR_STATE[bit 8]", "common", "", -1},
			9:  {"HWP_STATE", "HWP State", "CPUID.0DH:EAX.HWP_STATE[bit 9]", "intel", "AMDExtendedECX", 14}, // equivalent to AMD LWP
			10: {"XTILECFG", "Tile Configuration State", "CPUID.0DH:EAX.XTILECFG[bit 10]", "intel", "", -1},
			11: {"XTILEDATA", "Tile Data State", "CPUID.0DH:EAX.XTILEDATA[bit 11]", "intel", "", -1},
			// AMD specific extended states
			12: {"MCOMMIT_STATE", "MCOMMIT State", "CPUID.0DH:EAX.MCOMMIT_STATE[bit 12]", "amd", "", -1},
			13: {"XFD_STATE", "Extended Feature Disable State", "CPUID.0DH:EAX.XFD_STATE[bit 13]", "amd", "", -1},
		},
	}, "SGXExtensions": {
		name:      "Software Guard Extensions",
		leaf:      0x12,
		subleaf:   0,
		register:  0,
		group:     "Security",
		condition: func(f uint32) bool { return (extb>>2)&1 == 1 },
		features: map[int]Feature{
			0: {"SGX_LC", "SGX Launch Control", "CPUID.12H:EAX.SGX_LC[bit 0]", "intel", "", -1},
			1: {"SGX_KEYS", "SGX Attestation Keys", "CPUID.12H:EAX.SGX_KEYS[bit 1]", "intel", "", -1},
			2: {"SGX_TCB", "SGX TCB Versions", "CPUID.12H:EAX.SGX_TCB[bit 2]", "intel", "", -1},
			3: {"SGX_OVERSUB", "SGX Oversubscription", "CPUID.12H:EAX.SGX_OVERSUB[bit 3]", "intel", "", -1},
			4: {"SGX_KSS", "SGX Key Separation and Sharing", "CPUID.12H:EAX.SGX_KSS[bit 4]", "intel", "", -1},
			5: {"SGX_ENCLV", "SGX ENCLV Leaves", "CPUID.12H:EAX.SGX_ENCLV[bit 5]", "intel", "", -1},
			6: {"SGX_ENCLS", "SGX ENCLS Leaves", "CPUID.12H:EAX.SGX_ENCLS[bit 6]", "intel", "", -1},
			7: {"SGX_ENCLU", "SGX ENCLU Leaves", "CPUID.12H:EAX.SGX_ENCLU[bit 7]", "intel", "", -1},
			// AMD equivalent features (SEV)
			8:  {"SEV", "Secure Encrypted Virtualization", "CPUID.8000001FH:EAX.SEV[bit 1]", "amd", "", -1},
			9:  {"SEV_ES", "SEV Encrypted State", "CPUID.8000001FH:EAX.SEV_ES[bit 2]", "amd", "", -1},
			10: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000001FH:EAX.SEV_SNP[bit 3]", "amd", "", -1},
		},
	}, "AdvancedMatrixExtensions": {
		name:     "Advanced Matrix Extensions",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Instruction",
		features: map[int]Feature{
			0: {"AMX_BF16", "AMX BFloat16 Support", "CPUID.7:EDX.AMX_BF16[bit 0]", "intel", "", -1},
			1: {"AMX_TILE", "AMX Tile Architecture", "CPUID.7:EDX.AMX_TILE[bit 1]", "intel", "", -1},
			2: {"AMX_INT8", "AMX Int8 Support", "CPUID.7:EDX.AMX_INT8[bit 2]", "intel", "", -1},
			3: {"AMX_FP16", "AMX FP16 Support", "CPUID.7:EDX.AMX_FP16[bit 3]", "intel", "", -1},
			4: {"AMX_DPBUUD", "AMX DPBUUD Support", "CPUID.7:EDX.AMX_DPBUUD[bit 4]", "intel", "", -1},
			5: {"AMX_DPBUUDS", "AMX DPBUUDS Support", "CPUID.7:EDX.AMX_DPBUUDS[bit 5]", "intel", "", -1},
			6: {"AMX_DPBUUD_TILELOAD", "AMX DPBUUD Tileload", "CPUID.7:EDX.AMX_DPBUUD_TILELOAD[bit 6]", "intel", "", -1},
			7: {"AMX_DPBUUDS_TILELOAD", "AMX DPBUUDS Tileload", "CPUID.7:EDX.AMX_DPBUUDS_TILELOAD[bit 7]", "intel", "", -1},
			// AMD Matrix extensions
			8: {"MAI", "Matrix Acceleration Instructions", "CPUID.8000001FH:EAX.MAI[bit 0]", "amd", "", -1},
			9: {"MAIA", "Matrix Acceleration Instructions Advanced", "CPUID.8000001FH:EAX.MAIA[bit 1]", "amd", "", -1},
		},
	}, "SMM": {
		name:     "System Management Mode",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "System Management",
		features: map[int]Feature{
			0: {"SMM", "System Management Mode", "CPUID.1:EDX.SMM[bit 0]", "common", "", -1},
			1: {"SMM_MONITOR", "SMM Monitor Extensions", "CPUID.1:ECX.SMM_MONITOR[bit 1]", "common", "", -1},
			2: {"SMM_VMCALL", "SMM VMCALL", "CPUID.1:ECX.SMM_VMCALL[bit 2]", "intel", "", -1},
			3: {"SMM_VMLOAD", "SMM VM Load", "CPUID.1:ECX.SMM_VMLOAD[bit 3]", "intel", "", -1},
			4: {"SMM_EXIT", "SMM VM Exit", "CPUID.1:ECX.SMM_EXIT[bit 4]", "intel", "", -1},
			5: {"SMM_ENTRY", "SMM VM Entry", "CPUID.1:ECX.SMM_ENTRY[bit 5]", "intel", "", -1},
			6: {"SMM_MSRPROT", "SMM MSR Protection", "CPUID.1:ECX.SMM_MSRPROT[bit 6]", "common", "", -1},
			7: {"SMM_TSEG", "SMM TSEG Memory", "CPUID.1:ECX.SMM_TSEG[bit 7]", "common", "", -1},
			// AMD specific SMM features
			8:  {"SMM_LOCK", "SMM Code Access Check", "CPUID.8000000AH:EDX.SMM_LOCK[bit 0]", "amd", "", -1},
			9:  {"SMM_ASID", "SMM Address Space ID", "CPUID.8000000AH:EDX.SMM_ASID[bit 1]", "amd", "", -1},
			10: {"SMM_SKINIT", "SKINIT and DEV support", "CPUID.8000000AH:EDX.SMM_SKINIT[bit 2]", "amd", "", -1},
		},
	}, "Real-TimeInstructions": {
		name:     "Real-Time Instructions",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Instruction",
		features: map[int]Feature{
			0: {"HRESET", "History Reset", "CPUID.7:ECX.HRESET[bit 0]", "common", "", -1},
			1: {"LAM", "Linear Address Masking", "CPUID.7:ECX.LAM[bit 1]", "common", "", -1},
			2: {"FRED", "Flexible Return and Event Delivery", "CPUID.7:ECX.FRED[bit 2]", "intel", "", -1},
			3: {"LKGS", "Load and Zero Segment Registers", "CPUID.7:ECX.LKGS[bit 3]", "intel", "", -1},
			4: {"WRMSRNS", "Write MSR No Serializing", "CPUID.7:ECX.WRMSRNS[bit 4]", "intel", "", -1},
			5: {"AMX_FP16", "AMX FP16 Instructions", "CPUID.7:ECX.AMX_FP16[bit 5]", "intel", "", -1},
			6: {"HRESET_OPT", "Optimized History Reset", "CPUID.7:ECX.HRESET_OPT[bit 6]", "common", "", -1},
			7: {"AVX_VNNI_INT16", "AVX VNNI 16-bit Integer", "CPUID.7:ECX.AVX_VNNI_INT16[bit 7]", "intel", "", -1},
			// AMD specific real-time features
			8:  {"MWAITX", "MONITORX/MWAITX Instructions", "CPUID.80000008H:EBX.MWAITX[bit 0]", "amd", "", -1},
			9:  {"MONITORX", "MONITORX Support", "CPUID.80000008H:EBX.MONITORX[bit 1]", "amd", "", -1},
			10: {"MSRLOCK", "MSR Lock Support", "CPUID.80000008H:EBX.MSRLOCK[bit 2]", "amd", "", -1},
		},
	}, "CoreThread": {
		name:     "Core & Thread",
		leaf:     1,
		subleaf:  0,
		register: 1,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"APIC_IDS", "Max APIC IDs per Package", "CPUID.1:EBX.APIC_IDS[bits 23-16]", "common", "", -1},
			1: {"CORE_COUNT", "Core Count Support", "CPUID.4:EAX.CORE_COUNT[bit 1]", "common", "", -1},
			2: {"THREAD_MASK", "Thread Mask Width", "CPUID.1:EAX.THREAD_MASK[bits 15-14]", "common", "", -1},
			3: {"CORE_MASK", "Core Mask Width", "CPUID.1:EAX.CORE_MASK[bits 13-12]", "common", "", -1},
			4: {"PKG_MASK", "Package Mask Width", "CPUID.1:EAX.PKG_MASK[bits 11-10]", "common", "", -1},
			5: {"HYBRID_ARCH", "Hybrid Architecture", "CPUID.7:EDX.HYBRID_ARCH[bit 15]", "intel", "", -1},
			6: {"NATIVE_MODEL_ID", "Native Model ID", "CPUID.1A:EAX.NATIVE_MODEL_ID[bits 31-24]", "common", "", -1},
			7: {"CORE_TYPE", "Core Type", "CPUID.1A:EAX.CORE_TYPE[bits 23-16]", "common", "", -1},
			// AMD specific thread features
			8:  {"COMPUTE_UNIT_ID", "Compute Unit ID", "CPUID.8000001E:EBX[bits 7-0]", "amd", "", -1},
			9:  {"NODES_PER_PROCESSOR", "Nodes per Processor", "CPUID.8000001E:ECX[bits 10-8]", "amd", "", -1},
			10: {"NODE_ID", "Node ID", "CPUID.8000001E:ECX[bits 7-0]", "amd", "", -1},
			11: {"THREADS_PER_CORE", "Threads per Core", "CPUID.8000001E:EBX[bits 15-8]", "amd", "", -1},
		},
	}, "ErrorDetection": {
		name:     "Error Detection",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "Error Handling",
		features: map[int]Feature{
			0: {"MCA", "Machine Check Architecture", "CPUID.1:EDX.MCA[bit 14]", "common", "", -1},
			1: {"MCE", "Machine Check Exception", "CPUID.1:EDX.MCE[bit 7]", "common", "", -1},
			2: {"DEP", "Data Execution Prevention", "CPUID.80000001:EDX.NX[bit 20]", "common", "", -1},
			3: {"MCDT", "Machine Check Data Table", "CPUID.1:ECX.MCDT[bit 21]", "common", "", -1},
			4: {"ERROR_COUNT", "Error-Reporting Counter Size", "CPUID.1:ECX.ERROR_COUNT[bits 23-22]", "common", "", -1},
			5: {"THRESHOLD", "Error-Reporting Threshold", "CPUID.1:ECX.THRESHOLD[bits 25-24]", "common", "", -1},
			6: {"OVERFLOW", "Error Counter Overflow", "CPUID.1:ECX.OVERFLOW[bit 26]", "common", "", -1},
			7: {"RECOVERY", "Error Recovery Support", "CPUID.1:ECX.RECOVERY[bit 27]", "common", "", -1},
			// AMD specific error features
			8:  {"MCA_OVERFLOW", "MCA Overflow Recovery", "CPUID.1:EDX.MCA_OVERFLOW[bit 30]", "amd", "", -1},
			9:  {"SUCCOR", "Software Uncorrectable Error Containment and Recovery", "CPUID.8000001FH:EAX[bit 5]", "amd", "", -1},
			10: {"HWA", "Hardware Assert Support", "CPUID.8000001FH:EAX[bit 6]", "amd", "", -1},
			11: {"SCALABLE_MCA", "Scalable MCA Support", "CPUID.8000001FH:EAX[bit 7]", "amd", "", -1},
		},
	}, "Prefetch": {
		name:     "Prefetch",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Prefetch & BUS",
		features: map[int]Feature{
			0: {"PREFETCHW", "PREFETCHW Instruction", "CPUID.1:ECX.PREFETCHW[bit 0]", "common", "", -1},
			1: {"PREFETCHWT1", "PREFETCHWT1 Instruction", "CPUID.7:ECX.PREFETCHWT1[bit 0]", "common", "", -1},
			2: {"PREFETCH_FSE", "Frequency Speculative Prefetch", "CPUID.7:EDX.PREFETCH_FSE[bit 20]", "intel", "", -1},
			3: {"PREFETCH_DISABLE", "Software Prefetch Disable", "CPUID.7:EDX.PREFETCH_DISABLE[bit 21]", "common", "", -1},
			4: {"PREFETCH_SAME_CACHE", "Same-Cache-Line Prefetch Disable", "CPUID.7:EDX.PREFETCH_SAME_CACHE[bit 22]", "common", "", -1},
			5: {"PREFETCH_DIFF_CACHE", "Different-Cache-Line Prefetch Disable", "CPUID.7:EDX.PREFETCH_DIFF_CACHE[bit 23]", "common", "", -1},
			6: {"PREFETCH_L1D", "L1D Prefetch Control", "CPUID.7:EDX.PREFETCH_L1D[bit 24]", "common", "", -1},
			7: {"PREFETCH_L2", "L2 Prefetch Control", "CPUID.7:EDX.PREFETCH_L2[bit 25]", "common", "", -1},
			// AMD specific prefetch features
			8: {"PREFETCH_3DNOW", "3DNow! Prefetch Instructions", "CPUID.80000001H:ECX.3DNOWPREFETCH[bit 8]", "amd", "", -1},
			9: {"DATA_PREFETCH", "Data Cache Prefetch", "CPUID.80000008H:EBX[bit 4]", "amd", "", -1},
		},
	}, "BUS": {
		name:     "BUS",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Prefetch & BUS",
		features: map[int]Feature{
			0: {"BUS_LOCK_DETECT", "Bus Lock Detection", "CPUID.7:ECX.BUS_LOCK_DETECT[bit 0]", "common", "", -1},
			1: {"BUS_LOCK_INTR", "Bus Lock Converted to Interrupt", "CPUID.7:ECX.BUS_LOCK_INTR[bit 1]", "common", "", -1},
			2: {"BUS_LOCK_MONITOR", "Bus Lock Monitor Support", "CPUID.7:ECX.BUS_LOCK_MONITOR[bit 2]", "common", "", -1},
			3: {"SPLIT_LOCK_DETECT", "Split Lock Detection", "CPUID.7:ECX.SPLIT_LOCK_DETECT[bit 3]", "common", "", -1},
			4: {"SPLIT_LOCK_INTR", "Split Lock Converted to Interrupt", "CPUID.7:ECX.SPLIT_LOCK_INTR[bit 4]", "common", "", -1},
			5: {"MLB", "Memory Bandwidth Licensing", "CPUID.7:ECX.MLB[bit 5]", "intel", "", -1},
			6: {"PPIN", "Protected Processor Inventory Number", "CPUID.7:ECX.PPIN[bit 6]", "intel", "", -1},
			7: {"QOS_ENF", "Quality of Service Enforcement", "CPUID.7:ECX.QOS_ENF[bit 7]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_CORE
			// AMD specific bus features
			8:  {"SSBD_VIRT", "Speculative Store Bypass Disable for Virtualization", "CPUID.80000008H:EBX[bit 24]", "amd", "", -1},
			9:  {"VIRT_SSBD", "Virtualized Speculative Store Bypass Disable", "CPUID.80000008H:EBX[bit 25]", "amd", "", -1},
			10: {"SSB_NO", "Speculative Store Bypass Disable Not Required", "CPUID.80000008H:EBX[bit 26]", "amd", "", -1},
		},
	}, "ApplicationTargeted": {
		name:     "Application Targeted",
		leaf:     7,
		subleaf:  0,
		register: 2,
		features: map[int]Feature{
			0: {"WAITPKG", "TPAUSE/UMONITOR/UMWAIT", "CPUID.7:ECX.WAITPKG[bit 5]", "intel", "AMDExtendedECX", 26}, // equivalent to AMD MWAITX
			1: {"KEYLOCKER", "Key Locker", "CPUID.7:ECX.KEYLOCKER[bit 23]", "intel", "", -1},
			2: {"MSR_LIST", "Restricted MSR Permission List", "CPUID.7:ECX.MSR_LIST[bit 24]", "common", "", -1},
			3: {"BUS_QOS", "Bus QoS Control", "CPUID.7:ECX.BUS_QOS[bit 25]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_CORE
			4: {"PBRSB", "Per-Branch Reset Shadow Bits", "CPUID.7:ECX.PBRSB[bit 26]", "intel", "", -1},
			5: {"INL", "Instruction Length Logger", "CPUID.7:ECX.INL[bit 27]", "intel", "", -1},
			6: {"HFI", "Hardware Feedback Interface", "CPUID.7:ECX.HFI[bit 28]", "intel", "", -1},
			7: {"FAST_SHORT_REP", "Fast Short REP MOVSB/STOSB", "CPUID.7:ECX.FAST_SHORT_REP[bit 29]", "common", "", -1},
			8: {"RING3MWAIT", "Ring 3 MWAIT/MONITOR", "CPUID.7:ECX.RING3MWAIT[bit 30]", "common", "", -1},
			9: {"PFSD", "Prefetch Side-Channel Disabled", "CPUID.7:ECX.PFSD[bit 31]", "common", "", -1},
			// AMD specific application features
			10: {"FSRC", "Fast Short REP CMPSB", "CPUID.7:EDX.FSRC[bit 1]", "amd", "", -1},
			11: {"FSRS", "Fast Short REP STOSB", "CPUID.7:EDX.FSRS[bit 2]", "amd", "", -1},
			12: {"FZRM", "Fast Zero-length REP MOVSB", "CPUID.7:EDX.FZRM[bit 3]", "amd", "", -1},
		},
	}, "ExtendedRegisterEAX": {
		name:     "Extended Register State",
		leaf:     0xD,
		subleaf:  0,
		register: 0,
		features: map[int]Feature{
			0: {"XFD", "Extended Feature Disable", "CPUID.D:EAX.XFD[bit 4]", "common", "", -1},
			1: {"XSAVES", "XSAVES/XRSTORS and IA32_XSS", "CPUID.D:EAX.XSAVES[bit 3]", "common", "", -1},
			2: {"XGETBV_ECX1", "XGETBV with ECX=1", "CPUID.D:EAX.XGETBV_ECX1[bit 2]", "common", "", -1},
			3: {"XSS", "Extended Supervisor State", "CPUID.D:EAX.XSS[bit 1]", "common", "", -1},
			4: {"XSAVEOPT", "XSAVEOPT Instruction", "CPUID.D:EAX.XSAVEOPT[bit 0]", "common", "", -1},
			5: {"PT_STATE", "Intel PT State", "CPUID.D:EAX.PT_STATE[bit 8]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			6: {"TILECFG", "AMX Tile Configuration", "CPUID.D:EAX.TILECFG[bit 17]", "intel", "", -1},
			7: {"TILEDATA", "AMX Tile Data", "CPUID.D:EAX.TILEDATA[bit 18]", "intel", "", -1},
			// AMD specific extended register features
			8:  {"MCOMMIT_STATE", "MCOMMIT State", "CPUID.D:EAX.MCOMMIT_STATE[bit 19]", "amd", "", -1},
			9:  {"CET_USER_STATE", "CET User State", "CPUID.D:EAX.CET_USER_STATE[bit 20]", "amd", "", -1},
			10: {"CET_SUPER_STATE", "CET Supervisor State", "CPUID.D:EAX.CET_SUPER_STATE[bit 21]", "amd", "", -1},
		},
	}, "HWFeedbackEDX": {
		name:     "Hardware Feedback Interface",
		leaf:     7,
		subleaf:  0,
		register: 3,
		features: map[int]Feature{
			0: {"HFI_PERF", "Performance Capabilities", "CPUID.7:EDX.HFI_PERF[bit 0]", "intel", "AMDExtendedECX", 14}, // equivalent to AMD LWP
			1: {"HFI_ENERGY", "Energy Capabilities", "CPUID.7:EDX.HFI_ENERGY[bit 1]", "intel", "", -1},
			2: {"HFI_CACHE", "Cache Allocation", "CPUID.7:EDX.HFI_CACHE[bit 2]", "intel", "", -1},
			3: {"HFI_BW", "Memory Bandwidth", "CPUID.7:EDX.HFI_BW[bit 3]", "intel", "", -1},
			4: {"HFI_IPC", "Instructions Per Cycle", "CPUID.7:EDX.HFI_IPC[bit 4]", "intel", "", -1},
			5: {"HFI_THERM", "Thermal Feedback", "CPUID.7:EDX.HFI_THERM[bit 5]", "intel", "AMDExtendedECX", 13}, // equivalent to AMD WDT
			6: {"HFI_TIME", "Time Stamp Counter", "CPUID.7:EDX.HFI_TIME[bit 6]", "common", "", -1},
			7: {"HFI_CORE", "Core-specific Feedback", "CPUID.7:EDX.HFI_CORE[bit 7]", "intel", "", -1},
			// AMD specific feedback features
			8: {"PROC_FEEDBACK", "Processor Feedback Interface", "CPUID.80000007:EDX.PROC_FEEDBACK[bit 11]", "amd", "", -1},
			9: {"CPPC", "Collaborative Processor Performance Control", "CPUID.80000007:EDX.CPPC[bit 9]", "amd", "", -1},
		},
	}, "PlatformQOSEDX": {
		name:     "Platform QoS Monitoring",
		leaf:     0xF,
		subleaf:  0,
		register: 3,
		features: map[int]Feature{
			0: {"PQM", "Platform QoS Monitoring", "CPUID.F:EDX.PQM[bit 0]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_CORE
			1: {"PQE", "Platform QoS Enforcement", "CPUID.F:EDX.PQE[bit 1]", "intel", "", -1},
			2: {"L3_QOS", "L3 Cache QoS", "CPUID.F:EDX.L3_QOS[bit 2]", "intel", "", -1},
			3: {"L2_QOS", "L2 Cache QoS", "CPUID.F:EDX.L2_QOS[bit 3]", "intel", "", -1},
			4: {"MBA", "Memory Bandwidth Allocation", "CPUID.F:EDX.MBA[bit 4]", "intel", "", -1},
			5: {"CMT", "Cache Monitoring Technology", "CPUID.F:EDX.CMT[bit 5]", "intel", "", -1},
			6: {"CAT", "Cache Allocation Technology", "CPUID.F:EDX.CAT[bit 6]", "intel", "", -1},
			7: {"MBA_MAX", "Maximum Memory Bandwidth Allocation", "CPUID.F:EDX.MBA_MAX[bit 7]", "intel", "", -1},
			// AMD specific QOS features
			8: {"NB_PERF", "Northbridge Performance Counters", "CPUID.80000007:EDX.NB_PERF[bit 10]", "amd", "", -1},
			9: {"L3_PERFCTR", "L3 Cache Performance Counter Extensions", "CPUID.80000007:EDX.L3_PERFCTR[bit 11]", "amd", "", -1},
		},
	}, "SharedCacheEAX": {
		name:     "Shared Cache",
		leaf:     4,
		subleaf:  0,
		register: 0,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"SHARED_L1", "Shared L1 Cache", "CPUID.4:EAX.SHARED_L1[bit 14]", "common", "", -1},
			1: {"SHARED_L2", "Shared L2 Cache", "CPUID.4:EAX.SHARED_L2[bit 15]", "common", "", -1},
			2: {"SHARED_L3", "Shared L3 Cache", "CPUID.4:EAX.SHARED_L3[bit 16]", "common", "", -1},
			3: {"CACHE_INCLUSIVE", "Cache Inclusiveness", "CPUID.4:EDX.CACHE_INCLUSIVE[bit 1]", "common", "", -1},
			4: {"COMPLEX_INDEX", "Complex Cache Indexing", "CPUID.4:EDX.COMPLEX_INDEX[bit 2]", "common", "", -1},
			5: {"CACHE_SHARING", "Cache Sharing Supported", "CPUID.4:EAX.CACHE_SHARING[bit 25]", "common", "", -1},
			6: {"CACHE_PARTS", "Cache Partitioning Supported", "CPUID.4:EAX.CACHE_PARTS[bit 26]", "intel", "", -1},
			7: {"CACHE_LEVEL", "Cache Level", "CPUID.4:EAX.CACHE_LEVEL[bits 7-5]", "common", "", -1},
			// AMD specific cache features
			8:  {"L3_NOT_USED", "L3 Cache Not Used", "CPUID.8000001D:EAX.L3_NOT_USED[bit 6]", "amd", "", -1},
			9:  {"COMPUTE_UNIT_ID", "Compute Unit ID", "CPUID.8000001D:EAX.COMPUTE_UNIT_ID[bits 15-8]", "amd", "", -1},
			10: {"NUM_SHARING", "Number of Cores Sharing Cache", "CPUID.8000001D:EAX.NUM_SHARING[bits 25-14]", "amd", "", -1},
		},
	}, "memoryTypeEDX": {
		name:     "Memory Type and Attribute",
		leaf:     1,
		subleaf:  0,
		register: 3,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"PAT", "Page Attribute Table", "CPUID.1:EDX.PAT[bit 16]", "common", "", -1},
			1: {"PSE", "Page Size Extension", "CPUID.1:EDX.PSE[bit 3]", "common", "", -1},
			2: {"PGE", "Page Global Enable", "CPUID.1:EDX.PGE[bit 13]", "common", "", -1},
			3: {"MTRR", "Memory Type Range Registers", "CPUID.1:EDX.MTRR[bit 12]", "common", "", -1},
			4: {"SMRR", "System Management Range Registers", "CPUID.8000_0006:EAX.SMRR[bit 0]", "common", "", -1},
			5: {"PAT_EXTENDED", "Extended Page Attribute Table", "CPUID.8000_0001:EDX.PAT_EXTENDED[bit 16]", "amd", "", -1},
			6: {"NX", "No-Execute Page Protection", "CPUID.8000_0001:EDX.NX[bit 20]", "amd", "", -1},
			7: {"1GB_PAGE", "1-GByte Page Support", "CPUID.8000_0001:EDX.1GB_PAGE[bit 26]", "amd", "", -1},
			// Intel specific memory type features
			8: {"EPT_1GB", "Extended Page Tables 1GB pages", "CPUID.80000001:EDX.EPT_1GB[bit 24]", "intel", "", -1},
			9: {"MPX_BNDREGS", "Memory Protection Extensions Bound Registers", "CPUID.7:EBX.MPX_BNDREGS[bit 3]", "intel", "", -1},
		},
	}, "TempPowerECX": {
		name:     "Temperature and Power",
		leaf:     0x80000007,
		subleaf:  0,
		register: 3,
		features: map[int]Feature{
			0:  {"ACNT2", "ACNT2 Feature", "CPUID.8000_0007:EDX.ACNT2[bit 1]", "amd", "", -1},
			1:  {"CPB", "Core Performance Boost", "CPUID.8000_0007:EDX.CPB[bit 9]", "amd", "PowerEAX", 12}, // equivalent to Intel TURBO3
			2:  {"DTSC", "Invariant TSC", "CPUID.8000_0007:EDX.DTSC[bit 8]", "common", "", -1},
			3:  {"HW_PSTATE", "Hardware P-state control", "CPUID.8000_0007:EDX.HW_PSTATE[bit 7]", "amd", "PowerEAX", 6}, // equivalent to Intel HWP
			4:  {"PROC_FEEDBACK", "Processor Feedback Interface", "CPUID.8000_0007:EDX.PROC_FEEDBACK[bit 11]", "amd", "", -1},
			5:  {"TM", "Thermal Monitor", "CPUID.1:EDX.TM[bit 29]", "common", "", -1},
			6:  {"TM2", "Thermal Monitor 2", "CPUID.1:ECX.TM2[bit 8]", "common", "", -1},
			7:  {"TTP", "Temperature Threshold", "CPUID.6:EAX.TTP[bit 14]", "common", "", -1},
			8:  {"HWP", "Hardware P-States", "CPUID.6:EAX.HWP[bit 7]", "intel", "AMDExtendedECX", 14}, // equivalent to AMD LWP
			9:  {"HWP_NOT", "HWP Notification", "CPUID.6:EAX.HWP_NOT[bit 8]", "intel", "", -1},
			10: {"HWP_ACT", "HWP Activity Window", "CPUID.6:EAX.HWP_ACT[bit 9]", "intel", "", -1},
			11: {"HWP_EPP", "HWP Energy Performance Preference", "CPUID.6:EAX.HWP_EPP[bit 10]", "intel", "", -1},
			12: {"HWP_PLR", "HWP Package Level Request", "CPUID.6:EAX.HWP_PLR[bit 11]", "intel", "", -1},
			13: {"HDC", "Hardware Duty Cycling", "CPUID.6:EAX.HDC[bit 13]", "intel", "", -1},
			14: {"ENERGY_BIAS", "Energy Bias Preference", "CPUID.6:ECX.ENERGY_BIAS[bit 3]", "intel", "", -1},
			15: {"RAPL", "Running Average Power Limit", "CPUID.6:ECX.RAPL[bit 1]", "common", "", -1},
			// Additional AMD specific power features
			16: {"CONNECTED_STANDBY", "Connected Standby", "CPUID.80000007:EDX.CONNECTED_STANDBY[bit 5]", "amd", "", -1},
			17: {"RAPL_INTERFACE", "RAPL Interface", "CPUID.80000007:EDX.RAPL[bit 14]", "amd", "", -1},
		},
	}, "SpecialInsEBX": {
		name:     "Special Instructions",
		leaf:     0x80000008,
		subleaf:  0,
		register: 1,
		group:    "Instruction",
		features: map[int]Feature{
			0:  {"CLZERO", "CLZERO Instruction", "CPUID.80000008:EBX.CLZERO[bit 0]", "amd", "", -1},
			1:  {"RDPRU", "RDPRU Instruction", "CPUID.80000008:EBX.RDPRU[bit 4]", "amd", "", -1},
			2:  {"MCOMMIT", "MCOMMIT Instruction", "CPUID.80000008:EBX.MCOMMIT[bit 8]", "amd", "", -1},
			3:  {"WBNOINVD", "WBNOINVD Instruction", "CPUID.80000008:EBX.WBNOINVD[bit 9]", "common", "", -1},
			4:  {"RDSEED_INST", "RDSEED Instruction", "CPUID.7:EBX.RDSEED[bit 18]", "common", "", -1},
			5:  {"ADCX_ADOX", "ADCX/ADOX Instructions", "CPUID.7:EBX.ADX[bit 19]", "common", "", -1},
			6:  {"CLDEMOTE_INST", "CLDEMOTE Instruction", "CPUID.7:ECX.CLDEMOTE[bit 25]", "common", "", -1},
			7:  {"RDPID_INST", "RDPID Instruction", "CPUID.7:ECX.RDPID[bit 22]", "common", "", -1},
			8:  {"SERIALIZE", "SERIALIZE Instruction", "CPUID.7:EDX.SERIALIZE[bit 14]", "common", "", -1},
			9:  {"TSX_LDTRK", "TSX Suspend Load Address Tracking", "CPUID.7:EDX.TSX_LDTRK[bit 16]", "intel", "", -1},
			10: {"UINTR", "User Interrupts", "CPUID.7:EDX.UINTR[bit 5]", "intel", "", -1},
			11: {"CLFSH_OPT", "Optimized Cache Flushing", "CPUID.7:EBX.CLFSH_OPT[bit 23]", "common", "", -1},
			12: {"FSRM", "Fast Short REP MOVSB", "CPUID.7:EDX.FSRM[bit 4]", "common", "", -1},
			13: {"FZRM", "Fast Zero-length REP MOVSB", "CPUID.7:EDX.FZRM[bit 3]", "common", "", -1},
			14: {"FSRS", "Fast Short REP STOSB", "CPUID.7:EDX.FSRS[bit 2]", "common", "", -1},
			15: {"FSRC", "Fast Short REP CMPSB/SCASB", "CPUID.7:EDX.FSRC[bit 1]", "common", "", -1},
			// Intel specific instructions
			16: {"HRESET", "History Reset", "CPUID.7:EAX.HRESET[bit 22]", "intel", "", -1},
			17: {"LAM", "Linear Address Masking", "CPUID.7:ECX.LAM[bit 26]", "intel", "", -1},
			// Additional AMD specific instructions
			18: {"MONITORX", "MONITORX Instruction", "CPUID.80000008:EBX.MONITORX[bit 2]", "amd", "", -1},
			19: {"MSRWRITE", "MSR Write All", "CPUID.80000008:EBX.MSRWRITE[bit 3]", "amd", "", -1},
			20: {"INVLPGB", "INVLPGB Instruction", "CPUID.80000008:EBX.INVLPGB[bit 5]", "amd", "", -1},
		},
	}, "SystemPlatformEDX": {
		name:     "System/Platform",
		leaf:     0x80000007,
		subleaf:  0,
		register: 3,
		group:    "Platform & Configuration",
		features: map[int]Feature{
			0: {"TIME_STAMP_DISABLE", "Time Stamp Counter Disable", "CPUID.80000007:EDX.TSD[bit 8]", "common", "", -1},
			1: {"FREQUENCY_ID_CTRL", "Frequency ID Control", "CPUID.80000007:EDX.FID[bit 1]", "amd", "PowerEAX", 1}, // equivalent to Intel IDA
			2: {"VOLTAGE_ID_CTRL", "Voltage ID Control", "CPUID.80000007:EDX.VID[bit 2]", "amd", "", -1},
			3: {"THERMTRIP", "Thermal Trip", "CPUID.80000007:EDX.TTP[bit 3]", "amd", "", -1},
			4: {"HARDWARE_FEEDBACK", "Hardware Feedback", "CPUID.80000007:EDX.HWF[bit 4]", "amd", "HWFeedbackEDX", 0}, // equivalent to Intel HFI_PERF
			5: {"PROC_POWER_REPORTING", "Processor Power Reporting", "CPUID.80000007:EDX.PPR[bit 5]", "amd", "", -1},
			6: {"CONNECTED_STANDBY", "Connected Standby", "CPUID.80000007:EDX.CS[bit 6]", "amd", "", -1},
			7: {"RAPL_INTERFACE", "Running Average Power Limiting", "CPUID.80000007:EDX.RAPL[bit 7]", "common", "", -1},
			// Intel specific platform features
			8:  {"PACKAGE_THERMAL", "Package Thermal Management", "CPUID.6:EAX.PTM[bit 6]", "intel", "", -1},
			9:  {"HWP_INTERRUPT", "HWP Interrupt", "CPUID.6:EAX.HWP_INT[bit 8]", "intel", "", -1},
			10: {"ENERGY_PERF_BIAS", "Energy Performance Bias", "CPUID.6:ECX.EPB[bit 3]", "intel", "", -1},
		},
	}, "Encryption": {
		name:     "Encryption",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Security",
		features: map[int]Feature{
			0:  {"AES_128", "AES 128-bit Support", "CPUID.7:ECX.AES_128[bit 0]", "common", "", -1},
			1:  {"AES_256", "AES 256-bit Support", "CPUID.7:ECX.AES_256[bit 1]", "common", "", -1},
			2:  {"VAES_128", "Vector AES 128-bit", "CPUID.7:ECX.VAES_128[bit 2]", "common", "", -1},
			3:  {"VAES_256", "Vector AES 256-bit", "CPUID.7:ECX.VAES_256[bit 3]", "common", "", -1},
			4:  {"SHA_1", "SHA-1 Instructions", "CPUID.7:EBX.SHA1[bit 29]", "common", "", -1},
			5:  {"SHA_256", "SHA-256 Instructions", "CPUID.7:EBX.SHA256[bit 30]", "common", "", -1},
			6:  {"SHA_512", "SHA-512 Instructions", "CPUID.7:ECX.SHA512[bit 0]", "intel", "", -1},
			7:  {"GFNI", "Galois Field Instructions", "CPUID.7:ECX.GFNI[bit 8]", "common", "", -1},
			8:  {"VAES", "Vector AES Instructions", "CPUID.7:ECX.VAES[bit 9]", "common", "", -1},
			9:  {"VPCLMULQDQ", "Vector PCLMULQDQ", "CPUID.7:ECX.VPCLMULQDQ[bit 10]", "common", "", -1},
			10: {"KL", "Key Locker Instructions", "CPUID.7:ECX.KL[bit 23]", "intel", "", -1},
			// AMD specific encryption features
			11: {"SEV", "Secure Encrypted Virtualization", "CPUID.8000001FH:EAX.SEV[bit 1]", "amd", "", -1},
			12: {"SEV_ES", "SEV Encrypted State", "CPUID.8000001FH:EAX.SEV_ES[bit 2]", "amd", "", -1},
			13: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000001FH:EAX.SEV_SNP[bit 3]", "amd", "", -1},
			14: {"AESKEY128", "128-bit AES Key Generation", "CPUID.8000001FH:EAX.AESKEY128[bit 4]", "amd", "", -1},
			15: {"AESKEY256", "256-bit AES Key Generation", "CPUID.8000001FH:EAX.AESKEY256[bit 5]", "amd", "", -1},
		},
	}, "ExtendedMemoryEBX": {
		name:     "Extended Memory",
		leaf:     0x80000008,
		subleaf:  0,
		register: 1,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"RDPRU", "Read Processor Register User", "CPUID.80000008:EBX.RDPRU[bit 4]", "amd", "", -1},
			1: {"INVLPGB", "Invalidate TLB IPI Broadcast", "CPUID.80000008:EBX.INVLPGB[bit 3]", "amd", "", -1},
			2: {"MCOMMIT", "Memory Commit Instruction", "CPUID.80000008:EBX.MCOMMIT[bit 8]", "amd", "", -1},
			3: {"TLBFLUSH", "Selective TLB Flush", "CPUID.80000008:EBX.TLB_FLUSH[bit 10]", "amd", "", -1},
			4: {"SSBD_VIRT_SPEC", "Speculative Store Bypass Disable", "CPUID.80000008:EBX.SSBD[bit 24]", "amd", "", -1},
			5: {"VIRT_SSBD", "Virtualized Speculative Store Bypass Disable", "CPUID.80000008:EBX.VIRT_SSBD[bit 25]", "amd", "", -1},
			6: {"IBS_FETCH_CTL_MSR", "IBS Fetch Control MSR", "CPUID.80000008:EBX.IBS_FETCH_CTL[bit 12]", "amd", "", -1},
			7: {"IBS_OP_CTL_MSR", "IBS Op Control MSR", "CPUID.80000008:EBX.IBS_OP_CTL[bit 13]", "amd", "", -1},
			// Intel specific memory features
			8:  {"PKU", "Memory Protection Keys", "CPUID.7:ECX.PKU[bit 3]", "intel", "", -1},
			9:  {"PKS", "Supervisor Protection Keys", "CPUID.7:ECX.PKS[bit 31]", "intel", "", -1},
			10: {"UMIP", "User-Mode Instruction Prevention", "CPUID.7:ECX.UMIP[bit 2]", "common", "", -1},
			11: {"LA57", "57-bit Linear Addresses", "CPUID.7:ECX.LA57[bit 16]", "common", "", -1},
			12: {"EPT", "Extended Page Tables", "CPUID.1:ECX.EPT[bit 6]", "intel", "", -1},
			13: {"1GB_PAGES", "1GB Pages Support", "CPUID.80000001:EDX.PAGE1GB[bit 26]", "common", "", -1},
		},
	}, "AdvancedPowerManagement": {
		name:     "Advanced Power Management",
		leaf:     0x80000007,
		subleaf:  0,
		register: 3,
		group:    "Power Management",
		features: map[int]Feature{
			0:  {"APM_CNT", "Advanced Power Management Counter", "CPUID.80000007:EDX.APM_CNT[bit 0]", "amd", "", -1},
			1:  {"APM_THR", "Advanced Power Management Threshold", "CPUID.80000007:EDX.APM_THR[bit 1]", "amd", "", -1},
			2:  {"APM_IRQ", "Advanced Power Management IRQ", "CPUID.80000007:EDX.APM_IRQ[bit 2]", "amd", "", -1},
			3:  {"APM_LOG", "Advanced Power Management Logging", "CPUID.80000007:EDX.APM_LOG[bit 3]", "amd", "", -1},
			4:  {"WDT", "Watchdog Timer", "CPUID.80000007:EDX.WDT[bit 4]", "amd", "", -1},
			5:  {"NB_TS", "North Bridge Temperature Sensor", "CPUID.80000007:EDX.NB_TS[bit 5]", "amd", "", -1},
			6:  {"100MHZ_STEPS", "100MHz Multiplier Steps", "CPUID.80000007:EDX.100MHZ_STEPS[bit 6]", "amd", "", -1},
			7:  {"HWPS", "Hardware P-State Control", "CPUID.80000007:EDX.HWPS[bit 7]", "amd", "PowerEAX", 6}, // equivalent to Intel HWP
			8:  {"TSC_INVARIANT", "Invariant Time Stamp Counter", "CPUID.80000007:EDX.TSC_INVARIANT[bit 8]", "common", "", -1},
			9:  {"CPB", "Core Performance Boost", "CPUID.80000007:EDX.CPB[bit 9]", "amd", "PowerEAX", 12}, // equivalent to Intel TURBO3
			10: {"EFFECTIVE_FREQ", "Effective Frequency Interface", "CPUID.80000007:EDX.EFFECTIVE_FREQ[bit 10]", "amd", "", -1},
			11: {"PROC_FEEDBACK", "Processor Feedback Interface", "CPUID.80000007:EDX.PROC_FEEDBACK[bit 11]", "amd", "HWFeedbackEDX", 0}, // equivalent to Intel HFI_PERF
			12: {"PROC_POWER_REPORTING", "Core Power Reporting", "CPUID.80000007:EDX.PROC_POWER_REPORTING[bit 12]", "amd", "", -1},
			// Intel equivalent features
			13: {"EPB", "Energy Performance Bias", "CPUID.6:ECX.EPB[bit 3]", "intel", "", -1},
			14: {"ENERGY_PERF_BIAS", "Energy Performance Bias MSR", "CPUID.6:ECX.ENERGY_PERF_BIAS[bit 4]", "intel", "", -1},
		},
	}, "CorePerformance": {
		name:     "Core Performance",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"HYBRID_CPU", "Hybrid CPU Support", "CPUID.7:EDX.HYBRID[bit 15]", "intel", "", -1},
			1: {"CORE_BOOST", "Core Performance Boost", "CPUID.80000007:EDX.CPB[bit 9]", "amd", "PowerEAX", 12}, // equivalent to Intel TURBO3
			2: {"RAPL_POWER_UNIT", "RAPL Power Unit", "CPUID.606:ECX.POWER_UNIT[bits 3-0]", "common", "", -1},
			3: {"CORE_PERF_BOOST", "Core Performance Boost Technology", "CPUID.80000007:EDX.CPB[bit 9]", "amd", "PowerEAX", 12},
			4: {"CORE_PERF_BOOST_LOCK", "Core Performance Boost Lock", "CPUID.80000007:EDX.CPBL[bit 10]", "amd", "", -1},
			5: {"CORE_PERF_VERSION", "Core Performance Version", "CPUID.80000007:EDX.CPBV[bit 11]", "amd", "", -1},
			6: {"CORE_PERF_BOOST_P0", "Core Performance Boost P0", "CPUID.80000007:EDX.CPBP0[bit 12]", "amd", "", -1},
			7: {"CORE_PERF_BOOST_P1", "Core Performance Boost P1", "CPUID.80000007:EDX.CPBP1[bit 13]", "amd", "", -1},
			// Intel specific performance features
			8:  {"HWP_CAPS", "Hardware P-State Capabilities", "CPUID.6:EAX.HWP_CAPS[bit 13]", "intel", "", -1},
			9:  {"HWP_NOTIFICATION", "HWP Notification", "CPUID.6:EAX.HWP_NOTIFICATION[bit 14]", "intel", "", -1},
			10: {"HWP_ACTIVITY", "HWP Activity Window", "CPUID.6:EAX.HWP_ACTIVITY[bit 15]", "intel", "", -1},
			11: {"HWP_ENERGY_PERF", "HWP Energy Performance Preference", "CPUID.6:EAX.HWP_ENERGY_PERF[bit 16]", "intel", "", -1},
			12: {"HWP_PACKAGE_REQ", "HWP Package Level Request", "CPUID.6:EAX.HWP_PACKAGE_REQ[bit 17]", "intel", "", -1},
		},
	}, "TransactionalSynchronizationExtensions": {
		name:     "Transactional Synchronization Extensions",
		leaf:     7,
		subleaf:  0,
		register: 3,
		features: map[int]Feature{
			0: {"TSX_HLE", "Hardware Lock Elision", "CPUID.7:EBX.HLE[bit 4]", "intel", "", -1},
			1: {"TSX_RTM", "Restricted Transactional Memory", "CPUID.7:EBX.RTM[bit 11]", "intel", "AMDExtendedECX", 11}, // equivalent to AMD XOP
			2: {"TSX_FORCE_ABORT", "TSX Force Abort MSR", "CPUID.7:EDX.TSX_FORCE_ABORT[bit 13]", "intel", "", -1},
			3: {"TSX_CTRL", "TSX Control MSR", "CPUID.7:EDX.TSX_CTRL[bit 14]", "intel", "", -1},
			4: {"RTM_ALWAYS_ABORT", "RTM Always Abort Mode", "CPUID.7:EDX.RTM_ALWAYS_ABORT[bit 22]", "intel", "", -1},
			5: {"TSX_CPUID_CLEAR", "TSX CPUID Clear", "CPUID.7:EDX.TSX_CPUID_CLEAR[bit 23]", "intel", "", -1},
			6: {"TSX_MEM_CLEAR", "TSX Memory Clear", "CPUID.7:EDX.TSX_MEM_CLEAR[bit 24]", "intel", "", -1},
			7: {"TSX_ALLOW_RTM", "RTM Execution Allowed", "CPUID.7:EDX.TSX_ALLOW_RTM[bit 25]", "intel", "", -1},
			// AMD alternatives
			8: {"XOP", "Extended Operations", "CPUID.80000001:ECX.XOP[bit 11]", "amd", "", -1},
			9: {"TBM", "Trailing Bit Manipulation", "CPUID.80000001:ECX.TBM[bit 21]", "amd", "", -1},
		},
	}, "User-Mode": {
		name:     "User-Mode",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Security",
		features: map[int]Feature{
			0: {"FSGS_BASE", "FSGSBASE Instructions", "CPUID.7:EBX.FSGSBASE[bit 0]", "common", "", -1},
			1: {"UMIP", "User-Mode Instruction Prevention", "CPUID.7:ECX.UMIP[bit 2]", "common", "", -1},
			2: {"UINTR", "User Interrupts", "CPUID.7:ECX.UINTR[bit 22]", "intel", "", -1},
			3: {"USERSPACE_EXEC", "User-Space Execute Prevention", "CPUID.7:ECX.USERSPACE_EXEC[bit 17]", "common", "", -1},
			4: {"USER_MSR", "User Mode MSR Access", "CPUID.7:ECX.USER_MSR[bit 18]", "common", "", -1},
			5: {"PKU", "Protection Keys for User-Mode Pages", "CPUID.7:ECX.PKU[bit 3]", "intel", "", -1},
			6: {"PKS", "Protection Keys for Supervisor-Mode Pages", "CPUID.7:ECX.PKS[bit 31]", "intel", "", -1},
			7: {"RDPRU", "Read Processor Register at User Level", "CPUID.80000008:EBX.RDPRU[bit 4]", "amd", "", -1},
			// AMD specific user mode features
			8:  {"UMPL", "User Mode Privilege Level", "CPUID.80000001:ECX.UMPL[bit 2]", "amd", "", -1},
			9:  {"USER_MEM_ENCRYPT", "User Mode Memory Encryption", "CPUID.8000001F:EAX[bit 2]", "amd", "", -1},
			10: {"USER_MEM_ENCRYPT_TRAP", "User Mode Memory Encryption Trap", "CPUID.8000001F:EAX[bit 3]", "amd", "", -1},
		},
	}, "NestedVirtualization": {
		name:     "Nested Virtualization",
		leaf:     0x8000000A,
		subleaf:  0,
		register: 3,
		group:    "Virtualization",
		features: map[int]Feature{
			0: {"NPT", "Nested Page Tables", "CPUID.8000000A:EDX.NPT[bit 0]", "amd", "VirtualizationECX", 1}, // equivalent to Intel EPT
			1: {"NRIPS", "Nested Reduced RIP Save", "CPUID.8000000A:EDX.NRIPS[bit 3]", "amd", "", -1},
			2: {"VMCB_CLEAN", "VMCB Clean Bits", "CPUID.8000000A:EDX.VMCB_CLEAN[bit 4]", "amd", "", -1},
			3: {"NESTED_FLUSH", "Nested Flush By ASID", "CPUID.8000000A:EDX.FLUSH_BY_ASID[bit 6]", "amd", "", -1},
			4: {"NESTED_DEC_TRAP", "Nested Decode Assists", "CPUID.8000000A:EDX.DEC_TRAP[bit 7]", "amd", "", -1},
			5: {"PAUSE_FILTER", "Pause Intercept Filter", "CPUID.8000000A:EDX.PAUSE_FILTER[bit 10]", "amd", "", -1},
			6: {"NESTED_PAUSE_FILTER", "Nested Pause Filter Threshold", "CPUID.8000000A:EDX.PAUSE_FILTER_THRESHOLD[bit 12]", "amd", "", -1},
			7: {"VGIF", "Virtual Global Interrupt Flag", "CPUID.8000000A:EDX.VGIF[bit 16]", "amd", "", -1},
			// Intel equivalents/alternatives
			8:  {"VMX_EPT", "Extended Page Tables", "CPUID.1:ECX.EPT[bit 1]", "intel", "", -1},
			9:  {"VMFUNC", "VM Functions", "CPUID.1:ECX.VMFUNC[bit 13]", "intel", "", -1},
			10: {"VMPTRLD_VMPTRST", "VM Pointer Load/Store", "CPUID.1:ECX.VMPTRLD[bit 14]", "intel", "", -1},
			11: {"VMWRITE_VMREAD", "VM Write/Read", "CPUID.1:ECX.VMWRITE[bit 15]", "intel", "", -1},
			12: {"SHADOW_VMCS", "Shadow VMCS", "CPUID.1:ECX.SHADOW_VMCS[bit 14]", "intel", "", -1},
		},
	}, "MemoryBandwidth": {
		name:     "Memory Bandwidth",
		leaf:     0x10,
		subleaf:  0,
		register: 2,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"INTEL_RDT_M", "Intel RDT Monitoring", "CPUID.7:EBX.RDT_M[bit 12]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_CORE
			1: {"INTEL_RDT_A", "Intel RDT Allocation", "CPUID.7:EBX.RDT_A[bit 15]", "intel", "", -1},
			2: {"MBA", "Memory Bandwidth Allocation", "CPUID.10:ECX.MBA[bit 3]", "intel", "", -1},
			3: {"CQM", "Cache QoS Monitoring", "CPUID.F:EDX.CQM[bit 1]", "intel", "", -1},
			4: {"MBM_TOTAL", "Memory Bandwidth Monitoring Total", "CPUID.F:EDX.MBM_TOTAL[bit 2]", "intel", "", -1},
			5: {"MBCQA", "Memory Bandwidth Allocation Control", "CPUID.10:ECX.MBCQA[bit 4]", "intel", "", -1},
			6: {"L3_MONITORING", "L3 Cache Monitoring", "CPUID.F:EDX.L3_MONITORING[bit 0]", "intel", "", -1},
			7: {"L3_ALLOCATION", "L3 Cache Allocation", "CPUID.10:ECX.L3_ALLOCATION[bit 2]", "intel", "", -1},
			// AMD specific bandwidth monitoring
			8:  {"NB_PMC", "North Bridge Performance Monitor", "CPUID.80000007:EDX.NB_PMC[bit 9]", "amd", "", -1},
			9:  {"DF_PMC", "Data Fabric Performance Monitor", "CPUID.80000007:EDX.DF_PMC[bit 10]", "amd", "", -1},
			10: {"DRAM_PMC", "DRAM Performance Monitor", "CPUID.80000007:EDX.DRAM_PMC[bit 11]", "amd", "", -1},
			11: {"BW_THROTTLING", "Memory Bandwidth Throttling", "CPUID.80000008:EBX.BW_THROTTLING[bit 14]", "amd", "", -1},
		},
	}, "SpeculationControl": {
		name:     "Speculation Control",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Security",
		features: map[int]Feature{
			0: {"IBRS", "Indirect Branch Restricted Speculation", "CPUID.7:EDX.IBRS[bit 0]", "common", "", -1},
			1: {"STIBP", "Single Thread Indirect Branch Predictors", "CPUID.7:EDX.STIBP[bit 1]", "common", "", -1},
			2: {"SSBD", "Speculative Store Bypass Disable", "CPUID.7:EDX.SSBD[bit 2]", "common", "", -1},
			3: {"IBPB", "Indirect Branch Predictor Barrier", "CPUID.7:EDX.IBPB[bit 3]", "common", "", -1},
			4: {"L1D_FLUSH", "L1 Data Cache Flush", "CPUID.7:EDX.L1D_FLUSH[bit 4]", "common", "", -1},
			5: {"MD_CLEAR", "Machine Check Data Clear", "CPUID.7:EDX.MD_CLEAR[bit 5]", "common", "", -1},
			6: {"SRBDS_CTRL", "Special Register Buffer Data Sampling Control", "CPUID.7:EDX.SRBDS_CTRL[bit 9]", "intel", "", -1},
			7: {"FB_CLEAR", "Fill Buffer Clear", "CPUID.7:EDX.FB_CLEAR[bit 17]", "common", "", -1},
			// AMD specific speculation controls
			8:  {"VIRT_SSBD", "Virtualized Speculative Store Bypass Disable", "CPUID.80000008:EBX.VIRT_SSBD[bit 24]", "amd", "", -1},
			9:  {"SSB_NO", "Speculative Store Bypass Not Required", "CPUID.80000008:EBX.SSB_NO[bit 25]", "amd", "", -1},
			10: {"PSFD", "Predictive Store Forward Disable", "CPUID.80000008:EBX.PSFD[bit 26]", "amd", "", -1},
			11: {"SPEC_CTRL_NO", "SPEC_CTRL Not Required", "CPUID.80000008:EBX.SPEC_CTRL_NO[bit 27]", "amd", "", -1},
		},
	}, "BranchPrediction": {
		name:     "Branch Prediction",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"IBC", "Indirect Branch Control", "CPUID.7:EDX.IBC[bit 0]", "common", "", -1},
			1: {"IBPB_BRTYPE", "Indirect Branch Type Control", "CPUID.7:EDX.IBPB_BRTYPE[bit 1]", "common", "", -1},
			2: {"SRSO", "Special Register Stack Overflow", "CPUID.7:EDX.SRSO[bit 2]", "common", "", -1},
			3: {"RRSBA_CTRL", "Return Stack Buffer Advance Control", "CPUID.7:EDX.RRSBA_CTRL[bit 3]", "intel", "", -1},
			4: {"RRSBA_SIZE", "Return Stack Buffer Size", "CPUID.7:EDX.RRSBA_SIZE[bit 4]", "intel", "", -1},
			5: {"BHI_CTRL", "Branch History Control", "CPUID.7:EDX.BHI_CTRL[bit 5]", "common", "", -1},
			6: {"PACKAGE_BHI_CTRL", "Package Branch History Control", "CPUID.7:EDX.PACKAGE_BHI_CTRL[bit 6]", "intel", "", -1},
			7: {"BHB_CLEAR", "Branch History Buffer Clear", "CPUID.7:EDX.BHB_CLEAR[bit 7]", "common", "", -1},
			// AMD specific branch prediction features
			8:  {"BP_IBPB", "Indirect Branch Prediction Barrier", "CPUID.80000008:EBX.BP_IBPB[bit 12]", "amd", "", -1},
			9:  {"BP_IBRS", "Indirect Branch Restricted Speculation", "CPUID.80000008:EBX.BP_IBRS[bit 13]", "amd", "", -1},
			10: {"BP_STIBP", "Single Thread Indirect Branch Predictor", "CPUID.80000008:EBX.BP_STIBP[bit 14]", "amd", "", -1},
			11: {"BP_RSB", "Return Stack Buffer Control", "CPUID.80000008:EBX.BP_RSB[bit 15]", "amd", "", -1},
		},
	}, "ExtendedTopology": {
		name:     "Extended Topology",
		leaf:     1,
		subleaf:  0,
		register: 3,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"EXTENDED_TOPOLOGY", "Extended Topology Enumeration", "CPUID.B:EAX.EXTENDED_TOPOLOGY[bit 0]", "common", "", -1},
			1: {"CORE_TYPE", "Core Type", "CPUID.B:ECX.CORE_TYPE[bits 15-8]", "common", "", -1},
			2: {"CORE_ID", "Core ID", "CPUID.B:EDX.CORE_ID[bits 31-0]", "common", "", -1},
			3: {"THREAD_MASK_WIDTH", "Thread Mask Width", "CPUID.B:EAX.THREAD_MASK_WIDTH[bits 4-0]", "common", "", -1},
			4: {"CORE_MASK_WIDTH", "Core Mask Width", "CPUID.B:EAX.CORE_MASK_WIDTH[bits 12-8]", "common", "", -1},
			5: {"PACKAGE_MASK_WIDTH", "Package Mask Width", "CPUID.B:EAX.PACKAGE_MASK_WIDTH[bits 20-16]", "common", "", -1},
			6: {"LEVEL_NUMBER", "Level Number", "CPUID.B:ECX.LEVEL_NUMBER[bits 7-0]", "common", "", -1},
			7: {"LEVEL_TYPE", "Level Type", "CPUID.B:ECX.LEVEL_TYPE[bits 15-8]", "common", "", -1},
			// AMD specific topology features
			8:  {"COMPUTE_UNIT_ID", "Compute Unit Identifier", "CPUID.8000001E:EBX[bits 7-0]", "amd", "", -1},
			9:  {"CORES_PER_COMPUTE_UNIT", "Cores per Compute Unit", "CPUID.8000001E:EBX[bits 15-8]", "amd", "", -1},
			10: {"NODE_ID", "Node Identifier", "CPUID.8000001E:ECX[bits 7-0]", "amd", "", -1},
			11: {"NODES_PER_PROCESSOR", "Nodes per Processor", "CPUID.8000001E:ECX[bits 10-8]", "amd", "", -1},
		},
	}, "Vector Neural Network": {
		name:     "Vector Neural Network",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Instruction",
		features: map[int]Feature{
			0: {"AVX512_4VNNIW", "AVX512 Vector Neural Network Instructions Word variable precision", "CPUID.7:EDX.AVX512_4VNNIW[bit 2]", "intel", "", -1},
			1: {"AVX512_4FMAPS", "AVX512 Multiply Accumulation Single precision", "CPUID.7:EDX.AVX512_4FMAPS[bit 3]", "intel", "", -1},
			2: {"AVX512_VP2INTERSECT", "AVX512 Vector Pair Intersection", "CPUID.7:EDX.AVX512_VP2INTERSECT[bit 8]", "intel", "", -1},
			3: {"AMX_BF16", "Advanced Matrix Extensions BF16", "CPUID.7:EDX.AMX_BF16[bit 22]", "intel", "", -1},
			4: {"AMX_TILE", "Advanced Matrix Extensions Tile", "CPUID.7:EDX.AMX_TILE[bit 24]", "intel", "", -1},
			5: {"AMX_INT8", "Advanced Matrix Extensions INT8", "CPUID.7:EDX.AMX_INT8[bit 25]", "intel", "", -1},
			6: {"AVX_VNNI", "AVX Vector Neural Network Instructions", "CPUID.7:ECX.AVX_VNNI[bit 4]", "intel", "", -1},
			7: {"AVX512_BF16", "AVX512 BFloat16 Instructions", "CPUID.7:EAX.AVX512_BF16[bit 5]", "intel", "", -1},
			// AMD vector/matrix features
			8:  {"MSRMASKING", "MSR Register Mask Control", "CPUID.80000021:EAX[bit 0]", "amd", "", -1},
			9:  {"VMPL", "Virtual Machine Priority Levels", "CPUID.8000001F:EAX[bit 4]", "amd", "", -1},
			10: {"VMSAV", "Virtual Machine Save Area", "CPUID.8000001F:EAX[bit 5]", "amd", "", -1},
			11: {"MAI", "Matrix Acceleration Instructions", "CPUID.80000021:EAX[bit 1]", "amd", "", -1},
		},
	}, "InstructExecution": {
		name:     "Instruction Execution",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Instruction",
		features: map[int]Feature{
			0: {"CLDEMOTE", "Cache Line Demote", "CPUID.7:ECX.CLDEMOTE[bit 25]", "common", "", -1},
			1: {"MOVDIRI", "Move Direct Store Integer", "CPUID.7:ECX.MOVDIRI[bit 27]", "intel", "", -1},
			2: {"MOVDIR64B", "Move 64 Bytes Direct Store", "CPUID.7:ECX.MOVDIR64B[bit 28]", "intel", "", -1},
			3: {"ENQCMD", "Enqueue Command", "CPUID.7:ECX.ENQCMD[bit 29]", "intel", "", -1},
			4: {"BUSLOCK_DETECT", "Bus Lock Detection", "CPUID.7:ECX.BUSLOCK_DETECT[bit 30]", "common", "", -1},
			5: {"DIRECT_STORE", "Direct Store Support", "CPUID.7:ECX.DIRECT_STORE[bit 31]", "common", "", -1},
			6: {"ZERO_FCS_FDS", "Zero FCS and FDS Support", "CPUID.7:EBX.ZERO_FCS_FDS[bit 13]", "common", "", -1},
			7: {"INSTR_RETIRED_CNT", "Instructions Retired Counter Support", "CPUID.7:EBX.INSTR_RETIRED_CNT[bit 14]", "common", "", -1},
			// AMD specific instruction execution features
			8:  {"CLZERO", "Clear Zero Instruction", "CPUID.80000008:EBX.CLZERO[bit 0]", "amd", "", -1},
			9:  {"MCOMMIT", "Memory Commit Instruction", "CPUID.80000008:EBX.MCOMMIT[bit 8]", "amd", "", -1},
			10: {"WBNOINVD", "Write Back No Invalidate", "CPUID.80000008:EBX.WBNOINVD[bit 9]", "amd", "", -1},
			11: {"IBPB", "Indirect Branch Prediction Barrier", "CPUID.80000008:EBX.IBPB[bit 12]", "amd", "", -1},
		},
	}, "PlatformQOSExtended": {
		name:     "Platform QoS Extended",
		leaf:     0x10,
		subleaf:  0,
		register: 1,
		features: map[int]Feature{
			0: {"PQE_L3_MASK", "L3 Cache QoS Enforcement Mask Length", "CPUID.10:EBX.PQE_L3_MASK[bits 4-0]", "intel", "", -1},
			1: {"PQE_L2_MASK", "L2 Cache QoS Enforcement Mask Length", "CPUID.10:EBX.PQE_L2_MASK[bits 12-8]", "intel", "", -1},
			2: {"PQE_RDT_MASK", "RDT QoS Enforcement Mask Length", "CPUID.10:EBX.PQE_RDT_MASK[bits 20-16]", "intel", "", -1},
			3: {"MBA_MAX_DELAY", "Maximum MBA Delay Value", "CPUID.10:EBX.MBA_MAX_DELAY[bits 31-24]", "intel", "", -1},
			4: {"L3_OCCUPANCY", "L3 Cache Occupancy Monitoring", "CPUID.F:EDX.L3_OCCUPANCY[bit 0]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_CORE
			5: {"L3_TOTAL_BW", "L3 Cache Total Bandwidth Monitoring", "CPUID.F:EDX.L3_TOTAL_BW[bit 1]", "intel", "", -1},
			6: {"L3_LOCAL_BW", "L3 Cache Local Bandwidth Monitoring", "CPUID.F:EDX.L3_LOCAL_BW[bit 2]", "intel", "", -1},
			7: {"MBA_NUM_DELAY", "Number of MBA Delay Values", "CPUID.F:EDX.MBA_NUM_DELAY[bits 15-8]", "intel", "", -1},
			// AMD specific QoS features
			8:  {"L3_QOS_EXTENSION", "L3 Cache QoS Extension", "CPUID.8000001D:EDX[bit 1]", "amd", "", -1},
			9:  {"DF_QOS_EXTENSION", "Data Fabric QoS Extension", "CPUID.8000001D:EDX[bit 2]", "amd", "", -1},
			10: {"BW_THROTTLING", "Bandwidth Throttling", "CPUID.8000001D:EDX[bit 3]", "amd", "", -1},
		},
	}, "ExtendedStateSaveArea": {
		name:     "Extended State Save Area",
		leaf:     0xD,
		subleaf:  0,
		register: 0,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"XSAVEOPT_SIZE", "XSAVEOPT Save Area Size", "CPUID.D:EAX.XSAVEOPT_SIZE[bits 31-0]", "common", "", -1},
			1: {"XSAVE_USER", "User State Components Size", "CPUID.D:EBX.XSAVE_USER[bits 31-0]", "common", "", -1},
			2: {"XSAVE_SUPERVISOR", "Supervisor State Components Size", "CPUID.D:ECX.XSAVE_SUPERVISOR[bits 31-0]", "common", "", -1},
			3: {"XCR0_MAX_BIT", "Maximum XCR0 Bit", "CPUID.D:EAX.XCR0_MAX_BIT[bits 7-0]", "common", "", -1},
			4: {"XSAVE_ALIGN", "XSAVE Area Alignment", "CPUID.D:EAX.XSAVE_ALIGN[bits 31-8]", "common", "", -1},
			5: {"XFD_FIP", "Extended Feature Disable FIP", "CPUID.D:ECX.XFD_FIP[bit 4]", "common", "", -1},
			6: {"XSAVES_COMPACT", "XSAVES Compaction Extensions", "CPUID.D:ECX.XSAVES_COMPACT[bit 1]", "common", "", -1},
			7: {"XSAVE_YMM_OPT", "Optimized XSAVE for YMM", "CPUID.D:ECX.XSAVE_YMM_OPT[bit 2]", "common", "", -1},
			// Intel specific state features
			8: {"PT_STATE", "Intel PT State", "CPUID.D:EAX.PT_STATE[bit 8]", "intel", "", -1},
			9: {"AMX_STATE", "AMX State Components", "CPUID.D:EAX.AMX_STATE[bit 17]", "intel", "", -1},
			// AMD specific state features
			10: {"MCOMMIT_STATE", "MCOMMIT State Components", "CPUID.D:EAX.MCOMMIT_STATE[bit 16]", "amd", "", -1},
			11: {"CET_STATE", "CET State Components", "CPUID.D:EAX.CET_STATE[bit 18]", "amd", "", -1},
		},
	}, "MemoryEncryption": {
		name:     "Memory Encryption",
		leaf:     0x19,
		subleaf:  0,
		register: 1,
		group:    "Cache & Memory",
		features: map[int]Feature{
			0: {"TME_CAPABILITY", "Total Memory Encryption Capability", "CPUID.19:EBX.TME_CAPABILITY[bit 0]", "intel", "AMDExtendedECX", 7}, // equivalent to AMD SME
			1: {"TME_ENCRYPT_TCP", "TME Encryption of TCACHE Possible", "CPUID.19:EBX.TME_ENCRYPT_TCP[bit 1]", "intel", "", -1},
			2: {"TME_NFX", "TME No False Xstore", "CPUID.19:EBX.TME_NFX[bit 2]", "intel", "", -1},
			3: {"MKTME", "Multi-Key Total Memory Encryption", "CPUID.7:ECX.MKTME[bit 13]", "intel", "", -1},
			4: {"KL", "Key Locker", "CPUID.7:ECX.KL[bit 23]", "intel", "", -1},
			5: {"AESKLE", "AES Key Locker Instructions", "CPUID.19:EBX.AESKLE[bit 0]", "intel", "", -1},
			6: {"WIDE_KL", "Wide Key Locker", "CPUID.19:EBX.WIDE_KL[bit 2]", "intel", "", -1},
			7: {"ENCRYPT_ALL", "All Memory Encryption Support", "CPUID.19:EBX.ENCRYPT_ALL[bit 3]", "intel", "", -1},
			// AMD specific memory encryption features
			8:  {"SME", "Secure Memory Encryption", "CPUID.8000001F:EAX[bit 0]", "amd", "", -1},
			9:  {"SEV", "Secure Encrypted Virtualization", "CPUID.8000001F:EAX[bit 1]", "amd", "", -1},
			10: {"PAGE_FLUSH", "Page Flush MSR", "CPUID.8000001F:EAX[bit 2]", "amd", "", -1},
			11: {"SEV_ES", "SEV Encrypted State", "CPUID.8000001F:EAX[bit 3]", "amd", "", -1},
			12: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000001F:EAX[bit 4]", "amd", "", -1},
			13: {"VMPL", "VM Permission Levels", "CPUID.8000001F:EAX[bit 5]", "amd", "", -1},
		},
	}, "CoreComplexTopology": {
		name:     "Core Complex Topology",
		leaf:     0x1F,
		subleaf:  0,
		register: 0,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"CORE_COMPLEX_ID", "Core Complex Identification", "CPUID.1F:EAX.CORE_COMPLEX_ID[bits 7-0]", "amd", "", -1},
			1: {"CORE_TYPE_ID", "Core Type Identification", "CPUID.1F:EAX.CORE_TYPE_ID[bits 15-8]", "common", "", -1},
			2: {"THREAD_MASK_WIDTH", "Thread Address Mask Width", "CPUID.1F:EAX.THREAD_MASK_WIDTH[bits 23-16]", "common", "", -1},
			3: {"CORE_MASK_WIDTH", "Core Address Mask Width", "CPUID.1F:EAX.CORE_MASK_WIDTH[bits 31-24]", "common", "", -1},
			4: {"PHY_BITS", "Physical Address Bits", "CPUID.1F:EBX.PHY_BITS[bits 7-0]", "common", "", -1},
			5: {"CORE_SELECT_MASK", "Core Selection Mask", "CPUID.1F:EBX.CORE_SELECT_MASK[bits 15-8]", "amd", "", -1},
			6: {"NODE_ID_MASK", "Node ID Mask", "CPUID.1F:EBX.NODE_ID_MASK[bits 23-16]", "amd", "", -1},
			7: {"CLUSTER_ID_MASK", "Cluster ID Mask", "CPUID.1F:EBX.CLUSTER_ID_MASK[bits 31-24]", "intel", "", -1},
			// Additional AMD topology features
			8:  {"CCX_ID", "CCX Identifier", "CPUID.8000001E:EBX[bits 7-0]", "amd", "", -1},
			9:  {"CORE_COUNT", "Core Count per CCX", "CPUID.8000001E:EBX[bits 15-8]", "amd", "", -1},
			10: {"SOCKET_ID", "Socket Identifier", "CPUID.8000001E:ECX[bits 7-0]", "amd", "", -1},
			// Intel specific topology features
			11: {"HYBRID_TYPE", "Hybrid Core Type", "CPUID.1A:EAX[bits 31-24]", "intel", "", -1},
			12: {"EFFICIENCY_CLASS", "Core Efficiency Class", "CPUID.1A:EAX[bits 23-16]", "intel", "", -1},
		},
	}, "PerformanceMonitoring": {
		name:     "Performance Monitoring",
		leaf:     0x0A,
		subleaf:  0,
		register: 0,
		group:    "Monitoring & Performance",
		features: map[int]Feature{
			0: {"PEBS_TRAP", "Precise Event Based Sampling Trap", "CPUID.A:EAX.PEBS_TRAP[bit 12]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			1: {"PB_ADV_FORMAT", "Performance Monitoring Advanced Format", "CPUID.A:EAX.PB_ADV_FORMAT[bit 13]", "intel", "", -1},
			2: {"PEBS_OUTPUT", "PEBS Output Format Support", "CPUID.A:EAX.PEBS_OUTPUT[bit 14]", "intel", "", -1},
			3: {"PEBS_BASELINE", "PEBS Baseline Events Support", "CPUID.A:EAX.PEBS_BASELINE[bit 15]", "intel", "", -1},
			4: {"LBR_FMT", "Last Branch Record Format", "CPUID.A:EAX.LBR_FMT[bits 23-16]", "common", "", -1},
			5: {"PEBS_RECORD", "PEBS Record Format Number", "CPUID.A:EAX.PEBS_RECORD[bits 31-24]", "intel", "", -1},
			6: {"PMC_WIDTH", "Performance Counter Width", "CPUID.A:EAX.PMC_WIDTH[bits 7-0]", "common", "", -1},
			7: {"FW_WRITE", "Performance Monitor FW Write Support", "CPUID.A:EDX.FW_WRITE[bit 13]", "common", "", -1},
			// AMD specific performance monitoring features
			8:  {"IBS_FETCH", "Instruction Based Sampling Fetch", "CPUID.8000001B:EAX[bit 0]", "amd", "", -1},
			9:  {"IBS_OP", "Instruction Based Sampling Op", "CPUID.8000001B:EAX[bit 1]", "amd", "", -1},
			10: {"IBS_BR_TGT", "IBS Branch Target Address", "CPUID.8000001B:EAX[bit 2]", "amd", "", -1},
			11: {"IBS_EXT_CNT", "IBS Extended Count", "CPUID.8000001B:EAX[bit 3]", "amd", "", -1},
			12: {"IBS_OPDATA4", "IBS Op Data 4", "CPUID.8000001B:EAX[bit 4]", "amd", "", -1},
			13: {"NPB", "Northbridge Performance Monitoring", "CPUID.8000001B:EAX[bit 5]", "amd", "", -1},
		},
	}, "HybridArchitecture": {
		name:     "Hybrid Architecture",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Core & Thread",
		features: map[int]Feature{
			0: {"HYBRID_CPU", "Hybrid Processor Identification", "CPUID.7:EDX.HYBRID_CPU[bit 15]", "intel", "", -1},
			1: {"NATIVE_MODEL_ID", "Native Model ID Support", "CPUID.7:EDX.NATIVE_MODEL_ID[bit 16]", "intel", "", -1},
			2: {"CORE_TYPE", "Core Type Support", "CPUID.7:EDX.CORE_TYPE[bit 17]", "intel", "", -1},
			3: {"ECORE_BITMAP", "Efficiency Core Bitmap", "CPUID.7:EDX.ECORE_BITMAP[bit 18]", "intel", "", -1},
			4: {"PCORE_BITMAP", "Performance Core Bitmap", "CPUID.7:EDX.PCORE_BITMAP[bit 19]", "intel", "", -1},
			5: {"CORE_POWER_SHARE", "Core Power Sharing Support", "CPUID.7:EDX.CORE_POWER_SHARE[bit 20]", "intel", "", -1},
			6: {"CORE_BOOST_SHARE", "Core Boost Sharing Support", "CPUID.7:EDX.CORE_BOOST_SHARE[bit 21]", "intel", "", -1},
			7: {"CORE_SELECT", "Core Selection Support", "CPUID.7:EDX.CORE_SELECT[bit 22]", "intel", "", -1},
			// AMD alternative features
			8:  {"CCX_CORE_TYPE", "CCX Core Type", "CPUID.8000001E:EBX[bits 23-16]", "amd", "", -1},
			9:  {"CCX_CORE_BOOST", "CCX Core Boost Control", "CPUID.8000001E:EBX[bits 31-24]", "amd", "", -1},
			10: {"COMPUTE_UNIT_ID", "Compute Unit ID", "CPUID.8000001E:EBX[bits 7-0]", "amd", "", -1},
		},
	}, "ArchitecturalLBR": {
		name:     "Architectural LBR",
		leaf:     0x1C,
		subleaf:  0,
		register: 0,
		group:    "Debug & Trace",
		features: map[int]Feature{
			0: {"CPL_FILTERING", "LBR CPL Filtering", "CPUID.1CH:EAX.CPL_FILTERING[bit 0]", "common", "", -1},
			1: {"BRANCH_FILTERING", "LBR Branch Filtering", "CPUID.1CH:EAX.BRANCH_FILTERING[bit 1]", "common", "", -1},
			2: {"CALL_STACK", "LBR Call Stack Mode", "CPUID.1CH:EAX.CALL_STACK[bit 2]", "common", "", -1},
			3: {"LBR_FREEZE", "LBR Freeze On PMI", "CPUID.1CH:EAX.LBR_FREEZE[bit 3]", "common", "", -1},
			4: {"IP_FILTERING", "LBR IP Filtering", "CPUID.1CH:EAX.IP_FILTERING[bit 4]", "common", "", -1},
			5: {"MISPREDICT", "LBR Misprediction Info", "CPUID.1CH:EAX.MISPREDICT[bit 5]", "common", "", -1},
			6: {"CYCLE_COUNT", "LBR Cycle Count", "CPUID.1CH:EAX.CYCLE_COUNT[bit 6]", "common", "", -1},
			7: {"BRANCH_TYPE", "LBR Branch Type Field", "CPUID.1CH:EAX.BRANCH_TYPE[bit 7]", "common", "", -1},
			// AMD specific LBR features
			8:  {"LBR_TOPOLOGY", "LBR Topology Support", "CPUID.8000001B:EAX[bit 6]", "amd", "", -1},
			9:  {"LBR_DEEP", "Deep LBR Stack", "CPUID.8000001B:EAX[bit 7]", "amd", "", -1},
			10: {"LBR_CONTEXT", "LBR Context Support", "CPUID.8000001B:EAX[bit 8]", "amd", "", -1},
		},
	}, "ProcessorEventBasedSampling": {
		name:     "Processor Event Based Sampling",
		leaf:     0x0A,
		subleaf:  0,
		register: 0,
		group:    "Monitoring & Performance",
		features: map[int]Feature{
			0: {"PEBS_TRAP_BIT", "PEBS Trap Bit", "CPUID.0AH:EAX.PEBS_TRAP[bit 12]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			1: {"PEBS_SAVE_ARCH", "PEBS Save Architectural State", "CPUID.0AH:EAX.PEBS_SAVE_ARCH[bit 13]", "intel", "", -1},
			2: {"PEBS_RECORD_FORMAT", "PEBS Record Format", "CPUID.0AH:EAX.PEBS_RECORD_FORMAT[bits 17-14]", "intel", "", -1},
			3: {"PEBS_LL_SUPPORT", "PEBS Load Latency Support", "CPUID.0AH:EAX.PEBS_LL[bit 18]", "intel", "", -1},
			4: {"PEBS_BASELINE", "PEBS Baseline Support", "CPUID.0AH:EAX.PEBS_BASELINE[bit 19]", "intel", "", -1},
			5: {"PEBS_PRECISE_IP", "PEBS Precise IP", "CPUID.0AH:EAX.PEBS_IP[bits 21-20]", "intel", "", -1},
			6: {"PEBS_PSB", "PEBS PSB Support", "CPUID.0AH:EAX.PEBS_PSB[bit 22]", "intel", "", -1},
			7: {"PEBS_GPR", "PEBS GPR Extension Support", "CPUID.0AH:EAX.PEBS_GPR[bit 23]", "intel", "", -1},
			// AMD specific sampling features
			8:  {"IBS_FETCH_SAMPLING", "IBS Fetch Sampling", "CPUID.8000001B:EAX[bit 0]", "amd", "", -1},
			9:  {"IBS_OP_SAMPLING", "IBS Op Sampling", "CPUID.8000001B:EAX[bit 1]", "amd", "", -1},
			10: {"IBS_RIP_INVALID", "IBS RIP Invalid Checking", "CPUID.8000001B:EAX[bit 2]", "amd", "", -1},
			11: {"IBS_BRANCH_TARGET", "IBS Branch Target Address", "CPUID.8000001B:EAX[bit 3]", "amd", "", -1},
			12: {"IBS_EXTENDED_CNT", "IBS Extended Counter Support", "CPUID.8000001B:EAX[bit 4]", "amd", "", -1},
		},
	}, "PlatformConfiguration": {
		name:     "Platform Configuration",
		leaf:     9,
		subleaf:  0,
		register: 0,
		group:    "Platform & Configuration",
		features: map[int]Feature{
			0: {"PLATFORM_DCA", "Platform DCA Capability", "CPUID.9:EAX.PLATFORM_DCA[bit 0]", "intel", "", -1},
			1: {"DCA_CAP_PREF", "DCA Capability Prefetch", "CPUID.9:EAX.DCA_CAP_PREF[bits 7-1]", "intel", "", -1},
			2: {"SOCKET_ID", "Socket ID Support", "CPUID.9:EAX.SOCKET_ID[bits 15-8]", "common", "", -1},
			3: {"CORE_ID", "Core ID Support", "CPUID.9:EAX.CORE_ID[bits 23-16]", "common", "", -1},
			4: {"THREAD_ID", "Thread ID Support", "CPUID.9:EAX.THREAD_ID[bits 31-24]", "common", "", -1},
			5: {"PCONFIG", "Platform Configuration", "CPUID.7:EDX.PCONFIG[bit 18]", "intel", "", -1},
			6: {"CET_SS_CFG", "CET Shadow Stack Config", "CPUID.7:ECX.CET_SS_CFG[bit 7]", "common", "", -1},
			7: {"CORE_CAPABILITY", "Core Capability Information", "CPUID.7:EDX.CORE_CAPABILITY[bit 29]", "common", "", -1},
			// AMD specific platform features
			8:  {"COMPUTE_UNIT_ID", "Compute Unit ID", "CPUID.8000001E:EBX[bits 7-0]", "amd", "", -1},
			9:  {"NODES_PER_PROC", "Nodes per Processor", "CPUID.8000001E:ECX[bits 10-8]", "amd", "", -1},
			10: {"NODE_ID", "Node ID", "CPUID.8000001E:ECX[bits 7-0]", "amd", "", -1},
			11: {"PLATFORM_QOS", "Platform QoS Configuration", "CPUID.8000001D:EDX[bit 1]", "amd", "", -1},
		},
	}, "EnhancedAddressTranslation": {
		name:     "Enhanced Address Translation",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Virtualization",
		features: map[int]Feature{
			0: {"GAP_WALK", "Guest Address Page Walk", "CPUID.7:EDX.GAP_WALK[bit 0]", "intel", "", -1},
			1: {"PAGE1G", "1GByte Page Support", "CPUID.7:EDX.PAGE1G[bit 1]", "common", "", -1},
			2: {"NESTED_EPT", "Nested EPT Support", "CPUID.7:EDX.NESTED_EPT[bit 2]", "intel", "AMDExtendedECX", 2}, // equivalent to AMD NPT
			3: {"DIRTY_ACCESS", "Dirty/Access Bit Support", "CPUID.7:EDX.DIRTY_ACCESS[bit 3]", "common", "", -1},
			4: {"EPT_EXEC_ONLY", "EPT Execute-Only Pages", "CPUID.7:EDX.EPT_EXEC_ONLY[bit 4]", "intel", "", -1},
			5: {"CR3_LOAD_NOEXIT", "CR3-Load Exiting", "CPUID.7:EDX.CR3_LOAD_NOEXIT[bit 5]", "intel", "", -1},
			6: {"SHADOW_STACK", "Shadow Stack Support", "CPUID.7:EDX.SHADOW_STACK[bit 6]", "common", "", -1},
			7: {"EPT_MODE_BASED", "EPT Mode-Based Execute Control", "CPUID.7:EDX.EPT_MODE_BASED[bit 7]", "intel", "", -1},
			// AMD specific translation features
			8:  {"NPT", "Nested Page Tables", "CPUID.8000000A:EDX.NPT[bit 0]", "amd", "", -1},
			9:  {"NPT_1G", "1GB Page Support for NPT", "CPUID.8000000A:EDX.NPT_1G[bit 1]", "amd", "", -1},
			10: {"NPT_2MB", "2MB Page Support for NPT", "CPUID.8000000A:EDX.NPT_2MB[bit 2]", "amd", "", -1},
			11: {"TLB_FLUSH_ASID", "TLB Flush by ASID", "CPUID.8000000A:EDX.TLB_FLUSH_ASID[bit 6]", "amd", "", -1},
			12: {"AVIC", "Advanced Virtual Interrupt Controller", "CPUID.8000000A:EDX.AVIC[bit 4]", "amd", "", -1},
		},
	}, "MCAExtended": {
		name:     "MCA Extended",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "Error Detection & Correction",
		features: map[int]Feature{
			0: {"MCA_OVERFLOW", "MCA Overflow Recovery", "CPUID.1:ECX.MCA_OVERFLOW[bit 0]", "common", "", -1},
			1: {"CMCI", "Corrected Machine Check Interrupt", "CPUID.1:ECX.CMCI[bit 1]", "common", "", -1},
			2: {"THRESHOLD", "Machine Check Threshold", "CPUID.1:ECX.THRESHOLD[bit 2]", "common", "", -1},
			3: {"MCG_CTL_P", "Machine Check Global Control", "CPUID.1:ECX.MCG_CTL_P[bit 3]", "common", "", -1},
			4: {"MCG_EXT_P", "Machine Check Extended Control", "CPUID.1:ECX.MCG_EXT_P[bit 4]", "common", "", -1},
			5: {"MCG_EXT_CTL_P", "Machine Check Extended Control P", "CPUID.1:ECX.MCG_EXT_CTL_P[bit 5]", "common", "", -1},
			6: {"MCG_SER_P", "Machine Check Serialization", "CPUID.1:ECX.MCG_SER_P[bit 6]", "common", "", -1},
			7: {"MCG_EXT_CNT", "Machine Check Extended Count", "CPUID.1:ECX.MCG_EXT_CNT[bits 15-8]", "common", "", -1},
			// AMD specific MCA features
			8:  {"MCA_THRESHOLD_CNT", "MCA Threshold Count", "CPUID.80000007:EBX.MCA_THRESHOLD_CNT[bits 15-8]", "amd", "", -1},
			9:  {"MCA_EXTENDED_FORMAT", "Extended Error Format", "CPUID.80000007:EBX.MCA_EXTENDED_FORMAT[bit 16]", "amd", "", -1},
			10: {"MCA_SCALABLE", "Scalable MCA", "CPUID.80000007:EBX.MCA_SCALABLE[bit 17]", "amd", "", -1},
			11: {"MCA_TLB_ERROR", "TLB Error Reporting", "CPUID.80000007:EBX.MCA_TLB_ERROR[bit 18]", "amd", "", -1},
		},
	}, "DebugInterface": {
		name:     "Debug Interface",
		leaf:     7,
		subleaf:  0,
		register: 2,
		group:    "Debug & Trace",
		features: map[int]Feature{
			0: {"LBR_DEPTH_ADJ", "LBR Stack Depth Adjust", "CPUID.7:ECX.LBR_DEPTH_ADJ[bit 0]", "common", "", -1},
			1: {"FREEZE_WHILE_SMM", "Debug Freeze While SMM", "CPUID.7:ECX.FREEZE_WHILE_SMM[bit 1]", "common", "", -1},
			2: {"RTM_DEBUG", "RTM Debug Support", "CPUID.7:ECX.RTM_DEBUG[bit 2]", "intel", "", -1},
			3: {"DEBUG_INTERFACE", "Debug Interface Support", "CPUID.7:ECX.DEBUG_INTERFACE[bit 3]", "common", "", -1},
			4: {"TRACE_RESTRICT", "Trace Message Restrict", "CPUID.7:ECX.TRACE_RESTRICT[bit 4]", "common", "", -1},
			5: {"PTWRITE", "PTWrite Support", "CPUID.7:ECX.PTWRITE[bit 5]", "intel", "", -1},
			6: {"IP_TRACESTOP", "IP TraceStop Support", "CPUID.7:ECX.IP_TRACESTOP[bit 6]", "intel", "", -1},
			7: {"LIP_EVENT", "LIP Event Trace Support", "CPUID.7:ECX.LIP_EVENT[bit 7]", "intel", "", -1},
			// AMD specific debug features
			8:  {"IBS_DEBUG", "IBS Debug Support", "CPUID.8000001B:EAX[bit 5]", "amd", "", -1},
			9:  {"DEBUG_VMEXT", "Debug VM Extensions", "CPUID.8000000A:EDX.DEBUG_VMEXT[bit 5]", "amd", "", -1},
			10: {"DEBUG_ENCRYPT", "Debug Encryption Support", "CPUID.8000001F:EAX[bit 6]", "amd", "", -1},
			11: {"DEBUG_BREAKPOINT", "Hardware Breakpoint Support", "CPUID.80000001:EDX.DEBUG_BREAKPOINT[bit 11]", "amd", "", -1},
		},
	}, "SMMExtended": {
		name:     "SMM Extended",
		leaf:     1,
		subleaf:  0,
		register: 2,
		group:    "System Management",
		features: map[int]Feature{
			0: {"SMM_MONITOR", "SMM Monitor Extensions", "CPUID.1:ECX.SMM_MONITOR[bit 0]", "common", "", -1},
			1: {"SMM_VMCALL", "SMM VMCALL", "CPUID.1:ECX.SMM_VMCALL[bit 1]", "intel", "", -1},
			2: {"SMM_DBG_CTL", "SMM Debug Controls", "CPUID.1:ECX.SMM_DBG_CTL[bit 2]", "intel", "", -1},
			3: {"SMM_IO_CTL", "SMM I/O Controls", "CPUID.1:ECX.SMM_IO_CTL[bit 3]", "intel", "", -1},
			4: {"SMM_MSEG", "SMM MSEG Base Support", "CPUID.1:ECX.SMM_MSEG[bit 4]", "common", "", -1},
			5: {"SMM_SEG_PROT", "SMM Segment Protection", "CPUID.1:ECX.SMM_SEG_PROT[bit 5]", "common", "", -1},
			6: {"SMM_TR_CTL", "SMM TR Controls", "CPUID.1:ECX.SMM_TR_CTL[bit 6]", "intel", "", -1},
			7: {"SMM_BLOCKED", "SMM Block Detection", "CPUID.1:ECX.SMM_BLOCKED[bit 7]", "common", "", -1},
			// AMD specific SMM features
			8:  {"TSEG_LOCK", "TSEG Lock Support", "CPUID.8000000A:EDX.TSEG_LOCK[bit 8]", "amd", "", -1},
			9:  {"SMM_LOCK", "SMM Code Access Check", "CPUID.8000000A:EDX.SMM_LOCK[bit 9]", "amd", "", -1},
			10: {"SKINIT", "SKINIT and DEV support", "CPUID.8000000A:EDX.SKINIT[bit 10]", "amd", "", -1},
			11: {"SMM_MSR_PROT", "SMM MSR Protection", "CPUID.8000000A:EDX.SMM_MSR_PROT[bit 11]", "amd", "", -1},
		},
	}, "MiscellaneousExtended": {
		name:     "Miscellaneous Extended",
		leaf:     7,
		subleaf:  0,
		register: 3,
		group:    "Miscellaneous",
		features: map[int]Feature{
			0:  {"WAITPKG_BURST", "WAIT Package Burst Support", "CPUID.7:EDX.WAITPKG_BURST[bit 0]", "intel", "", -1},
			1:  {"BUS_LOCK_FINE", "Bus Lock Fine Grained", "CPUID.7:EDX.BUS_LOCK_FINE[bit 1]", "common", "", -1},
			2:  {"FRMCTL", "Frame Marker Control", "CPUID.7:EDX.FRMCTL[bit 2]", "intel", "", -1},
			3:  {"SPRBPEI", "Supervisor Mode Prevention", "CPUID.7:EDX.SPRBPEI[bit 3]", "common", "", -1},
			4:  {"PBNDKB", "Bind Near Data Keys", "CPUID.7:EDX.PBNDKB[bit 4]", "intel", "", -1},
			5:  {"MCDT_NO", "Machine Check Data No", "CPUID.7:EDX.MCDT_NO[bit 5]", "common", "", -1},
			6:  {"IOMMU_VP", "IOMMU Virtual Processor", "CPUID.7:EDX.IOMMU_VP[bit 6]", "common", "", -1},
			7:  {"CET_SSS", "CET Supervisor Shadow Stacks", "CPUID.7:EDX.CET_SSS[bit 7]", "common", "", -1},
			8:  {"MD_CLEAR_CAP", "MD_CLEAR Capability", "CPUID.7:EDX.MD_CLEAR_CAP[bit 8]", "common", "", -1},
			9:  {"PSCHANGE_MC_NO", "Page Size Change MCE", "CPUID.7:EDX.PSCHANGE_MC_NO[bit 9]", "common", "", -1},
			10: {"TSXLDTRK", "TSX Suspend Load Address", "CPUID.7:EDX.TSXLDTRK[bit 10]", "intel", "", -1},
			11: {"IBC_NO", "Indirect Branch Control No", "CPUID.7:EDX.IBC_NO[bit 11]", "common", "", -1},
			12: {"PPIN_CTL", "Protected Processor Inventory Number Control", "CPUID.7:EDX.PPIN_CTL[bit 12]", "common", "", -1},
			13: {"CORE_MNGR", "Core Manager Support", "CPUID.7:EDX.CORE_MNGR[bit 13]", "intel", "", -1},
			14: {"PCID_PR", "Process Context ID Preserve", "CPUID.7:EDX.PCID_PR[bit 14]", "intel", "", -1},
			15: {"OVERCLK_CTL", "Overclocking Controls", "CPUID.7:EDX.OVERCLK_CTL[bit 15]", "common", "", -1},
			// AMD specific miscellaneous features
			16: {"OSVW", "OS Visible Workaround", "CPUID.80000001:ECX.OSVW[bit 9]", "amd", "", -1},
			17: {"IBS_OP_CNT", "IBS Op Counter", "CPUID.8000001B:EAX[bit 9]", "amd", "", -1},
			18: {"MWAITX", "MONITORX/MWAITX Support", "CPUID.80000001:ECX.MWAITX[bit 29]", "amd", "", -1},
			19: {"SEV_SNP", "SEV Secure Nested Paging", "CPUID.8000001F:EAX[bit 4]", "amd", "", -1},
		},
	}, "ArchitecturalPerformanceMonitoring": {
		name:     "Architectural Performance Monitoring",
		leaf:     0x0A, // CPUID.0AH
		subleaf:  0,
		register: 0, // Most features in EAX, some in EDX
		group:    "Performance Monitoring",
		features: map[int]Feature{
			0: {"PMC_VERSION", "Performance Monitor Version", "CPUID.0AH:EAX.VERSION[bits 7-0]", "common", "", -1},
			1: {"PERF_BIAS", "Performance Bias Hint", "CPUID.0AH:ECX.PERF_BIAS[bit 3]", "intel", "", -1},
			2: {"LBR_FMT", "Last Branch Record Format", "CPUID.0AH:EDX.LBR_FMT[bits 5-0]", "common", "", -1},
			3: {"PEBS_FMT", "Precise Event Based Sampling Format", "CPUID.0AH:EAX.PEBS_FMT[bits 11-8]", "intel", "AMDExtendedECX", 10}, // equivalent to AMD IBS
			4: {"PEBS_TRAP", "PEBS Trap", "CPUID.0AH:EAX.PEBS_TRAP[bit 12]", "intel", "AMDExtendedECX", 10},                            // equivalent to AMD IBS
			5: {"PEBS_SAVE", "PEBS Save Architectural State", "CPUID.0AH:EAX.PEBS_SAVE[bit 13]", "intel", "", -1},
			6: {"PERF_METRICS", "Performance Metrics Available", "CPUID.0AH:EDX.PERF_METRICS[bit 12]", "common", "", -1},
			7: {"LLC_PERF", "LLC Performance Monitoring", "CPUID.0AH:EDX.LLC_PERF[bit 13]", "intel", "AMDExtendedECX", 21}, // equivalent to AMD PERFCTR_NB
			// AMD specific performance monitoring features
			8:  {"IBS_FETCH", "IBS Fetch Sampling", "CPUID.8000001B:EAX[bit 0]", "amd", "", -1},
			9:  {"IBS_OP", "IBS Op Sampling", "CPUID.8000001B:EAX[bit 1]", "amd", "", -1},
			10: {"NB_PERF", "North Bridge Performance Counter", "CPUID.8000001B:EAX[bit 5]", "amd", "", -1},
			11: {"IBS_COUNT_EXT", "IBS Count Extensions", "CPUID.8000001B:EAX[bit 4]", "amd", "", -1},
		},
	}, "MemoryProtectionKey": {
		name:     "Memory Protection Key Features",
		leaf:     7, // Most features in CPUID.7
		subleaf:  0,
		register: 2, // ECX register for most features
		group:    "Memory Protection",
		features: map[int]Feature{
			0: {"PKS", "Protection Keys for Supervisor-Mode Pages", "CPUID.7:ECX.PKS[bit 31]", "intel", "", -1},
			1: {"PKU", "Protection Keys for User-Mode Pages", "CPUID.7:ECX.PKU[bit 3]", "intel", "", -1},
			2: {"PKE", "Protection Key Enable", "CPUID.7:ECX.PKE[bit 4]", "intel", "", -1},
			3: {"HDT", "Hardware Duty Cycling", "CPUID.6:EAX.HDT[bit 13]", "intel", "AMDExtendedECX", 13}, // equivalent to AMD WDT
			4: {"OSPKE", "OS Protection Keys Enable", "CPUID.7:ECX.OSPKE[bit 4]", "intel", "", -1},
			5: {"PBRSB", "Platform Bus Response Shadow Bits", "CPUID.7:ECX.PBRSB[bit 26]", "intel", "", -1},
			6: {"SPKKEYLOCK", "Supervisor Protection Key Key Lock", "CPUID.7:ECX.SPKKEYLOCK[bit 29]", "intel", "", -1},
			7: {"MAPKEY", "Memory Access Protection Keys", "CPUID.7:ECX.MAPKEY[bit 30]", "intel", "", -1},
			// AMD memory protection features
			8:  {"SEV_MEM_ENCRYPT", "Memory Encryption Support", "CPUID.8000001F:EAX[bit 1]", "amd", "", -1},
			9:  {"SEV_ES_MEM_ENCRYPT", "Encrypted State Memory", "CPUID.8000001F:EAX[bit 2]", "amd", "", -1},
			10: {"VM_PERM_LEVELS", "VM Permission Levels", "CPUID.8000001F:EAX[bit 4]", "amd", "", -1},
			11: {"VM_REG_PROT", "VM Register Protection", "CPUID.8000001F:EAX[bit 5]", "amd", "", -1},
		},
	},
}

func cpuid(eax, ecx uint32) (a, b, c, d uint32)

func init() {
	maxFunc, maxExtFunc = getMaxFunctions()
	vendorID = getVendorID()
	isAMD = strings.Contains(strings.ToUpper(vendorID), "AMD")
}

func printBasicInfo() {
	fmt.Printf("Running on %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Detecting CPU features for x86/x64...\n")
	fmt.Printf("CPUID Max Standard Function: %d\n", maxFunc)
	fmt.Printf("CPUID Max Extended Function: 0x%08x\n", maxExtFunc)
	fmt.Printf("CPU Vendor ID: %s\n", vendorID)

	isIntel := strings.Contains(strings.ToUpper(vendorID), "INTEL")

	// Get processor info
	a, b, _, _ := cpuid(1, 0)
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

	fmt.Printf("\nProcessor Details:\n")
	fmt.Printf("  Stepping ID: %d\n", steppingID)
	fmt.Printf("  Model: %d (0x%x)\n", effectiveModel, effectiveModel)
	fmt.Printf("  Family: %d (0x%x)\n", effectiveFamily, effectiveFamily)
	fmt.Printf("  Processor Type: %d\n", processorType)

	// Get brand string
	if maxExtFunc >= 0x80000004 {
		var brand [48]byte
		for i := 0; i < 3; i++ {
			a, b, c, d := cpuid(0x80000002+uint32(i), 0)
			copy(brand[i*16:], int32ToBytes(a))
			copy(brand[i*16+4:], int32ToBytes(b))
			copy(brand[i*16+8:], int32ToBytes(c))
			copy(brand[i*16+12:], int32ToBytes(d))
		}
		fmt.Printf("  Brand String: %s\n", strings.TrimSpace(string(brand[:])))
	}

	fmt.Printf("\nCache Information:\n")

	// AMD Cache Detection
	if isAMD && maxExtFunc >= 0x8000001D {
		for i := uint32(0); ; i++ {
			a, b, c, _ := cpuid(0x8000001D, i)
			cacheType := a & 0x1F
			if cacheType == 0 {
				break
			}
			level := (a >> 5) & 0x7
			lineSize := (b & 0xFFF) + 1
			partitions := ((b >> 12) & 0x3FF) + 1
			associativity := ((b >> 22) & 0x3FF) + 1
			sets := c + 1
			size := lineSize * partitions * associativity * sets

			typeString := getCacheTypeString(cacheType)

			fmt.Printf("  L%d %s Cache:\n", level, typeString)
			fmt.Printf("    Size: %d KB\n", size/1024)
			fmt.Printf("    Ways: %d\n", associativity)
			fmt.Printf("    Line Size: %d bytes\n", lineSize)
			fmt.Printf("    Total Sets: %d\n", sets)

			maxCoresSharing := ((a >> 14) & 0xFFF) + 1
			fmt.Printf("    Max Cores Sharing: %d\n", maxCoresSharing)

			printCacheFlags(a)
		}
		// Intel Cache Detection
	} else if isIntel && maxFunc >= 4 {
		for i := uint32(0); ; i++ {
			a, b, c, _ := cpuid(4, i)
			cacheType := a & 0x1F
			if cacheType == 0 {
				break
			}
			level := (a >> 5) & 0x7
			lineSize := (b & 0xFFF) + 1
			partitions := ((b >> 12) & 0x3FF) + 1
			associativity := ((b >> 22) & 0x3FF) + 1
			sets := c + 1
			size := lineSize * partitions * associativity * sets

			typeString := getCacheTypeString(cacheType)

			fmt.Printf("  L%d %s Cache:\n", level, typeString)
			fmt.Printf("    Size: %d KB\n", size/1024)
			fmt.Printf("    Ways: %d\n", associativity)
			fmt.Printf("    Line Size: %d bytes\n", lineSize)
			fmt.Printf("    Total Sets: %d\n", sets)

			maxCoresSharing := ((a >> 14) & 0xFFF) + 1
			fmt.Printf("    Max Cores Sharing: %d\n", maxCoresSharing)

			printCacheFlags(a)
		}
	}

	fmt.Printf("\nTLB Information:\n")
	if isAMD && maxExtFunc >= 0x80000005 {
		printAMDTLBInfo()
	} else if isIntel {
		printIntelTLBInfo()
	}

	fmt.Printf("\nBasic Features:\n")
	_, b, _, _ = cpuid(1, 0)
	fmt.Printf("  Max Logical Processors: %d\n", (b>>16)&0xFF)
	fmt.Printf("  Initial APIC ID: %d\n", (b>>24)&0xFF)

	// Physical address and linear address bits
	if maxExtFunc >= 0x80000008 {
		a, _, c, _ := cpuid(0x80000008, 0)
		fmt.Printf("  Physical Address Bits: %d\n", a&0xFF)
		fmt.Printf("  Linear Address Bits: %d\n", (a>>8)&0xFF)

		// AMD specific core count info
		if isAMD {
			coreCount := (c & 0xFF) + 1
			fmt.Printf("  Core Count: %d\n", coreCount)
			threadPerCore := ((c >> 8) & 0xFF) + 1
			fmt.Printf("  Threads per Core: %d\n", threadPerCore)
		}
	}

	// Additional Intel-specific information
	if isIntel && maxFunc >= 0x1A {
		a, _, _, _ := cpuid(0x1A, 0)
		if a&1 != 0 {
			fmt.Printf("\nHybrid Information:\n")
			fmt.Printf("  Hybrid CPU: Yes\n")
			fmt.Printf("  Native Model ID: %d\n", (a>>24)&0xFF)
			fmt.Printf("  Core Type: %d\n", (a>>16)&0xFF)
		}
	}

	fmt.Printf("\nFeature Information:\n")
	fmt.Printf("\nFeature Information:\n")
	_, _, c, d := cpuid(1, 0) // Get standard features
	fmt.Printf("\nStandard Features ECX:\n")
	printFeatureFlags(cpuFeaturesList["StandardECX"].features, c)
	fmt.Printf("\nStandard Features EDX:\n")
	printFeatureFlags(cpuFeaturesList["StandardEDX"].features, d)

	if maxFunc >= 7 {
		_, b, c, d := cpuid(7, 0) // Get extended features
		fmt.Printf("\nExtended Features EBX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedEBX"].features, b)
		fmt.Printf("\nExtended Features ECX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedECX"].features, c)
		fmt.Printf("\nExtended Features EDX:\n")
		printFeatureFlags(cpuFeaturesList["ExtendedEDX"].features, d)
	}

	if isAMD && maxExtFunc >= 0x80000001 {
		_, _, c, _ := cpuid(0x80000001, 0) // Get AMD extended features
		fmt.Printf("\nAMD Extended Features ECX:\n")
		printFeatureFlags(cpuFeaturesList["AMDExtendedECX"].features, c)
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

func printCacheInfo(level uint32, typeString string, size, associativity, lineSize, sets uint32) {
	fmt.Printf("  L%d %s Cache:\n", level, typeString)
	fmt.Printf("    Size: %d KB\n", size/1024)
	fmt.Printf("    Ways: %d\n", associativity)
	fmt.Printf("    Line Size: %d bytes\n", lineSize)
	fmt.Printf("    Total Sets: %d\n", sets)
}

func printCacheFlags(a uint32) {
	if (a>>9)&1 == 1 {
		fmt.Printf("    Fully Associative: Yes\n")
	}
	if (a>>10)&1 == 1 {
		fmt.Printf("    Write Back: Yes\n")
	}
	if (a>>11)&1 == 1 {
		fmt.Printf("    Inclusive: Yes\n")
	}
}

func printAMDTLBInfo() {
	a, b, _, _ := cpuid(0x80000005, 0)
	fmt.Printf("  L1 Data TLB:\n")
	fmt.Printf("    2MB/4MB Pages: %d entries\n", (a>>16)&0xFF)
	fmt.Printf("    4KB Pages: %d entries\n", (a>>24)&0xFF)
	fmt.Printf("  L1 Instruction TLB:\n")
	fmt.Printf("    2MB/4MB Pages: %d entries\n", (b>>16)&0xFF)
	fmt.Printf("    4KB Pages: %d entries\n", (b>>24)&0xFF)

	if maxExtFunc >= 0x80000006 {
		a, b, _, _ = cpuid(0x80000006, 0)
		fmt.Printf("  L2 Data TLB:\n")
		fmt.Printf("    2MB/4MB Pages: %d entries\n", (a>>16)&0xFFF)
		fmt.Printf("    4KB Pages: %d entries\n", (a>>28)&0xF)
		fmt.Printf("  L2 Instruction TLB:\n")
		fmt.Printf("    2MB/4MB Pages: %d entries\n", (b>>16)&0xFFF)
		fmt.Printf("    4KB Pages: %d entries\n", (b>>28)&0xF)
	}
}

func printIntelTLBInfo() {
	if maxFunc >= 0x2 {
		a, b, c, d := cpuid(0x2, 0)
		fmt.Printf("  Intel TLB information available through descriptor bytes\n")
		fmt.Printf("  Raw descriptor values: EAX=%08x EBX=%08x ECX=%08x EDX=%08x\n", a, b, c, d)
	}
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

func printProcessorInfo() {
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

	vendorID := getVendorID()
	isAMD = strings.Contains(strings.ToUpper(vendorID), "AMD")

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

// Helper function to get vendor ID
func getVendorID() string {
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

func main() {
	fmt.Printf("Running on %s/%s\n", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOARCH {
	case "amd64", "386":
		fmt.Println("Detecting CPU features for x86/x64...")
		printBasicInfo()
		printProcessorInfo()
	default:
		fmt.Println("Unsupported architecture.")
	}
}
