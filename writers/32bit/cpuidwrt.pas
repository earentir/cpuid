program cpuidwrt;

{$APPTYPE CONSOLE}
{$MODE DELPHI}  { Delphi mode for better compatibility }
{$IFDEF UNIX}
  {$DEFINE POSIX}
{$ENDIF}

uses
  SysUtils, Classes;

type
  TEntry = record
    Leaf: LongWord;
    SubLeaf: LongWord;
    EAX: LongWord;
    EBX: LongWord;
    ECX: LongWord;
    EDX: LongWord;
  end;

var
  Entries: array of TEntry;

{------------------------------------------------------------
  cpuid executes the CPUID instruction.
  Parameters:
    leaf, subleaf: Input registers.
    eaxOut, ebxOut, ecxOut, edxOut: Output registers.
  Note: We push/pop EBX because it is callee‐saved.
-------------------------------------------------------------}
procedure cpuid(Leaf, SubLeaf: LongWord; var EAXOut, EBXOut, ECXOut, EDXOut: LongWord);
asm
  push ebx
  mov eax, Leaf
  mov ecx, SubLeaf
  cpuid
  mov EBXOut, ebx
  mov EAXOut, eax
  mov ECXOut, ecx
  mov EDXOut, edx
  pop ebx
end;

procedure AppendEntry(Leaf, SubLeaf, EAX, EBX, ECX, EDX: LongWord);
var
  L: Integer;
begin
  L := Length(Entries);
  SetLength(Entries, L + 1);
  with Entries[L] do
  begin
    Self.Leaf := Leaf;
    Self.SubLeaf := SubLeaf;
    Self.EAX := EAX;
    Self.EBX := EBX;
    Self.ECX := ECX;
    Self.EDX := EDX;
  end;
end;

procedure CaptureData;
var
  leaf, subleaf: LongWord;
  maxStandard, maxExtended: LongWord;
  a, b, c, d: LongWord;
begin
  SetLength(Entries, 0);

  {--- Capture Standard CPUID Leaves ---}
  cpuid(0, 0, maxStandard, b, c, d);  { maxStandard is in EAX }
  for leaf := 0 to maxStandard do
  begin
    if (leaf = 4) or (leaf = $B) or (leaf = $D) then
    begin
      subleaf := 0;
      while True do
      begin
        cpuid(leaf, subleaf, a, b, c, d);
        if (leaf = 4) and (subleaf > 0) and ((a and $1F) = 0) then Break;
        if (leaf = $B) and (subleaf > 0) and (a = 0) then Break;
        if (leaf = $D) and (subleaf > 0) and (a = 0) and (b = 0) and (c = 0) and (d = 0) then Break;
        AppendEntry(leaf, subleaf, a, b, c, d);
        Inc(subleaf);
      end;
    end
    else
    begin
      cpuid(leaf, 0, a, b, c, d);
      AppendEntry(leaf, 0, a, b, c, d);
    end;
  end;

  {--- Capture Extended CPUID Leaves ---}
  cpuid($80000000, 0, maxExtended, b, c, d);  { maxExtended in EAX }
  for leaf := $80000000 to maxExtended do
  begin
    if (leaf = $8000001D) then
    begin
      subleaf := 0;
      while True do
      begin
        cpuid(leaf, subleaf, a, b, c, d);
        if (subleaf > 0) and ((a and $1F) = 0) then Break;
        AppendEntry(leaf, subleaf, a, b, c, d);
        Inc(subleaf);
      end;
    end
    else
    begin
      cpuid(leaf, 0, a, b, c, d);
      AppendEntry(leaf, 0, a, b, c, d);
    end;
  end;
end;

procedure WriteJSON(const AFileName: string);
var
  i: Integer;
  SL: TStringList;
begin
  SL := TStringList.Create;
  try
    SL.Add('{');
    SL.Add('  "entries": [');
    for i := 0 to High(Entries) do
    begin
      with Entries[i] do
      begin
        { For simplicity, every entry is followed by a comma except the last }
        if i < High(Entries) then
          SL.Add(Format('    { "leaf": %u, "subleaf": %u, "eax": %u, "ebx": %u, "ecx": %u, "edx": %u },',
            [Leaf, SubLeaf, EAX, EBX, ECX, EDX]))
        else
          SL.Add(Format('    { "leaf": %u, "subleaf": %u, "eax": %u, "ebx": %u, "ecx": %u, "edx": %u }',
            [Leaf, SubLeaf, EAX, EBX, ECX, EDX]));
      end;
    end;
    SL.Add('  ]');
    SL.Add('}');
    SL.SaveToFile(AFileName);
  finally
    SL.Free;
  end;
end;

begin
  try
    CaptureData;
    WriteJSON('cpuid_data.json');
    Writeln('CPUID data captured in cpuid_data.json');
  except
    on E: Exception do
      Writeln('Error: ', E.Message);
  end;
end.
