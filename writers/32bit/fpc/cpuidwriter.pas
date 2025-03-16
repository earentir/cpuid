{ FreePascal CPUID File Writer }
program CPUIDWriter;

{$mode objfpc}
{$H+}

uses
  SysUtils;

type
  TCPUIDRegs = record
    a, b, c, d: LongWord;
  end;

function CPUID(eax, ecx: LongWord): TCPUIDRegs; assembler;
{$IFDEF CPU386}
asm
  { For 32-bit x86:
    Input:  eax and ecx are in registers.
    CPUID instruction returns:
      EAX -> a, EBX -> b, ECX -> c, EDX -> d.
    Store them into the result record.
  }
  cpuid
  mov [Result].TCPUIDRegs.a, eax
  mov [Result].TCPUIDRegs.b, ebx
  mov [Result].TCPUIDRegs.c, ecx
  mov [Result].TCPUIDRegs.d, edx
end;
{$ELSE}
begin
  // For non-x86 architectures, return zeros.
  Result.a := 0;
  Result.b := 0;
  Result.c := 0;
  Result.d := 0;
end;
{$ENDIF}

var
  regs: TCPUIDRegs;
  f: File;
begin
  regs := CPUID($0, 0);
  AssignFile(f, 'cpuid.bin');
  Rewrite(f, 1); // Open file in binary mode.
  BlockWrite(f, regs, SizeOf(regs));
  CloseFile(f);
end.
