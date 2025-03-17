@echo off
REM Build script for the Turbo Pascal CPUID project

REM Assemble the CPUID.ASM source file using TASM
echo Assembling CPUID.ASM...
tasm cpuid.asm
IF ERRORLEVEL 1 (
    echo Error: TASM failed to assemble CPUID.ASM.
    pause
    goto :EOF
)

REM Optionally, you can link the object file with TLINK if needed.
REM In many cases Turbo Pascal will automatically link CPUID.OBJ,
REM so the next step is to compile the Pascal source.

echo Compiling CPUIDTP.PAS with TPC...
tpc cpuidtp.pas
IF ERRORLEVEL 1 (
    echo Error: TPC failed to compile CPUIDTP.PAS.
    pause
    goto :EOF
)

echo Build completed successfully.
pause
