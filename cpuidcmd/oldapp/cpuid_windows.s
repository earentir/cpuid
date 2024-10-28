// +build windows
// +build amd64 386

// cpuid_windows.s

TEXT ·cpuid(SB), $0-24
    MOVL eax+0(FP), AX     // Load 'eax' argument into AX register (32-bit)
    MOVL ecx+4(FP), CX     // Load 'ecx' argument into CX register (32-bit)
    CPUID                  // Execute the CPUID instruction
    MOVL AX, a+8(FP)       // Store AX result into return variable 'a' (32-bit)
    MOVL BX, b+12(FP)      // Store BX result into return variable 'b' (32-bit)
    MOVL CX, c+16(FP)      // Store CX result into return variable 'c' (32-bit)
    MOVL DX, d+20(FP)      // Store DX result into return variable 'd' (32-bit)
    RET                    // Return from the function
