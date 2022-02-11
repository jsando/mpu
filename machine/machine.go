package machine

import "fmt"

const (
	ProgramCounter = 0
	StackPointer   = 2
)

type Opcode byte

// Opcode constants, from 0-15 (0x00-0x0f).
const (
	Add Opcode = iota
	Sub
	Mul
	Div
	Cmp
	And
	Or
	Xor
	Cpy
	Psh
	Pop
	Jmp
	Jeq
	Jne
	Jge
	Jlt
)

var mnemonics = []string{
	"add", "sub", "mul", "div", "cmp",
	"and", "or", "xor", "cpy", "psh",
	"pop", "jmp", "jeq", "jne", "jge",
	"jlt",
}

func (o Opcode) String() string {
	return mnemonics[o]
}

const (
	OpcodeMask = 0b0000_1111
	OpModeMask = 0b0111_0000
	SizeMask   = 0b1000_0000
)

type OperandMode byte

// Addressing mode constants, for bits 4-6.
const (
	ParamNone OperandMode = iota << 4
	ParamImm
	ParamAbsAbs
	ParamAbsImm
	ParamAbsInd
	ParamIndAbs
	ParamIndImm
	ParamIndInd
)

type OperandSize byte

// Instruction size modifier, encoded in bit 7.
const (
	SizeByte OperandSize = 0b1000_0000
	SizeWord OperandSize = 0
)

type Machine struct {
	memory   [65536]byte
	negative bool
	zero     bool
}

func NewMachine() *Machine {
	return &Machine{}
}

func NewMachineFromSlice(buf []byte) *Machine {
	m := NewMachine()
	for i := 0; i < len(buf); i++ {
		m.memory[i] = buf[i]
	}
	return m
}

func (m *Machine) PC() int {
	return m.ReadWord(0)
}

func (m *Machine) ReadWord(addr int) int {
	return int(m.memory[addr+1])<<8 + int(m.memory[addr])
}

func (m *Machine) WriteWord(addr, value int) {
	lo := byte(value & 0xff)
	hi := byte(value >> 8 & 0xff)
	m.memory[addr] = lo
	m.memory[addr+1] = hi
}

// Same as WriteWord but updates zero/minus flags
func (m *Machine) writeTarget(addr, value int) {
	m.WriteWord(addr, value)
	m.updateFlags(value)
	//fmt.Printf("  0x%0x <- 0x%0x [z:%t, n:%t]\n", addr, value, m.zero, m.negative)
}

func (m *Machine) updateFlags(value int) {
	m.negative = value > 0x7fff || value < 0
	m.zero = value == 0
}

func decodeOp(in byte) (opcode Opcode, mode OperandMode, size OperandSize) {
	opcode = Opcode(in & OpcodeMask)
	mode = OperandMode(in & OpModeMask)
	size = OperandSize(in & SizeMask)
	return
}

func EncodeOp(opcode Opcode, mode OperandMode, size OperandSize) byte {
	return byte(opcode) | byte(mode) | byte(size)
}

func (m *Machine) Run() {
	for {
		pc := m.PC()
		in := m.memory[pc]
		if in == 0 {
			return
		}
		opcode, mode, _ := decodeOp(in) // todo handle byte size
		opCount := 2
		if opcode > Cpy {
			opCount = 1
		}
		n, target, value1, value2 := m.fetchOperands(mode, pc+1, opCount)

		// trace output
		//m.trace(pc, opcode, mode, target, value1, value2)

		m.WriteWord(ProgramCounter, pc+n+1) // todo panic on pc overflow

		switch opcode {
		case Add:
			m.writeTarget(target, value1+value2)
		case Sub:
			m.writeTarget(target, value1-value2)
		case Mul:
			m.writeTarget(target, value1*value2)
		case Div:
			m.writeTarget(target, value1/value2)
		case Cmp:
			m.updateFlags(value1 - value2)
		case And:
			m.writeTarget(target, value1&value2)
		case Or:
			m.writeTarget(target, value1|value2)
		case Xor:
			m.writeTarget(target, value1^value2)
		case Cpy:
			m.writeTarget(target, value2)
		case Psh:
			m.writeTarget(m.ReadWord(StackPointer), value1)
			m.WriteWord(StackPointer, m.ReadWord(StackPointer)-2)
		case Pop:
			m.WriteWord(StackPointer, m.ReadWord(StackPointer)+2)
			m.writeTarget(target, m.ReadWord(m.ReadWord(StackPointer)))
		case Jmp:
			m.WriteWord(ProgramCounter, value1)
		case Jeq:
			if m.zero {
				m.WriteWord(ProgramCounter, value1)
			}
		case Jne:
			if !m.zero {
				m.WriteWord(ProgramCounter, value1)
			}
		case Jge:
			if !m.negative {
				m.WriteWord(ProgramCounter, value1)
			}
		case Jlt:
			if m.negative {
				m.WriteWord(ProgramCounter, value1)
			}
		}
	}
}

func (m *Machine) fetchOperands(mode OperandMode, pc int, oCount int) (n int, target int, value1 int, value2 int) {
	mode1, mode2 := getMode(mode)
	n = 2
	target, value1 = m.fetchOperand(mode1, pc)
	if oCount > 1 {
		n = 4
		_, value2 = m.fetchOperand(mode2, pc+2)
	}
	return
}

func (m *Machine) fetchOperand(mode AddressMode, pc int) (address, value int) {
	switch mode {
	case Immediate:
		address = pc
		value = m.ReadWord(pc)
	case Absolute:
		address = m.ReadWord(pc)
		value = m.ReadWord(address)
	case Indirect:
		address = m.ReadWord(m.ReadWord(pc))
		value = m.ReadWord(address)
	default:
		panic(fmt.Sprintf("illegal address mode: %d", mode))
	}
	return
}

func (m *Machine) trace(pc int, opcode Opcode, mode OperandMode, target int, value1 int, value2 int) {
	var args string
	if opcode < Psh {
		// two-args operation
		args = fmt.Sprintf("param1: %s, param2: %s",
			formatArg(mode, 1, value1), formatArg(mode, 2, value2))
	} else {
		// single args opcode
		args = fmt.Sprintf("param1: %s", formatArg(mode, 1, value1))
	}
	fmt.Printf("pc: 0x%04x %s target: 0x%04x, %s\n", pc, opcode, target, args)
}

func (m *Machine) Snapshot() []byte {
	return m.memory[:]
}

func formatArg(mode OperandMode, argNum int, argValue int) string {
	m1, m2 := getMode(mode)
	t := m1
	if argNum == 2 {
		t = m2
	}
	var format string
	if t == Immediate {
		format = "#0x%04x"
	} else if t == Absolute {
		format = "0x%04x"
	} else if t == Indirect {
		format = "*0x%04x"
	}
	return fmt.Sprintf(format, argValue)
}

type AddressMode int

const (
	None AddressMode = iota
	Immediate
	Absolute
	Indirect
)

// getMode splits the instruction-encoded form into the separate modes for each arg.
func getMode(mode OperandMode) (op1, op2 AddressMode) {
	switch mode {
	case ParamNone:
		return None, None
	case ParamImm:
		return Immediate, None
	case ParamAbsAbs:
		return Absolute, Absolute
	case ParamAbsImm:
		return Absolute, Immediate
	case ParamAbsInd:
		return Absolute, Indirect
	case ParamIndAbs:
		return Indirect, Absolute
	case ParamIndImm:
		return Indirect, Immediate
	case ParamIndInd:
		return Indirect, Indirect
	}
	panic(fmt.Sprintf("invalid mode %d", mode))
}

func encodeOperandMode(m1, m2 AddressMode) OperandMode {
	if m1 == Immediate {
		return ParamImm
	}
	if m1 == Absolute {
		if m2 == Immediate {
			return ParamAbsImm
		}
		if m2 == Absolute {
			return ParamAbsAbs
		}
		if m2 == Indirect {
			return ParamAbsInd
		}
		if m2 == None {
			return ParamAbsAbs
		}
	}
	if m1 == Indirect {
		if m2 == Immediate {
			return ParamIndImm
		}
		if m2 == Absolute {
			return ParamIndAbs
		}
		if m2 == Indirect {
			return ParamIndInd
		}
		if m2 == None {
			return ParamIndAbs
		}
	}
	panic("invalid mode combination")
}
