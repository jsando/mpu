package machine

import (
	"io"
	"os"
)

const (
	StdoutCommandNop uint16 = iota
	StdoutCommandWrite
)

type StdoutDevice struct {
}

func (d *StdoutDevice) Invoke(m *Machine, addr uint16) (err uint16) {
	cmd := m.readUint16(addr + 2)
	if cmd == StdoutCommandWrite {
		_, err := io.Copy(os.Stdout, &NullbyteReader{
			m:    m,
			addr: m.readUint16(addr + 4),
		})
		if err != nil {
			return ErrIOError
		}
		return ErrNoErr
	}
	return ErrInvalidCommand
}

type NullbyteReader struct {
	m    *Machine
	addr uint16
}

func (d *NullbyteReader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		b := d.m.memory[d.addr]
		if b == 0 {
			return n, io.EOF
		}
		//fmt.Printf("stdout: 0x%02x\n", b)
		p[n] = b
		n++
		d.addr++
		if d.addr == 0 {
			return n, io.EOF
		}
	}
	return
}
