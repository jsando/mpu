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

func TestWriteUpdatesFlags(t *testing.T) {
	machine := NewMachine([]byte{})
	machine.writeTarget(16, 0)
	if machine.negative {
		t.Errorf("neg flag should be cleared after zero write")
	}
	if !machine.zero {
		t.Error("zero flag should be set")
	}
	machine.writeTarget(16, 40000)
	if !machine.negative {
		t.Errorf("neg flag should be set")
	}
	if machine.zero {
		t.Error("zero flag should be clear")
	}
	machine.writeTarget(16, -1)
	if !machine.negative {
		t.Errorf("neg flag should be set")
	}
	if machine.zero {
		t.Error("zero flag should be clear")
	}
	// PutWord should not modify any flags
	machine.writeTarget(16, 1)
	if machine.zero || machine.negative {
		t.Error("s/b zero and not neg")
	}
	machine.memory.PutWord(16, 0)
	if machine.zero {
		t.Error("zero flag should be clear")
	}
	machine.memory.PutWord(16, 65535) // make sure bit 16 doesn't trigger negative flag
	if machine.negative {
		t.Error("neg flag should be clear")
	}
}

func TestReadWritePC(t *testing.T) {
	machine := NewMachine([]byte{})
	machine.pc = 0x1234
	if machine.memory.GetWord(PCAddr) != 0x1234 {
		t.Errorf("expected 0x1234, got %0x", machine.memory.GetWord(PCAddr))
	}
	machine.memory.PutWord(PCAddr, 0x5555)
	if machine.pc != 0x5555 {
		t.Errorf("expected pc update, got: 0x%0x", machine.pc)
	}
}

func TestReadWriteSP(t *testing.T) {
	machine := NewMachine([]byte{})
	machine.sp = 0x3579
	if machine.memory.GetWord(SPAddr) != 0x3579 {
		t.Errorf("expected 0x3579, got %0x", machine.memory.GetWord(SPAddr))
	}
	machine.memory.PutWord(SPAddr, 0x5555)
	if machine.sp != 0x5555 {
		t.Errorf("expected sp update, got: 0x%0x", machine.sp)
	}
}

func TestReadWriteFP(t *testing.T) {
	machine := NewMachine([]byte{})
	machine.fp = 0x3456
	if machine.memory.GetWord(FPAddr) != 0x3456 {
		t.Errorf("expected 0x1234, got %0x", machine.memory.GetWord(FPAddr))
	}
	machine.memory.PutWord(FPAddr, 0x5555)
	if machine.fp != 0x5555 {
		t.Errorf("expected fp update, got: 0x%0x", machine.fp)
	}
}

func TestReadByte(t *testing.T) {
	machine := NewMachine([]byte{20: 7 ^ 0xff + 1})
	val := machine.ReadInt8(20)
	if val != -7 {
		t.Errorf("expected: -7, got: %d", val)
	}
}

func TestRandom(t *testing.T) {
	// Yeah yeah I should be using a set seed value to make this deterministic but instead:
	// Pull 10 successive values and fail if they are all the same.
	machine := NewMachine([]byte{})
	r := machine.memory.GetWord(RandAddr)
	same := true
	for i := 0; i < 10; i++ {
		r2 := machine.memory.GetWord(RandAddr)
		if r != r2 {
			same = false
			break
		}
	}
	if same {
		t.Errorf("expected different random values but got 10 samples of: %d", r)
	}
}

type testHandler struct {
	Id        uint16
	ByteParam uint8
	WordParam uint16
}

func (t *testHandler) Handle(m Memory, addr uint16) (err uint16) {
	return 42
}

func TestIO(t *testing.T) {
	handler := &testHandler{}
	iod := NewDispatcher()
	iod.RegisterIOHandler(0x9999, handler)
	machine := NewMachineWithDevices(iod, []byte{
		20: 0x99, 0x99, 0x11, 0x33, 0x22,
	})
	machine.memory.PutWord(IOReqAddr, 20)
	result := machine.memory.GetWord(IOStatAddr)
	if result != 42 {
		t.Errorf("expected 42, got: %d", result)
	}
	if handler.ByteParam != 0x11 || handler.WordParam != 0x2233 {
		t.Errorf("bad unmarshall: %v", handler)
	}
}

func TestArithmetic(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Cpy, Absolute, 20, Immediate, 2) // 2
	tester.emit2(Add, Absolute, 20, Immediate, 2) // +2 = 4
	tester.emit2(Sub, Absolute, 20, Immediate, 1) // -1 = 3
	tester.emit2(Mul, Absolute, 20, Immediate, 5) // *5 = 15
	tester.emit2(Div, Absolute, 20, Immediate, 3) // /3 = 5
	tester.emit2(Inc, Absolute, 20, Implied, 0)   // +1 = 6
	tester.emit2(Inc, Absolute, 20, Implied, 0)   // +1 = 7
	tester.emit2(Dec, Absolute, 20, Implied, 0)   // -1 = 6
	tester.execute()
	tester.addressContains(t, 20, 6)
}

func TestBitops(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Or, Absolute, 20, Immediate, 0xc0c0)
	tester.emit2(And, Absolute, 20, Immediate, 0x4040)
	tester.emit2(Xor, Absolute, 20, Immediate, 0x0040)
	tester.execute()
	tester.addressContains(t, 20, 0x4000)
}

func TestCompare(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit2(Cpy, Absolute, 20, Immediate, 123)
	tester.emit2(Cmp, Absolute, 20, Immediate, 200)
	tester.execute()
	tester.addressContains(t, 20, 123)
	if !tester.machine.negative {
		t.Errorf("neg flag should be set")
	}
}

func TestStackops(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit1(Psh, Immediate, 123)
	tester.emit1(Psh, Immediate, 456)
	tester.emit1(Pop, Absolute, 20)
	tester.emit1(Pop, ImmediateByte, 2)
	tester.execute()
	tester.addressContains(t, 20, 456)
	if tester.machine.negative {
		t.Errorf("neg flag should not be set")
	}
	if tester.machine.zero {
		t.Error("zero flag should not be set")
	}
	sp := tester.machine.memory.GetWord(SPAddr)
	if sp != 0x1000 {
		t.Errorf("exptected sp 0x1000 but got: 0x%0x", sp)
	}
}

func TestJump(t *testing.T) {
	tester := NewMachineTester(0x100, 0x1000)
	tester.emit1(Jmp, Immediate, 0x108)
	tester.emit2(Cpy, Absolute, 20, Immediate, 123)
	tester.emit2(Cpy, Absolute, 20, Immediate, 456)
	tester.execute()
	tester.addressContains(t, 20, 456)
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

func (c *MachineTester) addressContains(t *testing.T, address uint16, value int) {
	word := int(c.machine.memory.GetWord(address))
	if word != value {
		t.Errorf("expected: 0x%0x, got: 0x%0x", value, word)
	}
}
