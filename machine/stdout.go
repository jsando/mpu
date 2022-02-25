package machine

import (
	"io"
	"os"
)

const (
	StdoutDeviceId     = 0x0100
	StdoutCommandWrite = 1
)

func RegisterStdIoHandlers(m *Machine) {
	m.RegisterIOHandler(StdoutDeviceId|StdoutCommandWrite, &StdoutWriteHandler{})
}

type StdoutWriteHandler struct {
	Id       uint16
	PZString uint16 // pointer to zero-terminated string
}

func (s *StdoutWriteHandler) Handle(m *Machine, addr uint16) (errCode uint16) {
	_, err := io.Copy(os.Stdout, &ZStringReader{
		m:    m,
		addr: s.PZString,
	})
	if err != nil {
		return ErrIOError
	}
	return ErrNoErr
}

type ZStringReader struct {
	m    *Machine
	addr uint16
}

func (d *ZStringReader) Read(p []byte) (n int, err error) {
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
