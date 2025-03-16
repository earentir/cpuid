; cpuid16.asm - NASM version for native 16-bit DOS
; Assemble with: nasm -f obj cpuid16.asm -o cpuid16.obj
[BITS 16]

global _cpuid16

; _cpuid16:
; Expects DS:SI to point to a 16-byte buffer.
; It sets EAX=0 (CPUID function 0), then executes CPUID using the
; 0x66 operand-size override so that a 32-bit CPUID is performed,
; and finally writes EAX, EBX, ECX, EDX (each 4 bytes) into the buffer.
_cpuid16:
    push bp
    mov bp, sp

    xor ax, ax              ; Set EAX = 0
    db 0x66, 0x0F, 0xA2      ; Execute CPUID with 0x66 override

    ; Write EAX, EBX, ECX, and EDX into the 16-byte buffer pointed to by DS:SI.
    db 0x66, 0x89, 0x04      ; MOV dword [SI], EAX
    add si, 4
    db 0x66, 0x89, 0x1C      ; MOV dword [SI], EBX
    add si, 4
    db 0x66, 0x89, 0x0C      ; MOV dword [SI], ECX
    add si, 4
    db 0x66, 0x89, 0x14      ; MOV dword [SI], EDX

    pop bp
    ret
