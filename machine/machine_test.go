package machine

import (
	"testing"
)

func TestOpString(t *testing.T) {
	tests := []struct {
		op       OpCode
		mnemonic string
	}{
		{op: Add, mnemonic: "add"},
		{op: Sub, mnemonic: "sub"},
		{op: Mul, mnemonic: "mul"},
		{op: Div, mnemonic: "div"},
		{op: Cmp, mnemonic: "cmp"},
		{op: And, mnemonic: "and"},
		{op: Or, mnemonic: "or"},
		{op: Xor, mnemonic: "xor"},
		{op: Cpy, mnemonic: "cpy"},
		{op: Psh, mnemonic: "psh"},
		{op: Pop, mnemonic: "pop"},
		{op: Jmp, mnemonic: "jmp"},
		{op: Jeq, mnemonic: "jeq"},
		{op: Jne, mnemonic: "jne"},
		{op: Jge, mnemonic: "jge"},
		{op: Jlt, mnemonic: "jlt"},
	}
	for _, test := range tests {
		if test.op.String() != test.mnemonic {
			t.Errorf("expected %s, got %s", test.mnemonic, test.op.String())
		}
	}
}

func TestReadWrite(t *testing.T) {
	machine := NewMachine([]byte{})
	machine.WriteWord(10, 0x1122)
	if machine.memory[10] != 0x22 || machine.memory[11] != 0x11 {
		t.Errorf("bad write")
	}
	val := machine.ReadWord(10)
	if val != 0x1122 {
		t.Errorf("bad read")
	}
	machine.writeTarget(10, 0)
	if machine.memory[10] != 0 || machine.memory[11] != 0 {
		t.Errorf("bad rewrite")
	}
	if machine.negative {
		t.Errorf("neg flag should be cleared after zero write")
	}
	if !machine.zero {
		t.Error("zero flag should be set")
	}
	machine.writeTarget(10, 40000)
	if !machine.negative {
		t.Errorf("neg flag should be set")
	}
	if machine.zero {
		t.Error("zero flag should be clear")
	}
	machine.writeTarget(10, -1)
	if !machine.negative {
		t.Errorf("neg flag should be set")
	}
	if machine.zero {
		t.Error("zero flag should be clear")
	}
}

func TestReadByte(t *testing.T) {
	machine := NewMachine([]byte{7 ^ 0xff + 1})
	val := machine.ReadInt8(0)
	if val != -7 {
		t.Errorf("expected: -7, got: %d", val)
	}
}

//func TestFetchOperands(t *testing.T) {
//	tests := []struct {
//		buf     []byte
//		mode    AddressMode
//		opcount int
//		count   int
//		target  int
//		value1  int
//		value2  int
//	}{
//		{[]byte{1, 2}, ParamImm, 1, 2, 0, 0x0201, 0},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsAbs, 2, 4, 4, 1, 2},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsImm, 2, 4, 4, 1, 6},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsInd, 2, 4, 4, 1, 6},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndAbs, 2, 4, 1, 0x600, 2},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndImm, 2, 4, 1, 0x600, 6},
//		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndInd, 2, 4, 1, 0x600, 6},
//	}
//	for _, test := range tests {
//		m := NewMachineFromSlice(test.buf)
//		fmt.Printf("test: %v\n", test)
//		count, target, value1, value2 := m.fetchOperands(test.mode, 0, test.opcount)
//		if count != test.count || target != test.target || value1 != test.value1 || value2 != test.value2 {
//			t.Errorf("fail, got value1: 0x%0x, value2: 0x%0x, target: 0x%0x, ", value1, value2, target)
//		}
//	}
//}

func TestArithmetic(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Cpy, Absolute, 10, Immediate, 2) // 2
	tester.emit2(Add, Absolute, 10, Immediate, 2) // +2 = 4
	tester.emit2(Sub, Absolute, 10, Immediate, 1) // -1 = 3
	tester.emit2(Mul, Absolute, 10, Immediate, 5) // *5 = 15
	tester.emit2(Div, Absolute, 10, Immediate, 3) // /3 = 5
	tester.execute()
	tester.addressContains(t, 10, 5)
}

func TestBitops(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Or, Absolute, 10, Immediate, 0xc0c0)
	tester.emit2(And, Absolute, 10, Immediate, 0x4040)
	tester.emit2(Xor, Absolute, 10, Immediate, 0x0040)
	tester.execute()
	tester.addressContains(t, 10, 0x4000)
}

func TestCompare(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Cpy, Absolute, 10, Immediate, 123)
	tester.emit2(Cmp, Absolute, 10, Immediate, 200)
	tester.execute()
	tester.addressContains(t, 10, 123)
	if !tester.machine.negative {
		t.Errorf("neg flag should be set")
	}
}

func TestStackops(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit1(Psh, Immediate, 123)
	tester.emit1(Psh, Immediate, 456)
	tester.emit1(Pop, Absolute, 10)
	tester.execute()
	tester.addressContains(t, 10, 456)
	if tester.machine.negative {
		t.Errorf("neg flag should not be set")
	}
	if tester.machine.zero {
		t.Error("zero flag should not be set")
	}
	sp := tester.machine.ReadWord(SPAddr)
	if sp != 0x1000-2 {
		t.Errorf("exptected sp 0x1000-2 but got: 0x%0x", sp)
	}
}

func TestJump(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit1(Jmp, Immediate, 0x108)
	tester.emit2(Cpy, Absolute, 10, Immediate, 123)
	tester.emit2(Cpy, Absolute, 10, Immediate, 456)
	tester.execute()
	tester.addressContains(t, 10, 456)
}

type MachineTester struct {
	code    []byte
	machine *Machine
}

func NewMachineTester(org int, stack int) *MachineTester {
	tester := &MachineTester{}
	tester.writeWord(org)
	tester.writeWord(stack)
	for i := 4; i < org; i++ {
		tester.writeByte(0)
	}
	return tester
}

func (c *MachineTester) writeByte(b byte) {
	c.code = append(c.code, b)
}

func (c *MachineTester) writeWord(i int) {
	lo := byte(i & 0xff)
	hi := byte(i >> 8 & 0xff)
	c.writeByte(lo)
	c.writeByte(hi)
}

func (c *MachineTester) emit1(op OpCode, mode AddressMode, param int) {
	insn := EncodeOp(op, mode, Implied)
	c.writeByte(insn)
	c.writeWord(param)
}

func (c *MachineTester) emit2(op OpCode, mode AddressMode, address int, mode2 AddressMode, param2 int) {
	insn := EncodeOp(op, mode, mode2)
	c.writeByte(insn)
	c.writeWord(address)
	c.writeWord(param2)
}

func (c *MachineTester) execute() {
	code := make([]byte, len(c.code)+1)
	copy(code, c.code)
	c.machine = NewMachine(code)
	c.machine.Run()
}

func (c *MachineTester) addressContains(t *testing.T, address, value int) {
	word := c.machine.ReadWord(address)
	if word != value {
		t.Errorf("expected: 0x%0x, got: 0x%0x", value, word)
	}
}
