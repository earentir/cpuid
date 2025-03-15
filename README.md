# cpuid Package Documentation
The cpuid package provides a set of functions and data structures to query detailed information about the host CPU using the cpuid instruction. It enables you to identify the CPU vendor, supported features, caches, TLB configurations, and more on x86 and x86_64 architectures.

## Key Features
- Vendor Identification
    Detects whether the CPU is from Intel, AMD, or another vendor.
    Obtains the vendor ID string (e.g., GenuineIntel, AuthenticAMD).
- CPU Model and Family
    Retrieves raw and effective CPU family, model, stepping IDs, and processor type.
    Computes the effective model and family values by considering extended family/model information.

### Brand String
Extracts the full CPU brand string, the human-readable CPU name often shown in system specifications.

### Core and Thread Topology
Determines the number of cores and threads per core.
Supports both Intel and AMD topologies, including detection via extended CPUID leaves.

### Addressing Capabilities
Provides the number of physical and linear address bits, which can be useful for memory management and virtualization.

### Feature Detection
Checks for support of various CPU instruction set extensions and features (e.g., SSE4.2, AVX, AES).
Enumerates feature categories, known features, and which are currently supported on the host CPU.

### Cache Information
Retrieves details about each cache level (L1, L2, L3).
Reports cache type (data, instruction, unified), size, associativity, line size, sets, and sharing details.

### TLB (Translation Lookaside Buffer) Details
Provides TLB configuration and associativity for different page sizes and levels (L1, L2, L3).

### Intel Hybrid CPU Support
Detects Intel’s hybrid architecture (e.g., Performance and Efficient cores).
Identifies the core type (P-core or E-core) when running on hybrid CPUs.

## Important Functions

### CPU Information

#### GetVendorID() string
Returns the CPU vendor string (e.g., "GenuineIntel" or "AuthenticAMD").

#### GetMaxFunctions() (uint32, uint32)
Returns the maximum supported standard and extended CPUID function values. These are essential inputs for other queries.

#### GetProcessorInfo(maxFunc, maxExtFunc uint32) ProcessorInfo
Accepts the maximum standard and extended function values and returns a ProcessorInfo struct containing:

Family, Model, Stepping, Extended Family/Model
Brand String
Vendor ID
Core Count, Threads per Core
Addressing capabilities (Physical/Linear bits)
Max Supported Functions

## Feature Queries
### GetAllFeatureCategories() []string
Returns a list of all recognized feature categories.

### GetAllFeatureCategoriesDetailed() map[string][]map[string]string
Returns a detailed map of all categories, each containing a list of features with descriptions and vendor information.

### GetAllKnownFeatures(category string) []string
Lists all known features for a specified category.

### GetSupportedFeatures(category string) []string
Lists all supported features for a specified category on the current CPU.

### IsFeatureSupported(featureName string) bool
Checks if a specific feature (by name) is supported by the current CPU.

## Cache and TLB Information
### GetCacheInfo(maxFunc, maxExtFunc uint32, vendorID string) []CPUCacheInfo
Returns a slice of CPUCacheInfo structs describing each cache level’s properties.

### PrintCacheTable(caches []CPUCacheInfo)
Prints a formatted table of cache information for human-readable output.

### GetTLBInfo(maxFunc, maxExtFunc uint32, vendorID string) TLBInfo
Returns a TLBInfo struct containing TLB details (entries, associativity, page sizes) for L1, L2, and L3 levels.

### PrintTLBInfo(info TLBInfo)
Prints a human-readable overview of the TLB configuration.

## Intel Hybrid CPU
### GetIntelHybrid() IntelHybridInfo
Returns IntelHybridInfo about hybrid Intel CPUs. Indicates if the CPU is hybrid and identifies the core type (P-core or E-core).
