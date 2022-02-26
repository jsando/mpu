package machine

import "testing"

func TestBytesAndWords(t *testing.T) {
	mem := NewByteSliceMemory([]Memory{}, []byte{})
	addr := uint16(16)
	mem.PutWord(addr, 0x1122)
	if mem.GetByte(addr) != 0x22 || mem.GetByte(addr+1) != 0x11 {
		t.Errorf("bad write")
	}
	mem.PutByte(addr, 0x69)
	val := mem.GetWord(addr)
	if val != 0x1169 {
		t.Errorf("bad read")
	}
}

func TestRegister(t *testing.T) {
	var val uint16
	reg := &Register{&val}
	reg.PutWord(0, 0x1234)
	if val != 0x1234 {
		t.Errorf("bad write")
	}
	reg.PutByte(0, 0x44)
	if val != 0x1244 {
		t.Errorf("bad write")
	}
	reg.PutByte(1, 0x44)
	if val != 0x4444 {
		t.Errorf("bad write")
	}
	if reg.GetWord(0) != 0x4444 {
		t.Errorf("bad read")
	}
	val = 0x1234
	if reg.GetByte(0) != 0x34 || reg.GetByte(1) != 0x12 {
		t.Errorf("bad read")
	}
	reg.Set(0x4321)
	if reg.Get() != 0x4321 {
		t.Errorf("bad read")
	}
}

func TestReadString(t *testing.T) {
	mem := NewByteSliceMemory([]Memory{}, []byte{
		20: 'h', 'e', 'l', 'l', 'o', 0,
	})
	s := mem.ReadZString(20)
	if s != "hello" {
		t.Errorf("expected 'hello', got '%s'", s)
	}
}
