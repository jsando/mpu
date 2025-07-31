package machine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestReadByte(t *testing.T) {
	machine := NewMachine([]byte{20: 7 ^ 0xff + 1})
	val := machine.ReadInt8(20)
	if val != -7 {
		t.Errorf("expected: -7, got: %d", val)
	}
}

//func TestRandom(t *testing.T) {
//	// Yeah yeah I should be using a set seed value to make this deterministic but instead:
//	// Pull 10 successive values and fail if they are all the same.
//	machine := NewMachine([]byte{})
//	r := machine.memory.GetWord(RandAddr)
//	same := true
//	for i := 0; i < 10; i++ {
//		r2 := machine.memory.GetWord(RandAddr)
//		if r != r2 {
//			same = false
//			break
//		}
//	}
//	if same {
//		t.Errorf("expected different random values but got 10 samples of: %d", r)
//	}
//}

type testHandler struct {
	Id        uint16
	Status    uint16
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
		20: 0x99, 0x99, 0x00, 0x00, 0x11, 0x33, 0x22,
	})
	iod.execIORequest(20)
	result := machine.memory.GetWord(22)
	if result != 42 {
		t.Errorf("expected 42, got: %d", result)
	}
	if handler.ByteParam != 0x11 || handler.WordParam != 0x2233 {
		t.Errorf("bad unmarshall: %v", handler)
	}
}

func TestArithmetic(t *testing.T) {
	a := 1000
	tester := NewMachineTester()
	tester.emit2(Cpy, Absolute, a, Immediate, 2) // 2
	tester.emit2(Add, Absolute, a, Immediate, 2) // +2 = 4
	tester.emit2(Sub, Absolute, a, Immediate, 1) // -1 = 3
	tester.emit2(Mul, Absolute, a, Immediate, 5) // *5 = 15
	tester.emit2(Div, Absolute, a, Immediate, 3) // /3 = 5
	tester.emit2(Inc, Absolute, a, Implied, 0)   // +1 = 6
	tester.emit2(Inc, Absolute, a, Implied, 0)   // +1 = 7
	tester.emit2(Dec, Absolute, a, Implied, 0)   // -1 = 6
	tester.execute()
	tester.addressContains(t, uint16(a), 6)
}

func TestBitops(t *testing.T) {
	tester := NewMachineTester()
	tester.emit2(Or, Absolute, 20, Immediate, 0xc0c0)
	tester.emit2(And, Absolute, 20, Immediate, 0x4040)
	tester.emit2(Xor, Absolute, 20, Immediate, 0x0040)
	tester.execute()
	tester.addressContains(t, 20, 0x4000)
}

func TestCompare(t *testing.T) {
	tester := NewMachineTester()
	tester.emit2(Cpy, Absolute, 20, Immediate, 123)
	tester.emit2(Cmp, Absolute, 20, Immediate, 200)
	tester.execute()
	tester.addressContains(t, 20, 123)
	if !tester.machine.negative {
		t.Errorf("neg flag should be set")
	}
}

func TestStackops(t *testing.T) {
	tester := NewMachineTester()
	tester.emit1(Psh, Immediate, 123)
	tester.emit1(Psh, Immediate, 456)
	tester.emit1(Pop, Absolute, 1000)
	tester.emit1(Pop, ImmediateByte, 2)
	tester.execute()
	tester.addressContains(t, 1000, 456)
	if tester.machine.negative {
		t.Errorf("neg flag should not be set")
	}
	if tester.machine.zero {
		t.Error("zero flag should not be set")
	}
	sp := tester.machine.Sp
	if sp != 0xffff {
		t.Errorf("exptected sp 0xffff but got: 0x%0x", sp)
	}
}

func TestJump(t *testing.T) {
	tester := NewMachineTester()
	tester.emit1(Jmp, Immediate, 8)
	tester.emit2(Cpy, Absolute, 1000, Immediate, 123)
	tester.emit2(Cpy, Absolute, 1000, Immediate, 456)
	tester.execute()
	tester.addressContains(t, 1000, 456)
}

type MachineTester struct {
	code    []byte
	machine *Machine
}

func NewMachineTester() *MachineTester {
	return &MachineTester{}
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
	assert.Equal(t, value, word, "Memory address 0x%04x contains unexpected value", address)
}

func TestSeaInstruction(t *testing.T) {
	// Test that SEA sets the assertion flag
	code := []byte{
		byte(EncodeOp(Sea, Implied, Implied)), // SEA
		byte(EncodeOp(Hlt, Implied, Implied)), // HLT
	}

	machine := NewMachine(code)

	// Initially assertion flag should be false
	assert.False(t, machine.assertion, "Assertion flag should be false initially")

	// Execute SEA instruction
	machine.step = true
	machine.Run()

	// Assertion flag should now be true
	assert.True(t, machine.assertion, "Assertion flag should be true after SEA")
}

func TestCmpWithAssertion(t *testing.T) {
	// Test CMP behavior with assertion flag set
	// Set up memory locations for comparison
	code := []byte{
		// Store values at addresses 0x100 and 0x102
		byte(EncodeOp(Cpy, Absolute, Immediate)), // CPY 0x100, #5
		0x00, 0x01,                               // addr = 0x100
		0x05, 0x00, // value = 5
		byte(EncodeOp(Cpy, Absolute, Immediate)), // CPY 0x102, #5
		0x02, 0x01,                               // addr = 0x102
		0x05, 0x00, // value = 5

		// First comparison (should pass)
		byte(EncodeOp(Sea, Implied, Implied)),    // SEA
		byte(EncodeOp(Cmp, Absolute, Immediate)), // CMP 0x100, #5
		0x00, 0x01,                               // addr = 0x100
		0x05, 0x00, // value = 5

		// Second comparison (should fail)
		byte(EncodeOp(Sea, Implied, Implied)),    // SEA
		byte(EncodeOp(Cmp, Absolute, Immediate)), // CMP 0x100, #3
		0x00, 0x01,                               // addr = 0x100
		0x03, 0x00, // value = 3

		byte(EncodeOp(Hlt, Implied, Implied)), // HLT
	}

	machine := NewMachine(code)
	machine.EnableTestMode()

	// Run the program
	machine.Run()

	// First CMP should pass (5 == 5), second should fail (5 != 3)
	assert.Equal(t, 1, machine.AssertionFailures(), "Should have 1 assertion failure")

	// Assertion flag should be cleared after CMP
	assert.False(t, machine.assertion, "Assertion flag should be cleared after CMP")
}

func TestCmpWithoutAssertion(t *testing.T) {
	// Test that normal CMP behavior is unchanged without SEA
	code := []byte{
		// Store value at address 0x100
		byte(EncodeOp(Cpy, Absolute, Immediate)), // CPY 0x100, #5
		0x00, 0x01,                               // addr = 0x100
		0x05, 0x00, // value = 5

		byte(EncodeOp(Cmp, Absolute, Immediate)), // CMP 0x100, #3
		0x00, 0x01,                               // addr = 0x100
		0x03, 0x00, // value = 3
		byte(EncodeOp(Hlt, Implied, Implied)), // HLT
	}

	machine := NewMachine(code)
	machine.EnableTestMode()

	// Run the program
	machine.Run()

	// No assertion failures expected
	assert.Equal(t, 0, machine.AssertionFailures(), "Should have no assertion failures")

	// Check that flags are set correctly
	assert.False(t, machine.zero, "Zero flag should be false (5 != 3)")
	assert.False(t, machine.negative, "Negative flag should be false (5 - 3 = 2)")
}
