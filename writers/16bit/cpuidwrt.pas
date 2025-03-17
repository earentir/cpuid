{$L CPUID.OBJ}

program CPUIDTP;

uses
  Crt, Dos;

type
  TEntry = record
    Leaf: LongInt;
    SubLeaf: LongInt;
    EAX: LongInt;
    EBX: LongInt;
    ECX: LongInt;
    EDX: LongInt;
  end;

var
  F: Text;

{--------------------------------------------------------------------
  CPUID is declared as an external procedure. It takes two 32‑bit input
  parameters (Leaf and SubLeaf) and four var parameters (to receive the
  32‑bit values returned in EAX, EBX, ECX, and EDX). The routine is
  provided by CPUID.ASM.
--------------------------------------------------------------------}
procedure CPUID(Leaf, SubLeaf: LongInt; var EAX, EBX, ECX, EDX: LongInt); external name '_CPUIDReal';

{--------------------------------------------------------------------
  WriteEntry writes one JSON object (an entry) to the text file.
--------------------------------------------------------------------}
procedure WriteEntry(Leaf, SubLeaf, EAX, EBX, ECX, EDX: LongInt; IsLast: Boolean);
begin
  Write(F, '    { "leaf": ', Leaf, ', "subleaf": ', SubLeaf,
    ', "eax": ', EAX, ', "ebx": ', EBX, ', "ecx": ', ECX, ', "edx": ', EDX, ' }');
  if not IsLast then
    Writeln(F, ',')
  else
    Writeln(F);
end;

var
  maxStandard, maxExtended: LongInt;
  leaf, subleaf: LongInt;
  a, b, c, d: LongInt;
begin
  ClrScr;
  Assign(F, 'cpuid_data.json');
  Rewrite(F);
  Writeln(F, '{');
  Writeln(F, '  "entries": [');

  { --- Capture Standard CPUID Leaves --- }
  { Call CPUID(0,0) which returns the maximum standard leaf in EAX. }
  CPUID(0, 0, maxStandard, b, c, d);
  for leaf := 0 to maxStandard do
  begin
    if (leaf = 4) or (leaf = $B) or (leaf = $D) then
    begin
      subleaf := 0;
      while True do
      begin
        CPUID(leaf, subleaf, a, b, c, d);
        { For leaf 4: if subleaf > 0 and the lower 5 bits of EAX are zero, break. }
        if (leaf = 4) and (subleaf > 0) and ((a and $1F) = 0) then Break;
        { For leaf $B: if subleaf > 0 and EAX is zero, break. }
        if (leaf = $B) and (subleaf > 0) and (a = 0) then Break;
        { For leaf $D: if subleaf > 0 and all registers are zero, break. }
        if (leaf = $D) and (subleaf > 0) and (a = 0) and (b = 0) and (c = 0) and (d = 0) then Break;
        WriteEntry(leaf, subleaf, a, b, c, d, False);
        Inc(subleaf);
      end;
    end
    else
    begin
      CPUID(leaf, 0, a, b, c, d);
      WriteEntry(leaf, 0, a, b, c, d, False);
    end;
  end;

  { --- Capture Extended CPUID Leaves --- }
  CPUID($80000000, 0, maxExtended, b, c, d);
  for leaf := $80000000 to maxExtended do
  begin
    if leaf = $8000001D then
    begin
      subleaf := 0;
      while True do
      begin
        CPUID(leaf, subleaf, a, b, c, d);
        if (subleaf > 0) and ((a and $1F) = 0) then Break;
        WriteEntry(leaf, subleaf, a, b, c, d, False);
        Inc(subleaf);
      end;
    end
    else
    begin
      CPUID(leaf, 0, a, b, c, d);
      WriteEntry(leaf, 0, a, b, c, d, False);
    end;
  end;

  { --- Finalize JSON --- }
  Writeln(F, '  ]');
  Writeln(F, '}');
  Close(F);

  WriteLn('CPUID data captured in cpuid_data.json');
  WriteLn('Press any key to exit...');
  ReadKey;
end.
