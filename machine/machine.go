package machine

import (
	"fmt"
	"time"
)

const (
	PCAddr     = 0  // Program counter
	SPAddr     = 2  // Stack pointer, points to last byte written
	FPAddr     = 4  // Frame pointer
	IOReqAddr  = 6  // Address of I/O commands are written here to execute
	IOStatAddr = 8  // I/O status of last command, 0 = success, != 0 error
	RandAddr   = 10 // Writes are ignored, reads return random uint8/uint16
)

// Machine implements MPU ... memory processing unit.
// It supports 27 instructions and 6 addressing modes.
type Machine struct {
	memory   Memory // 64kb of memory + dma overlay
	pc       uint16 // program counter ... shadowed on read/write to address $0
	sp       uint16 // stack pointer ... shadowed on read/write to address $2
	fp       uint16 // frame pointer ... shadowed on read/write to address $4
	negative bool   // Negative flag, set true if last value had the high bit set.
	zero     bool   // Zero flag, set true if last value had zero value.
	carry    bool   // Carry flag
	bytes    bool   // Bytes flag, if true then operations are on bytes instead of words
	step     bool   // Single step mode, if true breaks after executing 1 instruction
}

func NewMachineWithDevices(d *IODispatcher, image []byte) *Machine {
	// Hoist up initial register values from raw image, if provided.
	// This could be moved into the ByteSliceMemory constructor, except
	// not all Memory objects support this ... writes to the IODispatcher
	// will trigger an IO request.
	m := &Machine{
		pc: readOrDefault(image, PCAddr, 0x100),
		sp: readOrDefault(image, SPAddr, 0xffff),
		fp: readOrDefault(image, FPAddr, 0),
	}
	var unused1 uint16
	var unused2 uint16
	memory := NewByteSliceMemory(
		[]Memory{
			&Register{value: &m.pc},
			&Register{value: &m.sp},
			&Register{value: &m.fp},
			d,
			d.StatusRegister(),
			NewRNG(time.Now().Unix()),
			&Register{&unused1}, // Currently, 2 unused registers
			&Register{&unused2},
		},
		image,
	)
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
		in := m.memory.GetByte(m.pc)
		if in == 0 {
			return
		}
		var n uint16      // Number of bytes for each operand
		var bytes uint16  // Total count of operand bytes (to skip pc to next instruction)
		var target uint16 // the address being updated, ie often the address of value1
		value1 := 0       // value of first operand, if any
		value2 := 0       // value of second operand, if any
		opCode, m1, m2 := DecodeOp(in)
		if m1 != Implied {
			target, value1, n = m.fetchOperand(m1, m.pc+1)
			bytes = n
		}
		if m2 != Implied {
			_, value2, n = m.fetchOperand(m2, m.pc+1+bytes)
			bytes += n
		}
		m.pc = m.pc + uint16(bytes) + 1

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
				m.sp -= 1
			} else {
				m.sp -= 2
			}
			m.writeTarget(m.sp, value1)
		case Pop:
			if m1 == ImmediateByte {
				m.sp += uint16(value1)
			} else {
				m.writeTarget(target, int(m.memory.GetWord(m.sp)))
				if m.bytes {
					m.sp += 1
				} else {
					m.sp += 2
				}
			}
		case Jmp:
			m.pc = uint16(value1)
		case Jeq:
			if m.zero {
				m.pc = uint16(value1)
			}
		case Jne:
			if !m.zero {
				m.pc = uint16(value1)
			}
		case Jge:
			if !m.negative {
				m.pc = uint16(value1)
			}
		case Jlt:
			if m.negative {
				m.pc = uint16(value1)
			}
		case Jcs:
			if m.carry {
				m.pc = uint16(value1)
			}
		case Jcc:
			if !m.carry {
				m.pc = uint16(value1)
			}
		case Jsr:
			m.pushUint16(m.pc)
			m.pc = uint16(value1)
		case Ret:
			m.pc = m.popUint16()
		case Sav:
			// push frame pointer, copy sp->fp, adjust stack for locals
			m.pushUint16(m.fp)
			m.fp = m.sp
			m.sp -= uint16(value1)
		case Rst:
			// discard locals from stack, restore fp, ret
			m.sp = m.fp
			m.fp = m.popUint16()
			m.pc = m.popUint16()
		case Seb:
			m.bytes = true
		case Clb:
			m.bytes = false
		case Sec:
			m.carry = true
		case Clc:
			m.carry = false
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
	m.sp -= 2
	m.writeTarget(m.sp, value)
}

func (m *Machine) popWord() int {
	word := m.memory.GetWord(m.sp)
	m.sp += 2
	return int(word)
}

func (m *Machine) pushUint16(w uint16) {
	m.sp -= 2
	m.memory.PutWord(m.sp, w)
}

func (m *Machine) popUint16() uint16 {
	val := m.memory.GetWord(m.sp)
	m.sp += 2
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
		address = uint16(m.ReadInt8(pc) + int(m.fp))
		bytes = 1
	case RelativeIndirect:
		address = m.memory.GetWord(uint16((m.ReadInt8(pc) + int(m.fp))))
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
	m.pc = pc
	m.Run()
}

// Step executes a single instruction at the given address, and returns the new PC address.
func (m *Machine) Step(addr uint16) uint16 {
	m.step = true
	m.pc = addr
	m.Run()
	return m.pc
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

// Flags returns a snapshot of the current state of the registers and flags.
func (m *Machine) Flags() Flags {
	return Flags{
		PC:       m.pc,
		SP:       m.sp,
		FP:       m.fp,
		Negative: m.negative,
		Zero:     m.zero,
		Carry:    m.carry,
		Bytes:    m.bytes,
	}
}
