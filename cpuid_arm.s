// +build linux darwin
// +build arm

// cpuid_arm.s
//
// This implementation ignores the input parameters.
// It reads the following CP15 registers:
//   a: Main ID Register (MIDR)
//   b: Cache Type Register (CTR)
//   c: Processor Feature Register 0 (PFR0)
//   d: Multiprocessor Affinity Register (MPIDR)

TEXT ·cpuid(SB), $0-16
    // Read MIDR (Main ID Register)
    MRC p15, 0, r0, c0, c0, 0    // r0 = MIDR
    MOV r0, a+8(FP)             // store MIDR into return variable 'a'

    // Read CTR (Cache Type Register)
    MRC p15, 0, r0, c0, c0, 1    // r0 = CTR
    MOV r0, b+12(FP)            // store CTR into return variable 'b'

    // Read PFR0 (Processor Feature Register 0)
    MRC p15, 0, r0, c0, c1, 0    // r0 = PFR0
    MOV r0, c+16(FP)            // store PFR0 into return variable 'c'

    // Read MPIDR (Multiprocessor Affinity Register)
    MRC p15, 0, r0, c0, c0, 5    // r0 = MPIDR
    MOV r0, d+20(FP)            // store MPIDR into return variable 'd'
    RET
