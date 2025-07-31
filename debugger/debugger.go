// Copyright 2022 Jason Sando <jason.sando.lv@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package debugger

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
)

// Debugger provides debugging capabilities for MPU programs
type Debugger struct {
	machine   *machine.Machine
	memory    machine.Memory
	symbols   *asm.SymbolTable
	debugInfo []asm.DebugInfo
	sources   map[string][]string // File path -> lines of source code

	// Breakpoints
	breakpoints map[uint16]*Breakpoint
	bpIDCounter int

	// Execution control
	stepMode   StepMode
	stepTarget uint16 // Target PC for step operations
	callDepth  int    // Track call depth for step over/out

	// State
	running  bool
	mutex    sync.Mutex
	stopChan chan struct{}

	// Watch expressions
	watches map[string]struct{} // Set of variable names to watch

	// UI callbacks
	onBreakpoint func(bp *Breakpoint)
	onStep       func(pc uint16)
	onOutput     func(text string)
}

// StepMode defines how the debugger should step through code
type StepMode int

const (
	StepNone StepMode = iota
	StepInto
	StepOver
	StepOut
)

// Breakpoint represents a debugging breakpoint
type Breakpoint struct {
	ID        int
	Address   uint16
	File      string
	Line      int
	Enabled   bool
	HitCount  int
	Condition string // TODO: Implement conditional breakpoints
}

// NewDebugger creates a new debugger instance
func NewDebugger(m *machine.Machine, symbols *asm.SymbolTable, debugInfo []asm.DebugInfo) *Debugger {
	return &Debugger{
		machine:     m,
		memory:      m.Memory(),
		symbols:     symbols,
		debugInfo:   debugInfo,
		sources:     make(map[string][]string),
		breakpoints: make(map[uint16]*Breakpoint),
		watches:     make(map[string]struct{}),
		stopChan:    make(chan struct{}),
	}
}

// LoadSourceFile loads a source file for display during debugging
func (d *Debugger) LoadSourceFile(path string, lines []string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.sources[path] = lines
}

// SetBreakpoint sets a breakpoint at the given file:line
func (d *Debugger) SetBreakpoint(file string, line int) (*Breakpoint, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Find the PC address for this file:line
	addr, found := d.findAddressForLine(file, line)
	if !found {
		// Try to find nearest line with code
		var nearest *asm.DebugInfo
		minDist := 999999
		for i := range d.debugInfo {
			info := &d.debugInfo[i]
			if info.File == file {
				dist := abs(info.Line - line)
				if dist < minDist {
					minDist = dist
					nearest = info
				}
			}
		}

		if nearest != nil && minDist <= 2 {
			return nil, fmt.Errorf("no code at %s:%d (nearest code at line %d)", file, line, nearest.Line)
		}
		return nil, fmt.Errorf("no code at %s:%d", file, line)
	}

	// Check if breakpoint already exists
	if bp, exists := d.breakpoints[addr]; exists {
		bp.Enabled = true
		return bp, nil
	}

	// Create new breakpoint
	d.bpIDCounter++
	bp := &Breakpoint{
		ID:      d.bpIDCounter,
		Address: addr,
		File:    file,
		Line:    line,
		Enabled: true,
	}
	d.breakpoints[addr] = bp

	return bp, nil
}

// SetBreakpointAtAddress sets a breakpoint at a specific address
func (d *Debugger) SetBreakpointAtAddress(addr uint16) *Breakpoint {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if bp, exists := d.breakpoints[addr]; exists {
		bp.Enabled = true
		return bp
	}

	d.bpIDCounter++
	bp := &Breakpoint{
		ID:      d.bpIDCounter,
		Address: addr,
		Enabled: true,
	}

	// Try to find source location
	if info := d.findDebugInfo(addr); info != nil {
		bp.File = info.File
		bp.Line = info.Line
	}

	d.breakpoints[addr] = bp
	return bp
}

// RemoveBreakpoint removes a breakpoint by ID
func (d *Debugger) RemoveBreakpoint(id int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for addr, bp := range d.breakpoints {
		if bp.ID == id {
			delete(d.breakpoints, addr)
			return nil
		}
	}
	return fmt.Errorf("breakpoint %d not found", id)
}

// ListBreakpoints returns all breakpoints
func (d *Debugger) ListBreakpoints() []*Breakpoint {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var bps []*Breakpoint
	for _, bp := range d.breakpoints {
		bps = append(bps, bp)
	}
	return bps
}

// Run starts or continues execution until a breakpoint or completion
func (d *Debugger) Run() {
	d.mutex.Lock()
	d.running = true
	d.stepMode = StepNone
	d.mutex.Unlock()

	d.execute()
}

// StepInto executes one instruction
func (d *Debugger) StepInto() {
	d.mutex.Lock()
	d.stepMode = StepInto
	d.mutex.Unlock()

	d.execute()
}

// StepOver executes until the next instruction at the same call depth
func (d *Debugger) StepOver() {
	d.mutex.Lock()
	d.stepMode = StepOver
	d.callDepth = 0
	d.mutex.Unlock()

	d.execute()
}

// StepOut executes until returning from the current function
func (d *Debugger) StepOut() {
	d.mutex.Lock()
	d.stepMode = StepOut
	d.callDepth = 1
	d.mutex.Unlock()

	d.execute()
}

// Stop halts execution
func (d *Debugger) Stop() {
	d.mutex.Lock()
	d.running = false
	d.mutex.Unlock()

	select {
	case d.stopChan <- struct{}{}:
	default:
	}
}

// GetState returns the current machine state
func (d *Debugger) GetState() MachineState {
	flags := d.machine.Flags()
	pc := d.machine.Pc

	var currentFile string
	var currentLine int
	if info := d.findDebugInfo(pc); info != nil {
		currentFile = info.File
		currentLine = info.Line
	}

	return MachineState{
		PC:          pc,
		SP:          d.machine.Sp,
		FP:          d.machine.Fp,
		Negative:    flags.Negative,
		Zero:        flags.Zero,
		Carry:       flags.Carry,
		Bytes:       flags.Bytes,
		CurrentFile: currentFile,
		CurrentLine: currentLine,
	}
}

// MachineState represents the current state of the machine
type MachineState struct {
	PC, SP, FP     uint16
	Negative, Zero bool
	Carry, Bytes   bool
	CurrentFile    string
	CurrentLine    int
}

// ReadMemory reads memory at the given address
func (d *Debugger) ReadMemory(addr uint16, length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = d.memory.GetByte(addr + uint16(i))
	}
	return data
}

// findCurrentFunction returns the name of the function containing the given PC
func (d *Debugger) findCurrentFunction(pc uint16) string {
	// Find the function symbol that contains this PC
	var currentFunc string
	var funcStart uint16 = 0xFFFF

	// Iterate through all symbols to find function definitions
	for name, sym := range d.symbols.Symbols() {
		// Skip non-function symbols (functions don't have dots in their base name)
		if sym == nil || sym.IsFP() || contains(name, ".") {
			continue
		}

		// Check if this symbol's address is before PC but after our current best match
		symAddr := uint16(sym.Value())
		if symAddr <= pc && symAddr > funcStart {
			currentFunc = name
			funcStart = symAddr
		}
	}

	if currentFunc == "" {
		return "main"
	}
	return currentFunc
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetVariable looks up a variable by name and returns its value
func (d *Debugger) GetVariable(name string) (uint16, error) {
	// First try the exact name
	symbol := d.symbols.GetSymbol(name)

	// If not found, try with current function prefix
	if symbol == nil {
		currentFunc := d.findCurrentFunction(d.machine.Pc)
		symbol = d.symbols.GetSymbol(currentFunc + "." + name)
	}

	// If still not found, try with main. prefix for backwards compatibility
	if symbol == nil {
		symbol = d.symbols.GetSymbol("main." + name)
	}

	if symbol == nil {
		return 0, fmt.Errorf("symbol '%s' not found", name)
	}

	// Check if it's a frame-pointer relative variable
	if symbol.IsFP() {
		// Get current frame pointer
		fp := d.machine.Fp
		// Calculate actual address: FP + offset
		// Note: symbol.Value() contains the offset (negative for locals)
		addr := uint16(int(fp) + symbol.Value())
		return d.memory.GetWord(addr), nil
	}

	// Regular variable - read value at the symbol's address
	return d.memory.GetWord(uint16(symbol.Value())), nil
}

// GetCallStack returns the current call stack
func (d *Debugger) GetCallStack() []StackFrame {
	// TODO: Implement proper call stack tracking
	// This would require tracking JSR/RTS instructions
	return []StackFrame{}
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	PC       uint16
	Function string
	File     string
	Line     int
}

// Internal methods

func (d *Debugger) execute() {
	firstStep := true

	for {
		d.mutex.Lock()
		if !d.running && d.stepMode == StepNone {
			d.mutex.Unlock()
			break
		}

		pc := d.machine.Pc

		// Check for breakpoint (but not on first step when stepping)
		if !firstStep || d.stepMode == StepNone {
			if bp, exists := d.breakpoints[pc]; exists && bp.Enabled {
				bp.HitCount++
				d.running = false
				d.stepMode = StepNone
				d.mutex.Unlock()

				if d.onBreakpoint != nil {
					d.onBreakpoint(bp)
				}
				break
			}
		}

		// Track instruction for JSR/RTS detection
		inst := d.memory.GetByte(pc)
		opCode, _, _ := machine.DecodeOp(inst)

		d.mutex.Unlock()

		// Track call depth for step over/out
		if opCode == machine.Jsr {
			d.callDepth++
		} else if opCode == machine.Ret || opCode == machine.Rst {
			d.callDepth--
		}

		// Execute one instruction
		newPC := d.machine.Step(pc)

		// Check if we hit a HLT instruction or PC didn't change (halted)
		if newPC == pc || opCode == machine.Hlt {
			// Program halted
			if d.onStep != nil {
				d.onStep(pc)
			}
			break
		}

		// Now check if we should stop after executing the instruction
		if d.shouldStop() {
			d.mutex.Lock()
			d.running = false
			d.stepMode = StepNone
			d.mutex.Unlock()

			if d.onStep != nil {
				d.onStep(newPC)
			}
			break
		}

		firstStep = false

		// Check for stop signal
		select {
		case <-d.stopChan:
			d.mutex.Lock()
			d.running = false
			d.mutex.Unlock()
			return
		default:
		}
	}
}

func (d *Debugger) shouldStop() bool {
	switch d.stepMode {
	case StepInto:
		return true
	case StepOver:
		return d.callDepth <= 0
	case StepOut:
		return d.callDepth < 0
	default:
		return false
	}
}

func (d *Debugger) findAddressForLine(file string, line int) (uint16, bool) {
	// Try exact match first
	for _, info := range d.debugInfo {
		if info.File == file && info.Line == line {
			return info.PC, true
		}
	}

	// Try matching by base name
	baseFile := filepath.Base(file)
	for _, info := range d.debugInfo {
		if filepath.Base(info.File) == baseFile && info.Line == line {
			return info.PC, true
		}
	}

	return 0, false
}

// FindAddressForLine finds the PC address for a given file:line (public method)
func (d *Debugger) FindAddressForLine(file string, line int) (uint16, bool) {
	return d.findAddressForLine(file, line)
}

func (d *Debugger) findDebugInfo(pc uint16) *asm.DebugInfo {
	// Find the debug info with the closest PC <= target
	var best *asm.DebugInfo
	for i := range d.debugInfo {
		info := &d.debugInfo[i]
		if info.PC <= pc && (best == nil || info.PC > best.PC) {
			best = info
		}
	}
	return best
}

// AddWatch adds a variable to the watch list
func (d *Debugger) AddWatch(name string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.watches[name] = struct{}{}
}

// RemoveWatch removes a variable from the watch list
func (d *Debugger) RemoveWatch(name string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.watches, name)
}

// ListWatches returns all watched variables
func (d *Debugger) ListWatches() []string {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var watches []string
	for name := range d.watches {
		watches = append(watches, name)
	}
	return watches
}

// GetWatchValues returns current values of all watched variables
func (d *Debugger) GetWatchValues() map[string]string {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	values := make(map[string]string)
	for name := range d.watches {
		value, err := d.GetVariable(name)
		if err != nil {
			values[name] = fmt.Sprintf("Error: %v", err)
		} else {
			values[name] = fmt.Sprintf("%04x (%d)", value, value)
		}
	}
	return values
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
