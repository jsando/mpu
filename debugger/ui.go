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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/jsando/mpu/machine"
	"github.com/jsando/mpu/test"
)

// DebuggerUI provides a terminal interface for the debugger
type DebuggerUI struct {
	debugger *Debugger
	rl       *readline.Instance
	writer   io.Writer

	// Source file handling
	sourceReader *test.SourceReader
	lastFile     string
	lastLine     int
}

// NewDebuggerUI creates a new debugger UI
func NewDebuggerUI(debugger *Debugger) *DebuggerUI {
	return &DebuggerUI{
		debugger:     debugger,
		writer:       os.Stdout,
		sourceReader: test.NewSourceReader(),
	}
}

// Run starts the interactive debugger session
func (ui *DebuggerUI) Run() {
	// Configure readline
	config := &readline.Config{
		Prompt:          "(mpu-dbg) ",
		HistoryFile:     "/tmp/mpu_debugger_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		// Auto-complete commands
		AutoComplete: readline.NewPrefixCompleter(
			readline.PcItem("run"),
			readline.PcItem("r"),
			readline.PcItem("step"),
			readline.PcItem("s"),
			readline.PcItem("next"),
			readline.PcItem("n"),
			readline.PcItem("finish"),
			readline.PcItem("f"),
			readline.PcItem("break"),
			readline.PcItem("b"),
			readline.PcItem("delete"),
			readline.PcItem("d"),
			readline.PcItem("list"),
			readline.PcItem("l"),
			readline.PcItem("print"),
			readline.PcItem("p"),
			readline.PcItem("x"),
			readline.PcItem("info",
				readline.PcItem("regs"),
				readline.PcItem("registers"),
				readline.PcItem("break"),
				readline.PcItem("breakpoints"),
				readline.PcItem("stack"),
				readline.PcItem("watches"),
			),
			readline.PcItem("watch"),
			readline.PcItem("w"),
			readline.PcItem("unwatch"),
			readline.PcItem("disasm"),
			readline.PcItem("dis"),
			readline.PcItem("help"),
			readline.PcItem("h"),
			readline.PcItem("quit"),
			readline.PcItem("q"),
		),
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		fmt.Fprintf(ui.writer, "Error initializing readline: %v\n", err)
		fmt.Fprintln(ui.writer, "Falling back to basic input mode")
		ui.runBasicMode()
		return
	}
	defer rl.Close()

	ui.rl = rl
	ui.writer = rl.Stdout()

	ui.printWelcome()

	// Set up callbacks
	ui.debugger.onBreakpoint = ui.handleBreakpoint
	ui.debugger.onStep = ui.handleStep

	// Show initial state
	ui.showCurrentLocation()

	// Main command loop
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			if err == io.EOF {
				fmt.Fprintln(ui.writer, "\nExiting debugger.")
				break
			}
			fmt.Fprintf(ui.writer, "Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !ui.handleCommand(line) {
			break
		}
	}
}

func (ui *DebuggerUI) printWelcome() {
	fmt.Fprintln(ui.writer, "╔═════════════════════════╗")
	fmt.Fprintln(ui.writer, "║      MPU Debugger       ║")
	fmt.Fprintln(ui.writer, "╚═════════════════════════╝")
	fmt.Fprintln(ui.writer, "Type 'help' for available commands")
	fmt.Fprintln(ui.writer)
}

func (ui *DebuggerUI) handleCommand(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return true
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help", "h":
		ui.showHelp()

	case "run", "r":
		ui.cmdRun()

	case "step", "s":
		ui.cmdStep()

	case "next", "n":
		ui.cmdNext()

	case "finish", "f":
		ui.cmdFinish()

	case "break", "b":
		ui.cmdBreak(args)

	case "delete", "d":
		ui.cmdDelete(args)

	case "list", "l":
		ui.cmdList(args)

	case "print", "p":
		ui.cmdPrint(args)

	case "x":
		ui.cmdExamine(args)

	case "info":
		ui.cmdInfo(args)

	case "disasm", "dis":
		ui.cmdDisasm(args)

	case "watch", "w":
		ui.cmdWatch(args)

	case "unwatch":
		ui.cmdUnwatch(args)

	case "quit", "q":
		return false

	default:
		fmt.Fprintf(ui.writer, "Unknown command: %s\n", cmd)
		fmt.Fprintln(ui.writer, "Type 'help' for available commands")
	}

	return true
}

func (ui *DebuggerUI) showHelp() {
	fmt.Fprintln(ui.writer, "Available commands:")
	fmt.Fprintln(ui.writer, "  run (r)              - Continue execution")
	fmt.Fprintln(ui.writer, "  step (s)             - Step into (execute one instruction)")
	fmt.Fprintln(ui.writer, "  next (n)             - Step over (skip function calls)")
	fmt.Fprintln(ui.writer, "  finish (f)           - Step out (run until function returns)")
	fmt.Fprintln(ui.writer, "  break (b) [location] - Set breakpoint (file:line or address)")
	fmt.Fprintln(ui.writer, "  delete (d) [id]      - Delete breakpoint")
	fmt.Fprintln(ui.writer, "  list (l) [location]  - Show source code")
	fmt.Fprintln(ui.writer, "  print (p) [expr]     - Print variable or expression")
	fmt.Fprintln(ui.writer, "  x [addr] [count]     - Examine memory")
	fmt.Fprintln(ui.writer, "  info [what]          - Show information (regs, break, stack, watches)")
	fmt.Fprintln(ui.writer, "  watch (w) [var]      - Add variable to watch list")
	fmt.Fprintln(ui.writer, "  unwatch [var]        - Remove variable from watch list")
	fmt.Fprintln(ui.writer, "  disasm (dis) [addr]  - Disassemble instructions")
	fmt.Fprintln(ui.writer, "  quit (q)             - Exit debugger")
}

func (ui *DebuggerUI) cmdRun() {
	fmt.Fprintln(ui.writer, "Continuing...")
	ui.debugger.Run()
}

func (ui *DebuggerUI) cmdStep() {
	ui.debugger.StepInto()
}

func (ui *DebuggerUI) cmdNext() {
	ui.debugger.StepOver()
}

func (ui *DebuggerUI) cmdFinish() {
	ui.debugger.StepOut()
}

func (ui *DebuggerUI) cmdBreak(args []string) {
	if len(args) == 0 {
		// List breakpoints
		ui.listBreakpoints()
		return
	}

	location := args[0]

	// Check if it's an address (starts with 0x or is a number)
	if addr, err := ui.parseAddress(location); err == nil {
		bp := ui.debugger.SetBreakpointAtAddress(addr)
		fmt.Fprintf(ui.writer, "Breakpoint %d set at 0x%04x", bp.ID, bp.Address)
		if bp.File != "" {
			fmt.Fprintf(ui.writer, " (%s:%d)", filepath.Base(bp.File), bp.Line)
		}
		fmt.Fprintln(ui.writer)
		return
	}

	// Try file:line format
	parts := strings.Split(location, ":")
	if len(parts) == 2 {
		file := parts[0]
		line, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Fprintf(ui.writer, "Invalid line number: %s\n", parts[1])
			return
		}

		bp, err := ui.debugger.SetBreakpoint(file, line)
		if err != nil {
			fmt.Fprintf(ui.writer, "Error setting breakpoint: %v\n", err)
			return
		}

		fmt.Fprintf(ui.writer, "Breakpoint %d set at %s:%d (0x%04x)\n",
			bp.ID, filepath.Base(bp.File), bp.Line, bp.Address)
		return
	}

	fmt.Fprintln(ui.writer, "Invalid breakpoint location. Use file:line or address")
}

func (ui *DebuggerUI) cmdDelete(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: delete <breakpoint-id>")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(ui.writer, "Invalid breakpoint ID: %s\n", args[0])
		return
	}

	if err := ui.debugger.RemoveBreakpoint(id); err != nil {
		fmt.Fprintf(ui.writer, "Error: %v\n", err)
		return
	}

	fmt.Fprintf(ui.writer, "Breakpoint %d deleted\n", id)
}

func (ui *DebuggerUI) cmdList(args []string) {
	state := ui.debugger.GetState()

	file := state.CurrentFile
	line := state.CurrentLine

	if len(args) > 0 {
		// Parse location argument
		if parts := strings.Split(args[0], ":"); len(parts) == 2 {
			file = parts[0]
			if l, err := strconv.Atoi(parts[1]); err == nil {
				line = l
			}
		}
	}

	if file == "" {
		fmt.Fprintln(ui.writer, "No source location available")
		return
	}

	ui.showSource(file, line, 10)
}

func (ui *DebuggerUI) cmdPrint(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: print <variable>")
		return
	}

	name := args[0]

	// Special registers
	state := ui.debugger.GetState()
	switch name {
	case "pc", "PC":
		fmt.Fprintf(ui.writer, "$pc = 0x%04x\n", state.PC)
		return
	case "sp", "SP":
		fmt.Fprintf(ui.writer, "$sp = 0x%04x\n", state.SP)
		return
	case "fp", "FP":
		fmt.Fprintf(ui.writer, "$fp = 0x%04x\n", state.FP)
		return
	}

	// Try as variable
	value, err := ui.debugger.GetVariable(name)
	if err != nil {
		fmt.Fprintf(ui.writer, "Error: %v\n", err)
		return
	}

	fmt.Fprintf(ui.writer, "%s = 0x%04x (%d)\n", name, value, value)
}

func (ui *DebuggerUI) cmdExamine(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: x <address> [count]")
		return
	}

	addr, err := ui.parseAddress(args[0])
	if err != nil {
		fmt.Fprintf(ui.writer, "Invalid address: %s\n", args[0])
		return
	}

	count := 16
	if len(args) > 1 {
		if c, err := strconv.Atoi(args[1]); err == nil {
			count = c
		}
	}

	ui.showMemory(addr, count)
}

func (ui *DebuggerUI) cmdInfo(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: info <regs|break|stack|watches>")
		return
	}

	switch args[0] {
	case "regs", "registers":
		ui.showRegisters()
	case "break", "breakpoints":
		ui.listBreakpoints()
	case "stack":
		ui.showStack()
	case "watches", "watch":
		ui.listWatches()
	default:
		fmt.Fprintf(ui.writer, "Unknown info command: %s\n", args[0])
	}
}

func (ui *DebuggerUI) cmdDisasm(args []string) {
	state := ui.debugger.GetState()
	addr := state.PC

	if len(args) > 0 {
		if a, err := ui.parseAddress(args[0]); err == nil {
			addr = a
		}
	}

	ui.disassemble(addr, 10)
}

func (ui *DebuggerUI) cmdWatch(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: watch <variable>")
		return
	}

	name := args[0]
	ui.debugger.AddWatch(name)

	// Try to get current value
	value, err := ui.debugger.GetVariable(name)
	if err != nil {
		fmt.Fprintf(ui.writer, "Added watch on '%s' (currently unavailable: %v)\n", name, err)
	} else {
		fmt.Fprintf(ui.writer, "Added watch on '%s' = 0x%04x (%d)\n", name, value, value)
	}
}

func (ui *DebuggerUI) cmdUnwatch(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(ui.writer, "Usage: unwatch <variable>")
		return
	}

	name := args[0]
	ui.debugger.RemoveWatch(name)
	fmt.Fprintf(ui.writer, "Removed watch on '%s'\n", name)
}

// Helper methods

func (ui *DebuggerUI) handleBreakpoint(bp *Breakpoint) {
	fmt.Fprintf(ui.writer, "\nBreakpoint %d hit at 0x%04x", bp.ID, bp.Address)
	if bp.File != "" {
		fmt.Fprintf(ui.writer, " (%s:%d)", filepath.Base(bp.File), bp.Line)
	}
	fmt.Fprintln(ui.writer)
	ui.showCurrentLocation()
}

func (ui *DebuggerUI) handleStep(pc uint16) {
	ui.showCurrentLocation()
}

func (ui *DebuggerUI) showCurrentLocation() {
	state := ui.debugger.GetState()

	// Show registers on one line
	ui.showRegistersCompact()

	// Show current instruction
	fmt.Fprintf(ui.writer, "PC: 0x%04x", state.PC)
	if state.CurrentFile != "" {
		fmt.Fprintf(ui.writer, " (%s:%d)", filepath.Base(state.CurrentFile), state.CurrentLine)
	}
	fmt.Fprintln(ui.writer)

	// Show source if available
	if state.CurrentFile != "" {
		ui.showSource(state.CurrentFile, state.CurrentLine, 5)
	} else {
		// Show disassembly
		ui.disassemble(state.PC, 1)
	}
}

func (ui *DebuggerUI) showSource(file string, centerLine int, context int) {
	lines, startLine, err := ui.sourceReader.GetContext(file, centerLine, context, context)
	if err != nil {
		fmt.Fprintf(ui.writer, "Cannot read source file: %v\n", err)
		return
	}

	for i, line := range lines {
		lineNum := startLine + i
		marker := "  "
		if lineNum == centerLine {
			marker = "=>"
		}

		// Try to find address for this line
		addr, hasAddr := ui.debugger.FindAddressForLine(file, lineNum)
		if hasAddr {
			fmt.Fprintf(ui.writer, "%s %04x:%4d  %s\n", marker, addr, lineNum, line)
		} else {
			fmt.Fprintf(ui.writer, "%s      %4d  %s\n", marker, lineNum, line)
		}
	}
}

func (ui *DebuggerUI) showMemory(addr uint16, count int) {
	data := ui.debugger.ReadMemory(addr, count)

	for i := 0; i < len(data); i += 16 {
		fmt.Fprintf(ui.writer, "0x%04x: ", addr+uint16(i))

		// Hex bytes
		for j := 0; j < 16 && i+j < len(data); j++ {
			fmt.Fprintf(ui.writer, "%02x ", data[i+j])
		}

		// Padding
		for j := len(data) - i; j < 16; j++ {
			fmt.Fprint(ui.writer, "   ")
		}

		fmt.Fprint(ui.writer, " |")

		// ASCII
		for j := 0; j < 16 && i+j < len(data); j++ {
			ch := data[i+j]
			if ch >= 32 && ch <= 126 {
				fmt.Fprintf(ui.writer, "%c", ch)
			} else {
				fmt.Fprint(ui.writer, ".")
			}
		}

		fmt.Fprintln(ui.writer, "|")
	}
}

func (ui *DebuggerUI) showRegisters() {
	state := ui.debugger.GetState()

	fmt.Fprintf(ui.writer, "PC: 0x%04x  SP: 0x%04x  FP: 0x%04x\n",
		state.PC, state.SP, state.FP)
	fmt.Fprintf(ui.writer, "Flags: ")
	if state.Negative {
		fmt.Fprint(ui.writer, "N ")
	}
	if state.Zero {
		fmt.Fprint(ui.writer, "Z ")
	}
	if state.Carry {
		fmt.Fprint(ui.writer, "C ")
	}
	if state.Bytes {
		fmt.Fprint(ui.writer, "B ")
	}
	fmt.Fprintln(ui.writer)
}

func (ui *DebuggerUI) showRegistersCompact() {
	state := ui.debugger.GetState()

	// Format flags as zncb where lowercase=clear, uppercase=set
	flags := ""
	if state.Zero {
		flags += "Z"
	} else {
		flags += "z"
	}
	if state.Negative {
		flags += "N"
	} else {
		flags += "n"
	}
	if state.Carry {
		flags += "C"
	} else {
		flags += "c"
	}
	if state.Bytes {
		flags += "B"
	} else {
		flags += "b"
	}

	fmt.Fprintf(ui.writer, "[PC=%04x SP=%04x FP=%04x Flags=%s]",
		state.PC, state.SP, state.FP, flags)

	// Show watches on the same line
	watches := ui.debugger.GetWatchValues()
	if len(watches) > 0 {
		fmt.Fprint(ui.writer, " ")
		first := true
		for name, value := range watches {
			if !first {
				fmt.Fprint(ui.writer, ", ")
			}
			fmt.Fprintf(ui.writer, "%s: %s", name, value)
			first = false
		}
	}
	fmt.Fprintln(ui.writer)
}

func (ui *DebuggerUI) listBreakpoints() {
	bps := ui.debugger.ListBreakpoints()
	if len(bps) == 0 {
		fmt.Fprintln(ui.writer, "No breakpoints set")
		return
	}

	fmt.Fprintln(ui.writer, "Breakpoints:")
	for _, bp := range bps {
		status := "enabled"
		if !bp.Enabled {
			status = "disabled"
		}

		fmt.Fprintf(ui.writer, "  %d: 0x%04x", bp.ID, bp.Address)
		if bp.File != "" {
			fmt.Fprintf(ui.writer, " (%s:%d)", filepath.Base(bp.File), bp.Line)
		}
		fmt.Fprintf(ui.writer, " [%s, hit %d times]\n", status, bp.HitCount)
	}
}

func (ui *DebuggerUI) showStack() {
	// TODO: Implement call stack display
	fmt.Fprintln(ui.writer, "Call stack not yet implemented")
}

func (ui *DebuggerUI) listWatches() {
	watches := ui.debugger.ListWatches()
	if len(watches) == 0 {
		fmt.Fprintln(ui.writer, "No watches set")
		return
	}

	fmt.Fprintln(ui.writer, "Watches:")
	values := ui.debugger.GetWatchValues()
	for _, name := range watches {
		if value, ok := values[name]; ok {
			fmt.Fprintf(ui.writer, "  %s = %s\n", name, value)
		}
	}
}

func (ui *DebuggerUI) disassemble(addr uint16, count int) {
	currentAddr := addr
	for i := 0; i < count; i++ {
		inst := ui.debugger.memory.GetByte(currentAddr)
		op, m1, m2 := machine.DecodeOp(inst)

		fmt.Fprintf(ui.writer, "0x%04x:  %02x ", currentAddr, inst)

		instSize := 1
		args := ""

		if m1 != machine.Implied && m2 != machine.Implied {
			op1, n1 := ui.formatOperand(m1, currentAddr+1)
			instSize += n1
			op2, n2 := ui.formatOperand(m2, currentAddr+1+uint16(n1))
			instSize += n2
			args = op1 + "," + op2
		} else if m1 != machine.Implied && m2 == machine.Implied {
			op1, n := ui.formatOperand(m1, currentAddr+1)
			instSize += n
			args = op1
		}

		// Print hex bytes
		for j := 1; j < instSize && j < 5; j++ {
			fmt.Fprintf(ui.writer, "%02x ", ui.debugger.memory.GetByte(currentAddr+uint16(j)))
		}
		for j := instSize; j < 5; j++ {
			fmt.Fprint(ui.writer, "   ")
		}

		fmt.Fprintf(ui.writer, "%s %s\n", op, args)
		currentAddr += uint16(instSize)
	}
}

func (ui *DebuggerUI) formatOperand(mode machine.AddressMode, pc uint16) (string, int) {
	var op string
	var bytes int

	switch mode {
	case machine.Implied:
		// nothing to do
	case machine.Immediate:
		op = fmt.Sprintf("#0x%04x", ui.debugger.memory.GetWord(pc))
		bytes = 2
	case machine.ImmediateByte:
		op = fmt.Sprintf("#0x%02x", int8(ui.debugger.memory.GetByte(pc)))
		bytes = 1
	case machine.OffsetByte:
		value := int8(ui.debugger.memory.GetByte(pc))
		op = fmt.Sprintf("0x%04x (%+d)", int(pc-1)+int(value), value)
		bytes = 1
	case machine.Absolute:
		op = fmt.Sprintf("0x%04x", ui.debugger.memory.GetWord(pc))
		bytes = 2
	case machine.Indirect:
		op = fmt.Sprintf("*0x%04x", ui.debugger.memory.GetWord(pc))
		bytes = 2
	case machine.Relative:
		value := int8(ui.debugger.memory.GetByte(pc))
		op = fmt.Sprintf("fp%+d", value)
		bytes = 1
	case machine.RelativeIndirect:
		value := int8(ui.debugger.memory.GetByte(pc))
		op = fmt.Sprintf("*fp%+d", value)
		bytes = 1
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	return op, bytes
}

func (ui *DebuggerUI) parseAddress(s string) (uint16, error) {
	if strings.HasPrefix(s, "0x") {
		addr, err := strconv.ParseUint(s[2:], 16, 16)
		return uint16(addr), err
	}

	addr, err := strconv.ParseUint(s, 10, 16)
	return uint16(addr), err
}

// runBasicMode runs the debugger without readline support
func (ui *DebuggerUI) runBasicMode() {
	ui.printWelcome()

	// Set up callbacks
	ui.debugger.onBreakpoint = ui.handleBreakpoint
	ui.debugger.onStep = ui.handleStep

	// Show initial state
	ui.showCurrentLocation()

	// Main command loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(ui.writer, "(mpu-dbg) ")

		if !scanner.Scan() {
			fmt.Fprintln(ui.writer, "\nExiting debugger.")
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !ui.handleCommand(line) {
			break
		}
	}
}
