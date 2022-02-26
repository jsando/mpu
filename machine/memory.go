package machine

import "bytes"

// Memory defines the basic abstraction of reading/writing words and bytes to
// random access memory OR to memory-mapped registers.
type Memory interface {
	PutByte(addr uint16, b byte)
	GetByte(addr uint16) byte
	PutWord(addr uint16, w uint16)
	GetWord(addr uint16) uint16
	ReadZString(addr uint16) string
	BytesReaderAt(addr uint16) *bytes.Reader
}

// Register defines uint16 which is mapped to memory and therefore supports
// the put/get byte/word primitives.
type Register struct {
	value *uint16
}

func (r Register) BytesReaderAt(addr uint16) *bytes.Reader {
	panic("can't get reader on register")
}

func (r Register) ReadZString(addr uint16) string {
	panic("can't read string from register")
}

func (r *Register) PutByte(addr uint16, b byte) {
	// If address is odd, that's the high byte
	if addr&0x01 != 0 {
		*r.value = (*r.value & 0xff) | uint16(b)<<8
	} else {
		*r.value = (*r.value & 0xff00) | uint16(b)
	}
}

func (r *Register) GetByte(addr uint16) byte {
	// If address is odd, that's the high byte
	if addr&0x01 != 0 {
		return byte(*r.value >> 8)
	}
	return byte(*r.value & 0xff)
}

func (r *Register) PutWord(addr uint16, w uint16) {
	*r.value = w
}

func (r *Register) GetWord(addr uint16) uint16 {
	return *r.value
}

func (r *Register) Get() uint16 {
	return *r.value
}

func (r *Register) Set(w uint16) {
	*r.value = w
}

// ByteSliceMemory contains a set of (continuous) memory-mapped regions
// and a raw byte slice.  The memory mapped area must be contiguous, and
// start at offset zero, and must consist of all word-sized memory.
type ByteSliceMemory struct {
	mapped      []Memory // Optional memory-mapped, uh, memory ... indexed by address (incl. addr + 1)
	mappedCount uint16   // len(mapped), to avoid computing it at GHz/sec
	raw         []byte   // Raw underlying bytes
}

func (m *ByteSliceMemory) BytesReaderAt(addr uint16) *bytes.Reader {
	return bytes.NewReader(m.raw[addr:])
}

func (m *ByteSliceMemory) ReadZString(addr uint16) string {
	buf := m.raw[addr:]
	i := 0
	for _, b := range buf {
		if b == 0 {
			break
		}
		i++
	}
	return string(buf[:i])
}

func NewByteSliceMemory(registers []Memory, raw []byte) *ByteSliceMemory {
	b := &ByteSliceMemory{
		raw: make([]byte, 65536),
	}
	copy(b.raw, raw)
	for i := 0; i < len(registers); i++ {
		b.mapped = append(b.mapped, registers[i])
		b.mapped = append(b.mapped, registers[i])
	}
	b.mappedCount = uint16(len(b.mapped))
	return b
}

func (m *ByteSliceMemory) PutByte(addr uint16, b byte) {
	if addr < m.mappedCount {
		m.mapped[addr].PutByte(addr, b)
	} else {
		m.raw[addr] = b
	}
}

func (m *ByteSliceMemory) GetByte(addr uint16) byte {
	if addr < m.mappedCount {
		return m.mapped[addr].GetByte(addr)
	}
	return m.raw[addr]
}

func (m *ByteSliceMemory) PutWord(addr uint16, w uint16) {
	if addr < m.mappedCount {
		m.mapped[addr].PutWord(addr, w)
	} else {
		m.raw[addr] = byte(w & 0xff)
		m.raw[addr+1] = byte(w >> 8 & 0xff)
	}
}

func (m *ByteSliceMemory) GetWord(addr uint16) uint16 {
	if addr < m.mappedCount {
		return m.mapped[addr].GetWord(addr)
	}
	return uint16(m.raw[addr+1])<<8 + uint16(m.raw[addr])
}

func readOrDefault(image []byte, addr int, i uint16) uint16 {
	if len(image) > addr+1 {
		return uint16(image[addr]) | uint16(image[addr+1])<<8
	}
	return i
}
