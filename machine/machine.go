package machine

const (
	ProgramCounter = 0
	StackPointer   = 2
)

type Opcode byte

// Opcode constants, from 0-15 (0x00-0x0f).
const (
	OpAdd Opcode = iota
	OpSub
	OpMul
	OpDiv
	OpCmp
	OpAnd
	OpOr
	OpXor
	OpCpy
	OpPsh
	OpPop
	OpJmp
	OpJeq
	OpJne
	OpJge
	OpJlt
)

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
}

func (m *Machine) updateFlags(value int) {
	m.negative = value > 0x7fff
	m.zero = value == 0
}

func decodeOp(in byte) (opcode Opcode, mode OperandMode, size OperandSize) {
	opcode = Opcode(in & OpcodeMask)
	mode = OperandMode(in & OpModeMask)
	mode = OperandMode(in & OpModeMask)
	size = OperandSize(in & SizeMask)
	return
}

func EncodeOp(opcode Opcode, mode OperandMode, size OperandSize) byte {
	return byte(opcode) | byte(mode) | byte(size)
}

func (m *Machine) Run() {
	panic("todo")
}

func (m *Machine) Step() {
	pc := m.PC()
	in := m.memory[pc]
	opcode, mode, _ := decodeOp(in) // todo handle byte size
	opCount := 2
	if opcode > OpCpy {
		opCount = 1
	}
	n, target, value1, value2 := m.fetchOperands(mode, pc+1, opCount)
	m.WriteWord(ProgramCounter, pc+n+1) // todo panic on pc overflow

	switch opcode {
	case OpAdd:
		m.writeTarget(target, value1+value2)
	case OpSub:
		m.writeTarget(target, value1-value2)
	case OpMul:
		m.writeTarget(target, value1*value2)
	case OpDiv:
		m.writeTarget(target, value1/value2)
	case OpCmp:
		m.updateFlags(value1 - value2)
	case OpAnd:
		m.writeTarget(target, value1&value2)
	case OpOr:
		m.writeTarget(target, value1|value2)
	case OpXor:
		m.writeTarget(target, value1^value2)
	case OpCpy:
		m.writeTarget(target, value2)
	case OpPsh:
		m.writeTarget(m.ReadWord(StackPointer), value1)
		m.WriteWord(StackPointer, m.ReadWord(StackPointer)-2)
	case OpPop:
		m.WriteWord(StackPointer, m.ReadWord(StackPointer)+2)
		m.writeTarget(target, m.ReadWord(m.ReadWord(StackPointer)))
	case OpJmp:
		m.WriteWord(ProgramCounter, value1)
	case OpJeq:
		if m.zero {
			m.WriteWord(ProgramCounter, value1)
		}
	case OpJne:
		if !m.zero {
			m.WriteWord(ProgramCounter, value1)
		}
	case OpJge:
		if !m.negative {
			m.WriteWord(ProgramCounter, value1)
		}
	case OpJlt:
		if m.negative {
			m.WriteWord(ProgramCounter, value1)
		}
	}
}

func (m *Machine) fetchOperands(mode OperandMode, pc int, oCount int) (n int, target int, value1 int, value2 int) {
	switch mode {
	case ParamNone:
		panic("no such thing")
	case ParamImm:
		value1 = m.ReadWord(pc)
		n = 2
	case ParamAbsAbs:
		target = m.ReadWord(pc)
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(m.ReadWord(pc + 2))
			n = 4
		} else {
			n = 2
		}
	case ParamAbsImm:
		target = m.ReadWord(pc)
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(pc + 2)
			n = 4
		} else {
			n = 2
		}
	case ParamAbsInd:
		target = m.ReadWord(pc)
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(m.ReadWord(m.ReadWord(pc + 2)))
			n = 4
		} else {
			n = 2
		}
	case ParamIndAbs:
		target = m.ReadWord(m.ReadWord(pc))
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(m.ReadWord(pc + 2))
			n = 4
		} else {
			n = 2
		}
	case ParamIndImm:
		target = m.ReadWord(m.ReadWord(pc))
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(pc + 2)
			n = 4
		} else {
			n = 2
		}
	case ParamIndInd:
		target = m.ReadWord(m.ReadWord(pc))
		value1 = m.ReadWord(target)
		if oCount > 1 {
			value2 = m.ReadWord(m.ReadWord(m.ReadWord(pc + 2)))
			n = 4
		} else {
			n = 2
		}
	}
	return
}
