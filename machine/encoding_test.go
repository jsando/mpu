package machine

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddressMode_String(t *testing.T) {
	tests := []struct {
		mode AddressMode
		want string
	}{
		{mode: Implied, want: "implied"},
		{mode: Absolute, want: "absolute"},
		{mode: Immediate, want: "immediate"},
		{mode: Indirect, want: "indirect"},
		{mode: Relative, want: "relative"},
		{mode: RelativeIndirect, want: "relative-indirect"},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, test.mode.String())
	}
	assert.Panics(t, func() { AddressMode(255).String() }, "expected panic")
}

func TestOpCode_String(t *testing.T) {
	tests := []struct {
		op   OpCode
		want string
	}{
		{op: Hlt, want: "hlt"},
		{op: Add, want: "add"},
		{op: Sub, want: "sub"},
		{op: Mul, want: "mul"},
		{op: Div, want: "div"},
		{op: And, want: "and"},
		{op: Or, want: "or"},
		{op: Xor, want: "xor"},
		{op: Cpy, want: "cpy"},
		{op: Cmp, want: "cmp"},
		{op: Inc, want: "inc"},
		{op: Dec, want: "dec"},
		{op: Psh, want: "psh"},
		{op: Pop, want: "pop"},
		{op: Jsr, want: "jsr"},
		{op: Jmp, want: "jmp"},
		{op: Jeq, want: "jeq"},
		{op: Jne, want: "jne"},
		{op: Jge, want: "jge"},
		{op: Jlt, want: "jlt"},
		{op: Jcc, want: "jcc"},
		{op: Jcs, want: "jcs"},
		{op: Sav, want: "sav"},
		{op: Seb, want: "seb"},
		{op: Clb, want: "clb"},
		{op: Clc, want: "clc"},
		{op: Sec, want: "sec"},
		{op: Ret, want: "ret"},
		{op: Rst, want: "rst"},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, test.op.String())
	}
	assert.Panics(t, func() { OpCode(255).String() })
}

// Not trying to validate the entire matrix here, just verify a few
// are correct and that an undefined one returns a HLT.
func TestDecodeOp(t *testing.T) {
	tests := []struct {
		in     byte
		op     OpCode
		m1, m2 AddressMode
	}{
		{in: 0x00, op: Hlt},
		{in: 0x10, op: Add, m1: Absolute, m2: Absolute},
		{in: 0x1e, op: Xor, m1: Absolute, m2: Immediate},
		{in: 0xf0, op: Psh, m1: Immediate},
		{in: 0xcf, op: Cmp, m1: Relative, m2: Absolute},
		{in: 0x01, op: Hlt},
		{in: 0xff, op: Hlt},
	}
	for _, test := range tests {
		op, m1, m2 := DecodeOp(test.in)
		msg := fmt.Sprintf("for opcode %02x", test.in)
		assert.Equal(t, test.op, op, msg)
		assert.Equal(t, test.m1, m1, msg)
		assert.Equal(t, test.m2, m2, msg)
	}
}

func TestEncodeOp(t *testing.T) {
	tests := []struct {
		want   byte
		op     OpCode
		m1, m2 AddressMode
	}{
		{want: 0x00, op: Hlt},
		{want: 0x10, op: Add, m1: Absolute, m2: Absolute},
		{want: 0x1e, op: Xor, m1: Absolute, m2: Immediate},
		{want: 0xf0, op: Psh, m1: Immediate},
		{want: 0xcf, op: Cmp, m1: Relative, m2: Absolute},
	}
	for _, test := range tests {
		op := EncodeOp(test.op, test.m1, test.m2)
		assert.Equal(t, test.want, op)
	}
	assert.Panics(t, func() { EncodeOp(Clc, Absolute, Immediate) })
}
