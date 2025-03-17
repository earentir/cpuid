; CPUID.ASM
; Assemble with TASM or MASM as a 16-bit object using the .386 directive.
.MODEL SMALL
.386
.CODE
PUBLIC _CPUIDReal

;
; _CPUIDReal is called from Turbo Pascal with the following parameters on the stack:
;   [BP+4]  : 32-bit Leaf (low word at [BP+4], high word at [BP+6])
;   [BP+8]  : 32-bit SubLeaf (low word at [BP+8], high word at [BP+10])
;   [BP+12] : Pointer to store EAX (32-bit)
;   [BP+16] : Pointer to store EBX (32-bit)
;   [BP+20] : Pointer to store ECX (32-bit)
;   [BP+24] : Pointer to store EDX (32-bit)
;
_CPUIDReal PROC FAR
    push bp
    mov bp, sp

    { Load Leaf into EAX (using operand-size override to load 32 bits) }
    db 66h
    mov eax, dword ptr [bp+4]
    { Load SubLeaf into ECX }
    db 66h
    mov ecx, dword ptr [bp+8]

    { Execute CPUID }
    db 66h, 0Fh, A2h

    { Store results.
      The pointer to store EAX is at [BP+12], EBX at [BP+16], etc.
      Use the 66h prefix to move 32-bit values. }
    db 66h
    mov bx, word ptr [bp+12]
    mov dword ptr [bx], eax

    db 66h
    mov bx, word ptr [bp+16]
    mov dword ptr [bx], ebx

    db 66h
    mov bx, word ptr [bp+20]
    mov dword ptr [bx], ecx

    db 66h
    mov bx, word ptr [bp+24]
    mov dword ptr [bx], edx

    pop bp
    retf
_CPUIDReal ENDP
END
