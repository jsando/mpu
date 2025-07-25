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

package machine

import (
	"fmt"
)

// BaseDirEnv is the key for an environment variable to use for loading relative files.
const BaseDirEnv = "MPU_BASE_DIR"

// AssertionFailure records details of a failed assertion
type AssertionFailure struct {
	PC       uint16
	Expected int
	Actual   int
}

// Machine implements MPU ... memory processing unit.
// It supports 27 instructions and 6 addressing modes.
type Machine struct {
	memory         Memory            // 64kb of memory
	dispatcher     *IODispatcher     // IO dispatcher for device requests
	Pc             uint16            // program counter
	Sp             uint16            // stack pointer
	Fp             uint16            // frame pointer
	negative       bool              // Negative flag, set true if last value had the high bit set.
	zero           bool              // Zero flag, set true if last value had zero value.
	carry          bool              // Carry flag
	bytes          bool              // Bytes flag, if true then operations are on bytes instead of words
	step           bool              // Single step mode, if true breaks after executing 1 instruction
	assertion      bool              // Assertion flag, set by SEA instruction, affects next CMP
	testMode       bool              // Test mode, enables assertion checking
	assertionFails int               // Count of assertion failures
	lastFailure    *AssertionFailure // Details of the last assertion failure
}

func NewMachineWithDevices(d *IODispatcher, image []byte) *Machine {
	// Initialize machine with default register values
	m := &Machine{
		dispatcher: d,
		Pc:         0,      // Programs start at address 0
		Sp:         0xffff, // Stack starts at top of memory
		Fp:         0,      // Frame pointer starts at 0
	}
	// Create simple byte slice memory without any mapped registers
	memory := NewByteSliceMemory(nil, image)
	m.memory = memory
	d.memory = memory
	return m
}

func NewMachine(image []byte) *Machine {
	ioRequest := NewDefaultDispatcher()
	return NewMachineWithDevices(ioRequest, image)
}

func (m *Machine) Memory() Memory {
	return m.memory
}

func (m *Machine) Run() {
	for {
		in := m.memory.GetByte(m.Pc)
		var n uint16      // Number of bytes for each operand
		var bytes uint16  // Total count of operand bytes (to skip pc to next instruction)
		var target uint16 // the address being updated, ie often the address of value1
		value1 := 0       // value of first operand, if any
		value2 := 0       // value of second operand, if any
		opCode, m1, m2 := DecodeOp(in)

		// Check for halt after decoding
		if opCode == Hlt {
			return
		}
		if m1 != Implied {
			target, value1, n = m.fetchOperand(m1, m.Pc+1)
			bytes = n
		}
		if m2 != Implied {
			_, value2, n = m.fetchOperand(m2, m.Pc+1+bytes)
			bytes += n
		}
		m.Pc = m.Pc + uint16(bytes) + 1

		switch opCode {
		case Add: // todo carry
			m.writeTarget(target, value1+value2)
		case Sub: // todo carry
			m.writeTarget(target, value1-value2)
		case Mul:
			m.writeTarget(target, value1*value2)
		case Div:
			m.writeTarget(target, value1/value2)
		case Cmp:
			m.updateFlagsWord(value1 - value2)
			// Handle assertion if flag is set
			if m.assertion {
				m.assertion = false // Clear flag
				if m.testMode && value1 != value2 {
					m.assertionFails++
					// Record failure details - PC points to next instruction, so subtract bytes+1
					m.lastFailure = &AssertionFailure{
						PC:       m.Pc - uint16(bytes) - 1,
						Expected: value2,
						Actual:   value1,
					}
				}
			}
		case And:
			m.writeTarget(target, value1&value2)
		case Or:
			m.writeTarget(target, value1|value2)
		case Xor:
			m.writeTarget(target, value1^value2)
		case Cpy:
			m.writeTarget(target, value2)
		case Inc:
			m.writeTarget(target, value1+1)
		case Dec:
			m.writeTarget(target, value1-1)
		case Psh:
			if m.bytes {
				m.Sp -= 1
			} else {
				m.Sp -= 2
			}
			m.writeTarget(m.Sp, value1)
		case Pop:
			if m1 == ImmediateByte {
				m.Sp += uint16(value1)
			} else {
				m.writeTarget(target, int(m.memory.GetWord(m.Sp)))
				if m.bytes {
					m.Sp += 1
				} else {
					m.Sp += 2
				}
			}
		case Jmp:
			m.Pc = uint16(value1)
		case Jeq:
			if m.zero {
				m.Pc = uint16(value1)
			}
		case Jne:
			if !m.zero {
				m.Pc = uint16(value1)
			}
		case Jge:
			if !m.negative {
				m.Pc = uint16(value1)
			}
		case Jlt:
			if m.negative {
				m.Pc = uint16(value1)
			}
		case Jcs:
			if m.carry {
				m.Pc = uint16(value1)
			}
		case Jcc:
			if !m.carry {
				m.Pc = uint16(value1)
			}
		case Jsr:
			m.pushUint16(m.Pc)
			m.Pc = uint16(value1)
		case Ret:
			m.Pc = m.popUint16()
		case Sav:
			// push frame pointer, copy sp->fp, adjust stack for locals
			m.pushUint16(m.Fp)
			m.Fp = m.Sp
			m.Sp -= uint16(value1)
		case Rst:
			// discard locals from stack, restore fp, ret
			m.Sp = m.Fp
			m.Fp = m.popUint16()
			m.Pc = m.popUint16()
		case Seb:
			m.bytes = true
		case Clb:
			m.bytes = false
		case Sec:
			m.carry = true
		case Clc:
			m.carry = false
		case Sea:
			m.assertion = true
		case Req:
			// Execute request.  The operand is the address of the request structure.
			if m.dispatcher != nil {
				m.dispatcher.execIORequest(target)
			}
		}

		if m.step {
			break
		}
	}
}

// ReadInt8 reads the given addr from memory as a byte and casts it to a signed int8 (as an int).
func (m *Machine) ReadInt8(addr uint16) int {
	return int(int8(m.memory.GetByte(addr)))
}

// writeTarget writes the new value to the given addr, and updates the various MPU flags
// such as zero and negative.
func (m *Machine) writeTarget(addr uint16, value int) {
	if m.bytes {
		m.memory.PutByte(addr, byte(value))
		m.updateFlagsByte(value)
		//fmt.Printf("  0x%04x <- 0x%02x [z:%t, n:%t]\n", addr, value, m.zero, m.negative)
	} else {
		m.memory.PutWord(addr, uint16(value))
		m.updateFlagsWord(value)
		//fmt.Printf("  0x%04x <- 0x%04x [z:%t, n:%t]\n", addr, value, m.zero, m.negative)
	}
}

func (m *Machine) updateFlagsByte(value int) {
	m.negative = value&0x80 != 0
	m.zero = value == 0
}

func (m *Machine) updateFlagsWord(value int) {
	m.negative = value&0x8000 != 0
	m.zero = value == 0
}

func (m *Machine) pushWord(value int) {
	m.Sp -= 2
	m.writeTarget(m.Sp, value)
}

func (m *Machine) popWord() int {
	word := m.memory.GetWord(m.Sp)
	m.Sp += 2
	return int(word)
}

func (m *Machine) pushUint16(w uint16) {
	m.Sp -= 2
	m.memory.PutWord(m.Sp, w)
}

func (m *Machine) popUint16() uint16 {
	val := m.memory.GetWord(m.Sp)
	m.Sp += 2
	return val
}

func (m *Machine) fetchOperand(mode AddressMode, pc uint16) (address uint16, value int, bytes uint16) {
	switch mode {
	case Implied:
		// nothing to do
	case Immediate:
		address = pc
		value = int(m.memory.GetWord(pc))
		bytes = 2
		return
	case ImmediateByte:
		address = pc
		value = m.ReadInt8(pc)
		bytes = 1
		return
	case OffsetByte:
		address = pc
		value = int(address) - 1 + m.ReadInt8(pc)
		bytes = 1
		return
	case Absolute:
		address = m.memory.GetWord(pc)
		bytes = 2
	case Indirect:
		address = m.memory.GetWord(m.memory.GetWord((pc)))
		bytes = 2
	case Relative:
		address = uint16(m.ReadInt8(pc) + int(m.Fp))
		bytes = 1
	case RelativeIndirect:
		address = m.memory.GetWord(uint16((m.ReadInt8(pc) + int(m.Fp))))
		bytes = 1
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	if m.bytes {
		value = int(m.memory.GetByte(address))
	} else {
		value = int(m.memory.GetWord(address))
	}
	return
}

// RunAt runs code from the given program counter until a HLT is encountered.
func (m *Machine) RunAt(pc uint16) {
	m.step = false
	m.Pc = pc
	m.Run()
}

// Step executes a single instruction at the given address, and returns the new PC address.
func (m *Machine) Step(addr uint16) uint16 {
	m.step = true
	m.Pc = addr
	m.Run()
	return m.Pc
}

type Flags struct {
	PC       uint16
	SP       uint16
	FP       uint16
	Negative bool
	Zero     bool
	Carry    bool
	Bytes    bool
}

// EnableTestMode enables assertion checking for unit tests
func (m *Machine) EnableTestMode() {
	m.testMode = true
}

// AssertionFailures returns the count of assertion failures
func (m *Machine) AssertionFailures() int {
	return m.assertionFails
}

// LastAssertionFailure returns details of the last assertion failure
func (m *Machine) LastAssertionFailure() *AssertionFailure {
	return m.lastFailure
}

// Flags returns a snapshot of the current state of the registers and flags.
func (m *Machine) Flags() Flags {
	return Flags{
		PC:       m.Pc,
		SP:       m.Sp,
		FP:       m.Fp,
		Negative: m.negative,
		Zero:     m.zero,
		Carry:    m.carry,
		Bytes:    m.bytes,
	}
}
