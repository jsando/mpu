package machine

import "fmt"

type AddressMode byte

const (
	Implied AddressMode = iota
	Absolute
	Immediate
	ImmediateByte
	OffsetByte
	Indirect
	Relative
	RelativeIndirect
)

func (m AddressMode) String() string {
	switch m {
	case Implied:
		return "implied"
	case Absolute:
		return "absolute"
	case Immediate:
		return "immediate"
	case Indirect:
		return "indirect"
	case Relative:
		return "relative"
	case RelativeIndirect:
		return "relative-indirect"
	}
	panic("invalid mode")
}

type OpCode byte

const (
	Hlt OpCode = iota
	Add
	Sub
	Mul
	Div
	And
	Or
	Xor
	Cpy
	Cmp
	Inc
	Dec
	Psh
	Pop
	Jsr
	Jmp
	Jeq
	Jne
	Jge
	Jlt
	Jcc
	Jcs
	Seb
	Clb
	Clc
	Sec
	Ret
	Sav
	Rst
)

var mnemonics = []string{
	"hlt", "add", "sub", "mul", "div",
	"and", "or", "xor", "cpy", "cmp",
	"inc", "dec", "psh", "pop", "jmp",
	"jeq", "jne", "jge", "jlt", "jcc",
	"jcs", "jsr", "seb", "clb", "clc",
	"sec", "ret", "sav", "rst",
}

func (o OpCode) String() string {
	return mnemonics[o]
}

func DecodeOp(in byte) (OpCode, AddressMode, AddressMode) {
	encoding := opTable[in]
	return encoding.op, encoding.m1, encoding.m2
}

func EncodeOp(op OpCode, m1, m2 AddressMode) byte {
	for i := 0; i < len(opTable); i++ {
		enc := opTable[i]
		if enc.op == op && enc.m1 == m1 && enc.m2 == m2 {
			return byte(i)
		}
	}
	panic(fmt.Sprintf("invalid encoding: %s (%s, %s)", op, m1, m2))
}

// Encoding defines the opcode and two address modes for each instruction.
// TODO It might be easier to both define and lookup at runtime by using bitfields
// in an int, rather than fields in a struct.  If each operand mode is a flag
// then the combos can be too ... AbsAbs.
// Add | AbsAbs, Sub | AbsAbs, Mul | AbsAbs, ...
// may be worth a microbenchmark to see the runtime performance variations.
type Encoding struct {
	op OpCode
	m1 AddressMode
	m2 AddressMode
}

/*
	Instruction Encoding Table

The main 9 instructions (add, sub, mul, div, and, or, xor, cpy, cmp) support 20 modes:
	Abs,Abs     Ind,Abs     Rel,Abs     RelInd,Abs
	Abs,Imm     Ind,Imm     Rel,Imm		RelInd,Imm
	Abs,Ind     Ind,Ind     Rel,Ind		RelInd,Ind
	Abs,Rel     Ind,Rel     Rel,Rel		RelInd,Rel
	Abs,RelInd  Ind,RelInd  Rel,RelInd  RelInd,RelInd

The 8 jump instructions only support immediate mode.

There are 5 instructions with implied mode.

Push supports all 5 modes.  Pop with immediate value pops and discards that number of bytes.

*/
var opTable = []Encoding{
	0x00: {op: Hlt, m1: Implied, m2: Implied},

	0x10: {op: Add, m1: Absolute, m2: Absolute},
	{op: Sub, m1: Absolute, m2: Absolute},
	{op: Mul, m1: Absolute, m2: Absolute},
	{op: Div, m1: Absolute, m2: Absolute},
	{op: And, m1: Absolute, m2: Absolute},
	{op: Or, m1: Absolute, m2: Absolute},
	{op: Xor, m1: Absolute, m2: Absolute},
	{op: Cpy, m1: Absolute, m2: Absolute},

	{op: Add, m1: Absolute, m2: Immediate},
	{op: Sub, m1: Absolute, m2: Immediate},
	{op: Mul, m1: Absolute, m2: Immediate},
	{op: Div, m1: Absolute, m2: Immediate},
	{op: And, m1: Absolute, m2: Immediate},
	{op: Or, m1: Absolute, m2: Immediate},
	{op: Xor, m1: Absolute, m2: Immediate},
	{op: Cpy, m1: Absolute, m2: Immediate},

	0x20: {op: Add, m1: Absolute, m2: Indirect},
	{op: Sub, m1: Absolute, m2: Indirect},
	{op: Mul, m1: Absolute, m2: Indirect},
	{op: Div, m1: Absolute, m2: Indirect},
	{op: And, m1: Absolute, m2: Indirect},
	{op: Or, m1: Absolute, m2: Indirect},
	{op: Xor, m1: Absolute, m2: Indirect},
	{op: Cpy, m1: Absolute, m2: Indirect},

	{op: Add, m1: Absolute, m2: Relative},
	{op: Sub, m1: Absolute, m2: Relative},
	{op: Mul, m1: Absolute, m2: Relative},
	{op: Div, m1: Absolute, m2: Relative},
	{op: And, m1: Absolute, m2: Relative},
	{op: Or, m1: Absolute, m2: Relative},
	{op: Xor, m1: Absolute, m2: Relative},
	{op: Cpy, m1: Absolute, m2: Relative},

	0x30: {op: Add, m1: Absolute, m2: RelativeIndirect},
	{op: Sub, m1: Absolute, m2: RelativeIndirect},
	{op: Mul, m1: Absolute, m2: RelativeIndirect},
	{op: Div, m1: Absolute, m2: RelativeIndirect},
	{op: And, m1: Absolute, m2: RelativeIndirect},
	{op: Or, m1: Absolute, m2: RelativeIndirect},
	{op: Xor, m1: Absolute, m2: RelativeIndirect},
	{op: Cpy, m1: Absolute, m2: RelativeIndirect},

	{op: Add, m1: Indirect, m2: Absolute},
	{op: Sub, m1: Indirect, m2: Absolute},
	{op: Mul, m1: Indirect, m2: Absolute},
	{op: Div, m1: Indirect, m2: Absolute},
	{op: And, m1: Indirect, m2: Absolute},
	{op: Or, m1: Indirect, m2: Absolute},
	{op: Xor, m1: Indirect, m2: Absolute},
	{op: Cpy, m1: Indirect, m2: Absolute},

	0x40: {op: Add, m1: Indirect, m2: Immediate},
	{op: Sub, m1: Indirect, m2: Immediate},
	{op: Mul, m1: Indirect, m2: Immediate},
	{op: Div, m1: Indirect, m2: Immediate},
	{op: And, m1: Indirect, m2: Immediate},
	{op: Or, m1: Indirect, m2: Immediate},
	{op: Xor, m1: Indirect, m2: Immediate},
	{op: Cpy, m1: Indirect, m2: Immediate},

	{op: Add, m1: Indirect, m2: Relative},
	{op: Sub, m1: Indirect, m2: Relative},
	{op: Mul, m1: Indirect, m2: Relative},
	{op: Div, m1: Indirect, m2: Relative},
	{op: And, m1: Indirect, m2: Relative},
	{op: Or, m1: Indirect, m2: Relative},
	{op: Xor, m1: Indirect, m2: Relative},
	{op: Cpy, m1: Indirect, m2: Relative},

	0x50: {op: Add, m1: Indirect, m2: RelativeIndirect},
	{op: Sub, m1: Indirect, m2: RelativeIndirect},
	{op: Mul, m1: Indirect, m2: RelativeIndirect},
	{op: Div, m1: Indirect, m2: RelativeIndirect},
	{op: And, m1: Indirect, m2: RelativeIndirect},
	{op: Or, m1: Indirect, m2: RelativeIndirect},
	{op: Xor, m1: Indirect, m2: RelativeIndirect},
	{op: Cpy, m1: Indirect, m2: RelativeIndirect},

	{op: Add, m1: Relative, m2: Absolute},
	{op: Sub, m1: Relative, m2: Absolute},
	{op: Mul, m1: Relative, m2: Absolute},
	{op: Div, m1: Relative, m2: Absolute},
	{op: And, m1: Relative, m2: Absolute},
	{op: Or, m1: Relative, m2: Absolute},
	{op: Xor, m1: Relative, m2: Absolute},
	{op: Cpy, m1: Relative, m2: Absolute},

	0x60: {op: Add, m1: Relative, m2: Immediate},
	{op: Sub, m1: Relative, m2: Immediate},
	{op: Mul, m1: Relative, m2: Immediate},
	{op: Div, m1: Relative, m2: Immediate},
	{op: And, m1: Relative, m2: Immediate},
	{op: Or, m1: Relative, m2: Immediate},
	{op: Xor, m1: Relative, m2: Immediate},
	{op: Cpy, m1: Relative, m2: Immediate},

	{op: Add, m1: Relative, m2: Indirect},
	{op: Sub, m1: Relative, m2: Indirect},
	{op: Mul, m1: Relative, m2: Indirect},
	{op: Div, m1: Relative, m2: Indirect},
	{op: And, m1: Relative, m2: Indirect},
	{op: Or, m1: Relative, m2: Indirect},
	{op: Xor, m1: Relative, m2: Indirect},
	{op: Cpy, m1: Relative, m2: Indirect},

	0x70: {op: Add, m1: Relative, m2: Relative},
	{op: Sub, m1: Relative, m2: Relative},
	{op: Mul, m1: Relative, m2: Relative},
	{op: Div, m1: Relative, m2: Relative},
	{op: And, m1: Relative, m2: Relative},
	{op: Or, m1: Relative, m2: Relative},
	{op: Xor, m1: Relative, m2: Relative},
	{op: Cpy, m1: Relative, m2: Relative},

	{op: Add, m1: Relative, m2: RelativeIndirect},
	{op: Sub, m1: Relative, m2: RelativeIndirect},
	{op: Mul, m1: Relative, m2: RelativeIndirect},
	{op: Div, m1: Relative, m2: RelativeIndirect},
	{op: And, m1: Relative, m2: RelativeIndirect},
	{op: Or, m1: Relative, m2: RelativeIndirect},
	{op: Xor, m1: Relative, m2: RelativeIndirect},
	{op: Cpy, m1: Relative, m2: RelativeIndirect},

	0x80: {op: Add, m1: RelativeIndirect, m2: Absolute},
	{op: Sub, m1: RelativeIndirect, m2: Absolute},
	{op: Mul, m1: RelativeIndirect, m2: Absolute},
	{op: Div, m1: RelativeIndirect, m2: Absolute},
	{op: And, m1: RelativeIndirect, m2: Absolute},
	{op: Or, m1: RelativeIndirect, m2: Absolute},
	{op: Xor, m1: RelativeIndirect, m2: Absolute},
	{op: Cpy, m1: RelativeIndirect, m2: Absolute},

	{op: Add, m1: RelativeIndirect, m2: Immediate},
	{op: Sub, m1: RelativeIndirect, m2: Immediate},
	{op: Mul, m1: RelativeIndirect, m2: Immediate},
	{op: Div, m1: RelativeIndirect, m2: Immediate},
	{op: And, m1: RelativeIndirect, m2: Immediate},
	{op: Or, m1: RelativeIndirect, m2: Immediate},
	{op: Xor, m1: RelativeIndirect, m2: Immediate},
	{op: Cpy, m1: RelativeIndirect, m2: Immediate},

	0x90: {op: Add, m1: RelativeIndirect, m2: Indirect},
	{op: Sub, m1: RelativeIndirect, m2: Indirect},
	{op: Mul, m1: RelativeIndirect, m2: Indirect},
	{op: Div, m1: RelativeIndirect, m2: Indirect},
	{op: And, m1: RelativeIndirect, m2: Indirect},
	{op: Or, m1: RelativeIndirect, m2: Indirect},
	{op: Xor, m1: RelativeIndirect, m2: Indirect},
	{op: Cpy, m1: RelativeIndirect, m2: Indirect},

	{op: Add, m1: RelativeIndirect, m2: Relative},
	{op: Sub, m1: RelativeIndirect, m2: Relative},
	{op: Mul, m1: RelativeIndirect, m2: Relative},
	{op: Div, m1: RelativeIndirect, m2: Relative},
	{op: And, m1: RelativeIndirect, m2: Relative},
	{op: Or, m1: RelativeIndirect, m2: Relative},
	{op: Xor, m1: RelativeIndirect, m2: Relative},
	{op: Cpy, m1: RelativeIndirect, m2: Relative},

	0xA0: {op: Add, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Sub, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Mul, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Div, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: And, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Or, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Xor, m1: RelativeIndirect, m2: RelativeIndirect},
	{op: Cpy, m1: RelativeIndirect, m2: RelativeIndirect},

	0xB0: {op: Psh, m1: Absolute},
	{op: Pop, m1: Absolute},
	{op: Inc, m1: Absolute},
	{op: Dec, m1: Absolute},

	{op: Sec},
	{op: Clc},
	{op: Seb},
	{op: Clb},
	{op: Ret},
	{op: Rst},
	{op: Sav, m1: ImmediateByte},

	0xC0: {op: Psh, m1: Indirect},
	{op: Pop, m1: Indirect},
	{op: Inc, m1: Indirect},
	{op: Dec, m1: Indirect},

	{op: Cmp, m1: Absolute, m2: Immediate},
	{op: Cmp, m1: Absolute, m2: Absolute},
	{op: Cmp, m1: Absolute, m2: Indirect},
	{op: Cmp, m1: Absolute, m2: Relative},
	{op: Cmp, m1: Absolute, m2: RelativeIndirect},
	{op: Cmp, m1: Indirect, m2: Immediate},
	{op: Cmp, m1: Indirect, m2: Absolute},
	{op: Cmp, m1: Indirect, m2: Indirect},
	{op: Cmp, m1: Indirect, m2: Relative},
	{op: Cmp, m1: Indirect, m2: RelativeIndirect},
	{op: Cmp, m1: Relative, m2: Immediate},
	{op: Cmp, m1: Relative, m2: Absolute},

	0xD0: {op: Psh, m1: Relative},
	{op: Pop, m1: Relative},
	{op: Inc, m1: Relative},
	{op: Dec, m1: Relative},

	{op: Cmp, m1: Relative, m2: Indirect},
	{op: Cmp, m1: Relative, m2: Relative},
	{op: Cmp, m1: Relative, m2: RelativeIndirect},
	{op: Cmp, m1: RelativeIndirect, m2: Immediate},
	{op: Cmp, m1: RelativeIndirect, m2: Absolute},
	{op: Cmp, m1: RelativeIndirect, m2: Indirect},
	{op: Cmp, m1: RelativeIndirect, m2: Relative},
	{op: Cmp, m1: RelativeIndirect, m2: RelativeIndirect},

	0xE0: {op: Psh, m1: RelativeIndirect},
	{op: Pop, m1: RelativeIndirect},
	{op: Inc, m1: RelativeIndirect},
	{op: Dec, m1: RelativeIndirect},
	{op: Jmp, m1: Immediate},
	{op: Jeq, m1: OffsetByte},
	{op: Jne, m1: OffsetByte},
	{op: Jge, m1: OffsetByte},
	{op: Jlt, m1: OffsetByte},
	{op: Jcc, m1: OffsetByte},
	{op: Jcs, m1: OffsetByte},
	{op: Jsr, m1: Immediate},

	0xF0: {op: Psh, m1: Immediate},
	{op: Pop, m1: ImmediateByte},
}
