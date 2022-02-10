package machine

import "testing"

//func TestNewMachine(t *testing.T) {
//	machine := NewMachine()
//	machine.WriteWord(InitVector, 0x1000)
//	machine.Reset()
//	if machine.PC() != 0x1000 {
//		t.Errorf("expected pc of 0x1000, got %d", machine.PC())
//	}
//}
//
func TestDecode(t *testing.T) {
	tests := []struct {
		in     byte
		opcode Opcode
		mode   OperandMode
		size   OperandSize
	}{
		{0x30, OpAdd, ParamAbsImm, SizeWord},
		{0xa1, OpSub, ParamAbsAbs, SizeByte},
		{0x78, OpCpy, ParamIndInd, SizeWord},
	}
	for _, test := range tests {
		opcode, mode, size := decodeOp(test.in)
		if opcode != test.opcode || mode != test.mode || size != test.size {
			t.Errorf("bad decode for %0x, got op=%d, mode=%d, size:%v", test.in, opcode, mode, size)
		}
	}
}

func TestFetchOperands(t *testing.T) {
	tests := []struct {
		buf    []byte
		mode   OperandMode
		count  int
		target int
		value1 int
		value2 int
	}{
		{[]byte{1, 2}, ParamImm, 2, 0, 0x0201, 0},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsAbs, 4, 4, 1, 2},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsImm, 4, 4, 1, 6},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamAbsInd, 4, 4, 1, 6},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndAbs, 4, 1, 0x600, 2},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndImm, 4, 1, 0x600, 6},
		{[]byte{4, 0, 6, 0, 1, 0, 2, 0}, ParamIndInd, 4, 1, 0x600, 6},
	}
	for _, test := range tests {
		m := NewMachineFromSlice(test.buf)
		count, target, value1, value2 := m.fetchOperands(test.mode, 0, 2)
		if count != test.count || target != test.target || value1 != test.value1 || value2 != test.value2 {
			t.Errorf("fail, got value1: 0x%0x, value2: 0x%0x, target: 0x%0x, ", value1, value2, target)
		}
	}
}
