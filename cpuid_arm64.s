// +build linux darwin
// +build arm64

// cpuid_arm64.s
//
// This implementation ignores the input parameters.
// It reads the following system registers:
//   a: MIDR_EL1 (Main ID Register)
//   b: CTR_EL0 (Cache Type Register)
//   c: ID_AA64PFR0_EL1 (Processor Feature Register 0)
//   d: MPIDR_EL1 (Multiprocessor Affinity Register)
// Note: Although these registers are 64-bit, we only return their lower 32 bits.

TEXT ·cpuid(SB), $0-32
    // Read MIDR_EL1 into X0, then extract lower 32 bits.
    MRS     X0, MIDR_EL1
    UXTW    X0, X0         // Zero-extend lower 32 bits
    STRW    W0, a+8(FP)    // store into return variable 'a'

    // Read CTR_EL0
    MRS     X0, CTR_EL0
    UXTW    X0, X0
    STRW    W0, b+12(FP)   // store into 'b'

    // Read ID_AA64PFR0_EL1 (Processor Feature Register 0)
    MRS     X0, ID_AA64PFR0_EL1
    UXTW    X0, X0
    STRW    W0, c+16(FP)   // store into 'c'

    // Read MPIDR_EL1 (Multiprocessor Affinity Register)
    MRS     X0, MPIDR_EL1
    UXTW    X0, X0
    STRW    W0, d+20(FP)   // store into 'd'
    RET
