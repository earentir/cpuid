program cpuidwrt;

{$M 1024,0,655360}  { Memory model directive for 16-bit DOS }
{$APPTYPE CONSOLE}
{$IFDEF FPC}
  {$MODE TP}       { Turbo Pascal compatibility mode for native 16-bit DOS }
{$ENDIF}

uses
  Dos;

type
  TBuffer = array[0..15] of Byte;

{ Declare the external assembly routine.
  The routine _cpuid16 is assembled in cpuid16.obj and expects a pointer
  to a 16-byte buffer (in DS:SI).
}
procedure cpuid16(var Buffer: TBuffer); external;

var
  Buffer: TBuffer;
  F: file;
begin
  cpuid16(Buffer);
  Assign(F, 'cpuid.bin');
  Rewrite(F, 1);  { Binary file, 1-byte record size }
  BlockWrite(F, Buffer, SizeOf(Buffer));
  Close(F);
end.
