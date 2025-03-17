#!/bin/bash
# build.sh - Build script for cpuidwrt.pas (FreePascal) and cpuidwrt.c (C)
# Produces Linux and Windows executables in both 32-bit and 64-bit modes.

set -e

echo "== FreePascal Builds =="

# Linux 32-bit FreePascal build: force i386 target.
echo "Building FreePascal Linux 32-bit..."
fpc -Pi386 cpuidwrt.pas -o cpuidwrt_linux_fpc32 || { echo "FPC Linux 32-bit build failed"; exit 1; }

# Linux 64-bit FreePascal build: force x86_64 target.
echo "Building FreePascal Linux 64-bit..."
fpc -Px86_64 cpuidwrt.pas -o cpuidwrt_linux_fpc64 || { echo "FPC Linux 64-bit build failed"; exit 1; }

# Windows 32-bit FreePascal build.
echo "Building FreePascal Windows 32-bit..."
fpc -Twin32 -Pi386 cpuidwrt.pas -o cpuidwrt_win32_fpc.exe || { echo "FPC Windows 32-bit build failed"; exit 1; }

# Windows 64-bit FreePascal build.
echo "Building FreePascal Windows 64-bit..."
fpc -Twin64 -Px86_64 cpuidwrt.pas -o cpuidwrt_win64_fpc.exe || { echo "FPC Windows 64-bit build failed"; exit 1; }

echo "== C Builds =="

# Linux 32-bit C build.
echo "Building C Linux 32-bit..."
gcc -m32 cpuidwrt.c -o cpuidwrt_linux_c32 || { echo "GCC Linux 32-bit build failed"; exit 1; }

# Linux 64-bit C build.
echo "Building C Linux 64-bit..."
gcc -m64 cpuidwrt.c -o cpuidwrt_linux_c64 || { echo "GCC Linux 64-bit build failed"; exit 1; }

# Windows 32-bit C build using MinGW cross-compiler.
echo "Building C Windows 32-bit..."
i686-w64-mingw32-gcc -m32 cpuidwrt.c -o cpuidwrt_win32_c.exe || { echo "MinGW Windows 32-bit build failed"; exit 1; }

# Windows 64-bit C build using MinGW cross-compiler.
echo "Building C Windows 64-bit..."
x86_64-w64-mingw32-gcc -m64 cpuidwrt.c -o cpuidwrt_win64_c.exe || { echo "MinGW Windows 64-bit build failed"; exit 1; }

echo "Build completed successfully."
