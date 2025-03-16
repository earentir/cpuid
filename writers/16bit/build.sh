#!/bin/bash
set -e

# Check if nasm is installed; if not, install it.
if ! command -v nasm >/dev/null 2>&1; then
    echo "nasm not found. Installing nasm..."
    sudo dnf install -y nasm
else
    echo "nasm is installed."
fi

# Check if FreePascal (fpc) is installed; if not, install it.
if ! command -v fpc >/dev/null 2>&1; then
    echo "FreePascal (fpc) not found. Installing fpc..."
    sudo dnf install -y fpc
else
    echo "FreePascal (fpc) is installed."
fi

# Check if OpenWatcom's compiler (wcl386) is available.
if ! command -v wcl386 >/dev/null 2>&1; then
    echo "OpenWatcom (wcl386) not found. Please install OpenWatcom manually."
    exit 1
else
    echo "OpenWatcom (wcl386) is installed."
fi

# Assemble the 16-bit DOS CPUID routine using NASM.
echo "Assembling cpuid16.asm..."
nasm -f obj cpuid16.asm -o cpuid16.obj

# Compile the DOS C version using OpenWatcom.
echo "Compiling cpuidwrt.c with OpenWatcom..."
wcl386 -bt=dos -ms -zp4 cpuidwrt.c cpuid16.obj -fo=cpuidwrt.com

# Compile the DOS FreePascal version.
echo "Compiling cpuidwrt.pas with FreePascal for DOS..."
fpc -Tdos cpuidwrt.pas

echo "Build complete."
