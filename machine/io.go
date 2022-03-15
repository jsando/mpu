package machine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// IOHandler represents a single command handler within a device.  Handlers must
// be registered via RegisterIOHandler.  When invoked, machine will use encoding/binary
// to unmarshall the data pointed to into the handler and then call its Handle() method.
type IOHandler interface {
	Handle(m Memory, addr uint16) (errCode uint16)
}

const (
	ErrNoErr uint16 = iota
	ErrInvalidHandler
	ErrIOError
)

// IODispatcher implements Memory, to map to an address, and provides another Memory object
// to view the status of the last IO request.
type IODispatcher struct {
	last           uint16
	lastRegister   *Register
	status         uint16
	statusRegister *Register
	memory         Memory            // IO gets passed a pointer into memory where the command is located
	ioHandlers     map[int]IOHandler // Registered io handlers
	traceIO        bool
}

func NewDefaultDispatcher() *IODispatcher {
	d := NewDispatcher()
	d.RegisterIOHandler(StdoutDeviceId|StdoutCommandWrite, &StdoutWriteHandler{})
	RegisterSDLHandlers(d)
	return d
}

func NewDispatcher() *IODispatcher {
	d := &IODispatcher{
		ioHandlers: make(map[int]IOHandler),
	}
	d.lastRegister = &Register{&d.last}
	d.statusRegister = &Register{&d.status}
	return d
}

func (d *IODispatcher) RegisterIOHandler(id int, h IOHandler) {
	d.ioHandlers[id] = h
}

func (d *IODispatcher) StatusRegister() Memory {
	return d.statusRegister
}

func (d *IODispatcher) execIORequest(addr uint16) {
	id := int(d.memory.GetWord(addr))
	handler := d.ioHandlers[id]
	if handler == nil {
		d.status = ErrInvalidHandler
		_, _ = fmt.Fprintf(os.Stderr, "io request to unknown handler (0x%04x)\n", id)
		return
	}
	err := binary.Read(d.memory.BytesReaderAt(addr), binary.LittleEndian, handler)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "io request decode error (handler=0x%04x, error=%s)\n", id, err.Error())
		d.status = ErrIOError
		return
	}
	errCode := handler.Handle(d.memory, addr)
	if d.traceIO || errCode != ErrNoErr {
		_, _ = fmt.Fprintf(os.Stderr, "io request (handler: 0x%04x, parameters: %v, status: %d)\n", id, handler, errCode)
	}
	d.status = errCode
}

func (d *IODispatcher) BytesReaderAt(addr uint16) *bytes.Reader {
	panic("can't get reader on iodispatcher")
}

func (d *IODispatcher) ReadZString(addr uint16) string {
	panic("not supported")
}

func (d *IODispatcher) PutByte(addr uint16, b byte) {
	panic("not supported")
}

func (d *IODispatcher) GetByte(addr uint16) byte {
	return d.lastRegister.GetByte(addr)
}

func (d *IODispatcher) PutWord(addr uint16, w uint16) {
	d.execIORequest(w)
}

func (d *IODispatcher) GetWord(addr uint16) uint16 {
	return d.lastRegister.GetWord(addr)
}

func LogIOError(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
