// Copyright 2022 Jason Sando <jason.sando.lv@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package machine

import (
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
	memory     Memory            // IO gets passed a pointer into memory where the command is located
	ioHandlers map[int]IOHandler // Registered io handlers
	traceIO    bool
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
	return d
}

func (d *IODispatcher) RegisterIOHandler(id int, h IOHandler) {
	d.ioHandlers[id] = h
}

func (d *IODispatcher) execIORequest(addr uint16) {
	id := int(d.memory.GetWord(addr))
	handler := d.ioHandlers[id]
	if handler == nil {
		d.memory.PutWord(addr+2, ErrInvalidHandler)
		_, _ = fmt.Fprintf(os.Stderr, "io request to unknown handler (0x%04x)\n", id)
		return
	}
	err := binary.Read(d.memory.BytesReaderAt(addr), binary.LittleEndian, handler)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "io request decode error (handler=0x%04x, error=%s)\n", id, err.Error())
		d.memory.PutWord(addr+2, ErrIOError)
		return
	}
	errCode := handler.Handle(d.memory, addr)
	if d.traceIO || errCode != ErrNoErr {
		_, _ = fmt.Fprintf(os.Stderr, "io request (handler: 0x%04x, parameters: %v, status: %d)\n", id, handler, errCode)
	}
	d.memory.PutWord(addr+2, errCode)
}

func LogIOError(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
