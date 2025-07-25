# MPU Architecture Changes - Unmapped Registers

## Overview

This document describes the major architectural changes made to MPU to remove memory-mapped registers and introduce the PMI (Peripheral Management Interface) instruction.

## Key Changes

### 1. Removed Memory-Mapped Registers

Previously, MPU had the following memory-mapped registers at addresses 0x00-0x0F:
- `0x00-0x01`: PC (Program Counter)
- `0x02-0x03`: SP (Stack Pointer)  
- `0x04-0x05`: FP (Frame Pointer)
- `0x06-0x07`: IO Request Register
- `0x08-0x09`: IO Status Register
- `0x0A-0x0B`: Random Number Register

These have been removed from memory mapping. PC, SP, and FP are now internal CPU registers only.

### 2. Programs Can Start at Address 0

With memory-mapped registers removed, programs can now start execution at address 0x0000 instead of 0x0100. This gives programs access to the full 64KB of RAM.

### 3. New Request Instruction

The Request instruction 'req' replaces the old IO request mechanism:
- Syntax: `req <address>`
- Takes the address of a request structure
- Request structure format:
  - Offset 0: Request ID (2 bytes)
  - Offset 2: Status (2 bytes) - written by PMI handler
  - Offset 4+: Request-specific parameters

### 4. Updated IO Handlers

All IO handlers have been updated to work with PMI:
- Random number generator (ID: 0x0001)
- Stdout write (ID: 0x0101)
- SDL graphics operations (IDs: 0x0201-0x020C)

### 5. Register Access

- PC, SP, FP are no longer directly accessible by programs
- Programs use existing instructions (jmp, jsr, ret, psh, pop, sav, rst) to manipulate these registers
- Direct access could be added later with new instructions if needed

## Benefits

1. **Simplified Architecture**: No more confusing memory-mapped registers
2. **Full RAM Access**: Programs can use all 64KB of memory
3. **Cleaner IO Interface**: PMI provides a consistent interface for all peripherals
4. **Better Encapsulation**: CPU registers are properly internal

## Migration Guide

To update existing programs:

1. Remove any `org 0x100` directives - programs can start at 0
2. Replace `cpy <io_register>, #<request>` with `pmi <request>`
3. Update request structures to include status field at offset 2
4. Update any code that directly accessed PC, SP, or FP at memory addresses

## Example

Before:
```assembly
    org 0x100
main:
    cpy 0x06, #print_req
    hlt

print_req:
    dw 0x0101       // stdout write
    dw hello_msg
hello_msg:
    db "Hello!", 0
```

After:
```assembly
// Programs can now start at address 0!
main:
    pmi print_req
    hlt

print_req:
    dw 0x0101       // stdout write request ID
    dw 0            // status (written by PMI)
    dw hello_msg
hello_msg:
    db "Hello!", 0
```