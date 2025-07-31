# MPU Enhanced Debugger

The MPU debugger provides advanced debugging capabilities for MPU assembly programs, including breakpoints, source-level debugging, and memory inspection.

## Features

- **Source-level debugging**: Step through your assembly code with source file and line information
- **Breakpoints**: Set breakpoints by file:line or memory address
- **Step operations**: Step into, step over, and step out of function calls
- **Memory inspection**: Examine memory contents and variables
- **Register inspection**: View current CPU state including PC, SP, FP, and flags
- **Disassembly**: View disassembled instructions when source is unavailable

## Usage

Run a program with the debugger using the `-d` flag:

```bash
mpu run -d program.s
```

## Commands

### Execution Control
- `run` or `r` - Continue execution until breakpoint or program end
- `step` or `s` - Step into (execute one instruction)
- `next` or `n` - Step over (skip function calls)
- `finish` or `f` - Step out (run until current function returns)
- `quit` or `q` - Exit debugger

### Breakpoints
- `break <location>` or `b <location>` - Set breakpoint
  - By file:line: `b main.s:10`
  - By address: `b 0x1000`
  - List all: `b` (no arguments)
- `delete <id>` or `d <id>` - Delete breakpoint by ID

### Code Viewing
- `list` or `l` - Show source code around current location
- `list <location>` - Show source at specific location
- `disasm` or `dis` - Disassemble instructions at current location
- `disasm <addr>` - Disassemble at specific address

### Inspection
- `print <var>` or `p <var>` - Print variable value
  - Special registers: `p pc`, `p sp`, `p fp`
  - Variables: `p myvar`
- `x <addr> [count]` - Examine memory (hex dump)
- `info regs` - Show all registers
- `info break` - List breakpoints
- `info stack` - Show call stack (when implemented)

## Example Session

```
$ mpu run -d example/debug_test.s
╔═══════════════════════════════════════════╗
║     MPU Debugger - Enhanced Edition       ║
╚═══════════════════════════════════════════╝
Type 'help' for available commands

PC: 0x0000 (debug_test.s:4)
=>    4      cpy r1,#10      // r1 = 10
      5      cpy r2,#20      // r2 = 20

(mpu-dbg) b 8
Breakpoint 1 set at debug_test.s:8 (0x000a)
(mpu-dbg) r
Continuing...

Breakpoint 1 hit at 0x000a (debug_test.s:8)
PC: 0x000a (debug_test.s:8)
=>    8      cmp r3,#30      // check if r3 == 30
      9      jeq success     // jump to success if equal

(mpu-dbg) p r3
r3 = 0x001e (30)
(mpu-dbg) info regs
PC: 0x000a  SP: 0xffff  FP: 0x0000
Flags: 
(mpu-dbg) n
PC: 0x000d (debug_test.s:9)
=>    9      jeq success     // jump to success if equal
     10      hlt             // halt if not equal

(mpu-dbg) s
PC: 0x000f (debug_test.s:13)
=>   13      cpy 0x2000,#1   // write 1 to memory
     14      hlt             // halt

(mpu-dbg) x 0x2000 4
0x2000: 00 00 00 00
(mpu-dbg) n
PC: 0x0012 (debug_test.s:14)
=>   14      hlt             // halt
     15      

(mpu-dbg) x 0x2000 4
0x2000: 01 00 00 00
(mpu-dbg) q
```

## Future Enhancements

### VSCode Extension
A VSCode extension could provide:
- Visual breakpoints in the editor gutter
- Variable watches
- Call stack visualization
- Memory view panels
- Integrated terminal for debugger commands

### Web-based Debugger
An interactive web interface could enable:
- Online MPU development and debugging
- Visual memory maps
- Step-by-step execution visualization
- Shareable debugging sessions

## Implementation Notes

The debugger leverages the source mapping added to the MPU assembler, which tracks the relationship between assembled addresses and source file locations. This enables accurate source-level debugging even after the assembly process.

The debugger operates by intercepting the machine's execution at each instruction, checking for breakpoints, and implementing stepping logic by tracking JSR/RTS instructions for call depth management.