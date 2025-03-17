@echo off
REM Build script for the DOS CPUID C project

REM Assemble the CPUID.ASM file using TASM.
echo Assembling CPUID.ASM...
tasm cpuid.asm
IF ERRORLEVEL 1 (
    echo Error: TASM failed to assemble CPUID.ASM.
    pause
    goto :EOF
)

REM Compile the C source file.
REM (Replace 'tcc' with your DOS C compiler command if different.)
echo Compiling cpuid.c...
tcc -ml cpuid.c cpuid.obj
IF ERRORLEVEL 1 (
    echo Error: Compilation failed.
    pause
    goto :EOF
)

echo Build completed successfully.
pause
