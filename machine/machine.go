package machine

import (
	"fmt"
	"io"
	"os"
)

const (
	PCAddr = 0 // Program counter
	SPAddr = 2 // Stack pointer, points to last byte written
	FPAddr = 4 // Frame pointer
)

// Machine implements MPU ... memory processing unit.
// It supports 27 instructions and 6 addressing modes.
type Machine struct {
	memory   []byte // 64kb of memory
	pc       uint16 // program counter ... shadowed on read/write to address $0
	sp       uint16 // stack pointer ... shadowed on read/write to address $2
	fp       uint16 // frame pointer ... shadowed on read/write to address $4
	negative bool   // Negative flag, set true if last value had the high bit set.
	zero     bool   // Zero flag, set true if last value had zero value.
	carry    bool   // Carry flag
	bytes    bool   // Bytes flag, if true then operations are on bytes instead of words
	step     bool   // Single step mode, if true breaks after executing 1 instruction
}

func NewMachine(image []byte) *Machine {
	m := &Machine{
		memory: make([]byte, 65536),
		pc:     0x100,
		sp:     0xffff,
	}
	copy(m.memory, image)
	m.pc = m.readUint16(PCAddr)
	m.sp = m.readUint16(SPAddr)
	m.fp = m.readUint16(FPAddr)
	return m
}

func (m *Machine) readUint16(addr uint16) uint16 {
	return uint16(m.memory[addr+1])<<8 + uint16(m.memory[addr])
}

func (m *Machine) ReadInt8(addr int) int {
	return int(int8(m.memory[addr]))
}

func (m *Machine) ReadWord(addr int) int {
	if addr == PCAddr {
		return int(m.pc)
	}
	if addr == SPAddr {
		return int(m.sp)
	}
	if addr == FPAddr {
		return int(m.fp)
	}
	return int(m.memory[addr+1])<<8 + int(m.memory[addr])
}

func (m *Machine) WriteWord(addr, value int) {
	if addr == PCAddr {
		m.pc = uint16(value)
	} else if addr == SPAddr {
		m.sp = uint16(value)
	} else if addr == FPAddr {
		m.fp = uint16(value)
	} else {
		lo := byte(value & 0xff)
		hi := byte(value >> 8 & 0xff)
		m.memory[addr] = lo
		m.memory[addr+1] = hi
	}
}

// Same as WriteWord but updates zero/minus flags
func (m *Machine) writeTarget(addr, value int) {
	if m.bytes {
		if addr == PCAddr || addr == SPAddr || addr == FPAddr {
			panic("attempt to write byte value to pc/sp/fp")
		}
		m.memory[addr] = byte(value)
		m.updateFlagsByte(value)
	} else {
		m.WriteWord(addr, value)
		m.updateFlagsWord(value)
	}
	//fmt.Printf("  0x%0x <- 0x%0x [z:%t, n:%t]\n", addr, value, m.zero, m.negative)
}

func (m *Machine) updateFlagsByte(value int) {
	m.negative = value&0x80 != 0
	m.zero = value == 0
}

func (m *Machine) updateFlagsWord(value int) {
	m.negative = value&0x8000 != 0
	m.zero = value == 0
}

func (m *Machine) Run() {
	for {
		in := m.memory[m.pc]
		if in == 0 {
			return
		}
		target := 0 // the address being updated, ie often the address of value1
		value1 := 0
		value2 := 0
		opCode, m1, m2 := DecodeOp(in)
		bytes := 0
		if opCode < Inc {
			var b1, b2 int
			target, value1, b1 = m.fetchOperand(m1, int(m.pc+1))
			_, value2, b2 = m.fetchOperand(m2, int(m.pc+3))
			bytes = b1 + b2
		} else if opCode <= Jsr {
			var b1 int
			target, value1, b1 = m.fetchOperand(m1, int(m.pc+1))
			bytes = b1
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
			m.writeTarget(int(m.sp), value1)
			if m.bytes {
				m.sp -= 1
			} else {
				m.sp -= 2
			}
		case Pop:
			if m1 == Immediate {
				m.sp += uint16(value1)
			} else {
				if m.bytes {
					m.sp += 1
				} else {
					m.sp += 2
				}
				m.writeTarget(target, m.ReadWord(int(m.sp)))
			}
		case Jmp:
			m.pc = uint16(value1)
		case Jeq:
			if m.zero {
				m.pc = offset(m.pc, value1)
			}
		case Jne:
			if !m.zero {
				m.pc = offset(m.pc, value1)
			}
		case Jge:
			if !m.negative {
				m.pc = offset(m.pc, value1)
			}
		case Jlt:
			if m.negative {
				m.pc = offset(m.pc, value1)
			}
		case Jcs:
			if m.carry {
				m.pc = offset(m.pc, value1)
			}
		case Jcc:
			if !m.carry {
				m.pc = offset(m.pc, value1)
			}
		case Jsr:
			m.pushWord(int(m.pc))
			m.pc = uint16(value1)
		case Ret:
			m.pc = m.popUint16()
		case Sav:
			// push frame pointer, copy sp->fp, adjust stack for locals
			m.pushWord(int(m.fp))
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

func (m *Machine) pushWord(value int) {
	m.writeTarget(int(m.sp), value)
	m.sp -= 2
}

func (m *Machine) popWord() int {
	m.sp += 2
	return m.ReadWord(int(m.sp))
}

func (m *Machine) popUint16() uint16 {
	m.sp += 2
	return m.readUint16(m.sp)
}

func offset(addr uint16, offset int) uint16 {
	return uint16(int(addr) + offset)
}

func (m *Machine) fetchOperand(mode AddressMode, pc int) (address, value, bytes int) {
	switch mode {
	case Implied:
		// nothing to do
	case Immediate:
		address = pc
		value = m.ReadWord(pc)
		bytes = 2
		return
	case ImmediateByte:
		address = pc
		value = m.ReadInt8(pc)
		bytes = 1
		return
	case OffsetByte:
		address = pc
		value = address + m.ReadInt8(pc)
		bytes = 1
		return
	case Absolute:
		address = m.ReadWord(pc)
		bytes = 2
	case Indirect:
		address = m.ReadWord(m.ReadWord(pc))
		bytes = 2
	case Relative:
		address = m.ReadInt8(pc) + int(m.fp)
		bytes = 1
	case RelativeIndirect:
		address = m.ReadWord(m.ReadInt8(pc) + int(m.fp))
		bytes = 1
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	if m.bytes {
		value = int(m.memory[address])
	} else {
		value = m.ReadWord(address)
	}
	return
}

func (m *Machine) Dump(w io.Writer, start int, end int) {
	ascii := make([]byte, 16)
	charIndex := 0
	flush := func() {
		for charIndex < 16 {
			ascii[charIndex] = ' '
			charIndex++
			fmt.Fprintf(w, "   ")
		}
		fmt.Fprintf(w, " |%s|\n", string(ascii))
		charIndex = 0
	}
	for addr := start; addr <= end; addr++ {
		if addr == start || charIndex == 16 {
			if addr != start {
				flush()
			}
			fmt.Fprintf(w, "%04x  ", addr)
		}
		fmt.Fprintf(w, "%02x ", m.memory[addr])
		ch := m.memory[addr]
		if ch >= 32 && ch <= 126 {
			ascii[charIndex] = ch
		} else {
			ascii[charIndex] = '.'
		}
		charIndex++
	}
	flush()
}

// List will disassemble n instructions starting at addr, and return the
// pc location following the last instruction.
func (m *Machine) List(w io.Writer, addr int, n int) int {
	for i := 0; i < n; i++ {
		in := m.memory[addr]
		op, m1, m2 := DecodeOp(in)
		var value1 uint16
		var value2 uint16
		var opCount int
		if op > Hlt && op < Inc {
			opCount = 2
			value1 = m.readUint16(uint16(addr + 1))
			value2 = m.readUint16(uint16(addr + 3))
			fmt.Fprintf(w, "0x%04x  %02x %02x %02x %02x %02x  %s %s,%s\n",
				addr, m.memory[addr], m.memory[addr+1], m.memory[addr+2], m.memory[addr+3], m.memory[addr+4],
				op, formatArg(m1, value1), formatArg(m2, value2))
		} else if op > Hlt && op <= Jsr {
			opCount = 1
			value1 = m.readUint16(uint16(addr + 1))
			fmt.Fprintf(w, "0x%04x  %02x %02x %02x        %s %s\n",
				addr, m.memory[addr], m.memory[addr+1], m.memory[addr+2],
				op, formatArg(m1, value1))
		} else {
			opCount = 0
			fmt.Fprintf(w, "0x%04x  %02x              %s\n",
				addr, m.memory[addr], op)
		}

		addr += 2*opCount + 1
	}
	return addr
}

func (m *Machine) RunAt(addr int) {
	m.step = false
	m.pc = uint16(addr)
	m.Run()
}

func (m *Machine) Step(addr int) int {
	m.step = true
	m.pc = uint16(addr)
	m.List(os.Stdout, int(m.pc), 1)
	m.Run()
	m.step = false
	fmt.Printf("[status pc=%04x sp=%04x n=%d z=%d c=%d b=%d]\n", m.pc, m.sp, boolInt(m.negative), boolInt(m.zero),
		boolInt(m.carry), boolInt(m.bytes))
	return int(m.pc)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatArg(mode AddressMode, value uint16) string {
	switch mode {
	case Immediate:
		return fmt.Sprintf("#0x%04x", value)
	case Absolute:
		return fmt.Sprintf("0x%04x", value)
	case Indirect:
		return fmt.Sprintf("*0x%04x", value)
	case Relative:
		return fmt.Sprintf("fp+%d", value)
	case RelativeIndirect:
		return fmt.Sprintf("*fp+%d", value)
	}
	return "none"
}
