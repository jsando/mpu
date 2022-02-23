package machine

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

const (
	PCAddr   = 0  // Program counter
	SPAddr   = 2  // Stack pointer, points to last byte written
	FPAddr   = 4  // Frame pointer
	IOReq    = 6  // Address of I/O commands are written here to execute
	IORes    = 8  // I/O status of last command, 0 = success, != 0 error
	RandAddr = 10 // Writes are ignored, reads return random uint8/uint16
)

// Machine implements MPU ... memory processing unit.
// It supports 27 instructions and 6 addressing modes.
type Machine struct {
	memory   []byte // 64kb of memory
	pc       uint16 // program counter ... shadowed on read/write to address $0
	sp       uint16 // stack pointer ... shadowed on read/write to address $2
	fp       uint16 // frame pointer ... shadowed on read/write to address $4
	rgen     *rand.Rand
	negative bool // Negative flag, set true if last value had the high bit set.
	zero     bool // Zero flag, set true if last value had zero value.
	carry    bool // Carry flag
	bytes    bool // Bytes flag, if true then operations are on bytes instead of words
	step     bool // Single step mode, if true breaks after executing 1 instruction
	devices  []IODevice
}

type IODevice interface {
	Invoke(m *Machine, addr uint16) (err uint16)
}

func NewMachine(image []byte) *Machine {
	m := &Machine{
		memory:  make([]byte, 65536),
		pc:      0x100,
		sp:      0xffff,
		rgen:    rand.New(rand.NewSource(time.Now().UnixNano())),
		devices: make([]IODevice, 256),
	}
	copy(m.memory, image)
	m.pc = m.readUint16(PCAddr)
	m.sp = m.readUint16(SPAddr)
	m.fp = m.readUint16(FPAddr)
	m.RegisterDevice(&StdoutDevice{}, 1)
	m.RegisterDevice(NewSDLDevice(m), 2)
	return m
}

func (m *Machine) RegisterDevice(device IODevice, deviceId int) {
	m.devices[deviceId] = device
}

func (m *Machine) readUint16(addr uint16) uint16 {
	return uint16(m.memory[addr+1])<<8 + uint16(m.memory[addr])
}

func (m *Machine) writeUint16(addr uint16, val uint16) {
	m.memory[addr] = byte(val & 0xff)
	m.memory[addr+1] = byte(val >> 8 & 0xff)
}

func (m *Machine) ReadInt8(addr int) int {
	return int(int8(m.memory[addr]))
}

func (m *Machine) ReadWord(addr int) int {
	if addr < 16 {
		return m.doSpecialReadWord(addr)
	}
	return int(m.memory[addr+1])<<8 + int(m.memory[addr])
}

// ReadString reads a null-terminated string from memory.
func (m *Machine) ReadString(addr uint16) string {
	buf := m.memory[addr:]
	i := 0
	for _, b := range buf {
		if b == 0 {
			break
		}
		i++
	}
	return string(buf[:i])
}

func (m *Machine) WriteWord(addr, value int) {
	if addr < 16 {
		m.doSpecialWriteWord(addr, value)
	} else {
		m.memory[addr] = byte(value & 0xff)
		m.memory[addr+1] = byte(value >> 8 & 0xff)
	}
}

func (m *Machine) doSpecialWriteWord(addr int, value int) {
	if addr == PCAddr {
		m.pc = uint16(value)
	} else if addr == SPAddr {
		m.sp = uint16(value)
	} else if addr == FPAddr {
		m.fp = uint16(value)
	} else if addr == IOReq {
		m.execIORequest(value)
	}
}

const (
	ErrNoErr uint16 = iota
	ErrInvalidDevice
	ErrInvalidCommand
	ErrIOError
)

func (m *Machine) execIORequest(addr int) {
	deviceId := int(m.readUint16(uint16(addr)))
	if deviceId >= len(m.devices) {
		m.writeUint16(IORes, ErrInvalidDevice)
		return
	}
	device := m.devices[deviceId]
	if device == nil {
		m.writeUint16(IORes, ErrInvalidDevice)
		return
	}
	err := device.Invoke(m, uint16(addr))
	m.writeUint16(IORes, err)
}

// Same as WriteWord but updates zero/minus flags
func (m *Machine) writeTarget(addr, value int) {
	if m.bytes {
		if addr == PCAddr || addr == SPAddr || addr == FPAddr {
			panic("attempt to write byte value to pc/sp/fp")
		}
		m.memory[addr] = byte(value)
		m.updateFlagsByte(value)
		//fmt.Printf("  0x%04x <- 0x%02x [z:%t, n:%t]\n", addr, value, m.zero, m.negative)
	} else {
		m.WriteWord(addr, value)
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

func (m *Machine) Run() {
	for {
		in := m.memory[m.pc]
		if in == 0 {
			return
		}
		n := 0      // Number of bytes for each operand
		bytes := 0  // Total count of operand bytes (to skip pc to next instruction)
		target := 0 // the address being updated, ie often the address of value1
		value1 := 0 // value of first operand, if any
		value2 := 0 // value of second operand, if any
		opCode, m1, m2 := DecodeOp(in)
		if m1 != Implied {
			target, value1, n = m.fetchOperand(m1, int(m.pc+1))
			bytes = n
		}
		if m2 != Implied {
			_, value2, n = m.fetchOperand(m2, int(m.pc+1)+bytes)
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
			m.writeTarget(int(m.sp), value1)
		case Pop:
			if m1 == ImmediateByte {
				m.sp += uint16(value1)
			} else {
				m.writeTarget(target, m.ReadWord(int(m.sp)))
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

func (m *Machine) pushWord(value int) {
	m.sp -= 2
	m.writeTarget(int(m.sp), value)
}

func (m *Machine) popWord() int {
	word := m.ReadWord(int(m.sp))
	m.sp += 2
	return word
}

func (m *Machine) pushUint16(addr uint16) {
	m.sp -= 2
	m.writeUint16(m.sp, addr)
}

func (m *Machine) popUint16() uint16 {
	val := m.readUint16(m.sp)
	m.sp += 2
	return val
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
		value = address - 1 + m.ReadInt8(pc)
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
		bytes := 0
		var args string
		if m1 != Implied && m2 != Implied {
			op1, n := m.formatOperand(m1, addr+1)
			bytes += n
			op2, n := m.formatOperand(m2, addr+1+n)
			bytes += n
			args = op1 + "," + op2
		} else if m1 != Implied && m2 == Implied {
			op1, n := m.formatOperand(m1, addr+1)
			bytes += n
			args = op1
		}
		fmt.Fprintf(w, "0x%04x  %02x ", addr, in)
		for j := 0; j < 4; j++ {
			if j < bytes {
				fmt.Fprintf(w, "%02x ", m.memory[addr+j+1])
			} else {
				fmt.Fprintf(w, "   ")
			}
		}
		fmt.Fprintf(w, "%s %s\n", op, args)
		addr = addr + 1 + bytes
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
	fmt.Printf("[status pc=%04x sp=%04x fp=%04x n=%d z=%d c=%d b=%d]\n", m.pc, m.sp, m.fp, boolInt(m.negative), boolInt(m.zero),
		boolInt(m.carry), boolInt(m.bytes))
	return int(m.pc)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m *Machine) formatOperand(mode AddressMode, pc int) (op string, bytes int) {
	switch mode {
	case Implied:
		// nothing to do
	case Immediate:
		op = fmt.Sprintf("#0x%04x", m.ReadWord(pc))
		bytes = 2
	case ImmediateByte:
		op = fmt.Sprintf("#0x%02x", m.ReadInt8(pc))
		bytes = 1
	case OffsetByte:
		value := m.ReadInt8(pc)
		op = fmt.Sprintf("0x%04x (%d)", (pc-1)+value, value)
		bytes = 1
	case Absolute:
		op = fmt.Sprintf("0x%04x", m.ReadWord(pc))
		bytes = 2
	case Indirect:
		op = fmt.Sprintf("*0x%04x", m.ReadWord(pc))
		bytes = 2
	case Relative:
		value := m.ReadInt8(pc)
		op = fmt.Sprintf("fp%+d", value)
		bytes = 1
	case RelativeIndirect:
		value := m.ReadInt8(pc)
		op = fmt.Sprintf("*fp%+d", value)
		bytes = 1
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	return
}

func (m *Machine) doSpecialReadWord(addr int) int {
	if addr == PCAddr {
		return int(m.pc)
	}
	if addr == SPAddr {
		return int(m.sp)
	}
	if addr == FPAddr {
		return int(m.fp)
	}
	if addr == RandAddr {
		return int(uint16(m.rgen.Intn(65536)))
	}
	return 0
}
