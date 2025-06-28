# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MPU (Memory Processing Unit) is a 16-bit virtual computer system written in Go. It includes:
- A complete virtual machine with no general-purpose registers (all operations on memory)
- An assembler for MPU assembly language
- SDL-based graphics and audio support
- Example programs including games (pong, blocks) and demos

## Common Development Commands

### Building
```bash
# Install dependencies (macOS)
brew install sdl2{,_image,_mixer,_ttf,_gfx} pkg-config

# Install dependencies (Ubuntu)
apt install pkg-config build-essential libsdl2{,-image,-mixer,-ttf,-gfx}-dev

# Build the project
go build
```

### Running Programs
```bash
# Run an assembly program directly
./mpu run example/hello.s

# Run with monitor/debugger mode
./mpu run -m example/hello.s

# Compile assembly to binary
./mpu build -o output.bin input.s

# Format assembly source
./mpu fmt input.s
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./asm
go test ./machine

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestParseLineNumbers ./asm
```

## Code Architecture

### Package Structure
- **`/asm`** - Assembler implementation
  - `lexer.go` - Tokenizes assembly source files
  - `parser.go` - Parses tokens into AST statements
  - `linker.go` - Links and generates machine code
  - `printer.go` - Formats assembly source
  - `symbols.go` - Symbol table management
  
- **`/machine`** - Virtual machine implementation
  - `machine.go` - Core CPU emulation and instruction execution
  - `memory.go` - 64KB memory management
  - `encoding.go` - Instruction encoding/decoding
  - `io.go` - Peripheral Management Interface (PMI)
  - `graphics.go` - SDL graphics adapter
  - `stdout.go` - Standard output device
  - `rng.go` - Random number generator

### Key Design Patterns

1. **Memory-Mapped Architecture**: No general-purpose registers; PC, SP, and FP are at memory addresses 0x00-0x05

2. **Variable-Length Instructions**: Opcodes are 1 byte, followed by 0-4 bytes depending on addressing mode

3. **PMI (Peripheral Management Interface)**: All I/O goes through memory addresses 0x06 (request) and 0x08 (status)

4. **Two-Pass Assembly**: Parser creates AST, then linker resolves symbols and generates code

5. **Addressing Modes**: Immediate (#), Absolute (a), Indirect (*), Relative (r), and combinations

### Testing Approach
- Standard Go testing with `*_test.go` files
- Uses `github.com/stretchr/testify` for assertions
- Parser tests verify AST construction
- Machine tests verify instruction execution
- Memory tests verify read/write operations

### Important Files to Read First
1. `/machine/machine.go` - Core CPU implementation
2. `/asm/parser.go` - Assembly language syntax
3. `/example/hello.s` - Simple assembly example
4. `/example/stdlib/stdio.s` - Standard library implementation