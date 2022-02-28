package machine

import (
	"os"
)

const (
	StdoutDeviceId     = 0x0100
	StdoutCommandWrite = 1
)

type StdoutWriteHandler struct {
	Id       uint16
	PZString uint16 // pointer to zero-terminated string
}

func (s *StdoutWriteHandler) Handle(m Memory, addr uint16) (errCode uint16) {
	// This could use copy to avoid creating a string just to print it, but this
	// was simpler to code in the Memory interface for now.  For a toy 16 bit project
	// I doubt anything writing to stdout is going to be a bottleneck.
	str := m.ReadZString(s.PZString)
	_, err := os.Stdout.WriteString(str)
	if err != nil {
		return ErrIOError
	}
	return ErrNoErr
}
