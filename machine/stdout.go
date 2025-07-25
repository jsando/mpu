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
	"math/rand"
	"os"
)

const (
	StdioDeviceId = 0x0100
	StdoutWrite   = 1
	GetRandom     = 2
)

type StdoutWriteHandler struct {
	IOHeader
	PZString uint16 // pointer to zero-terminated string
}

func (s *StdoutWriteHandler) Handle(m Memory, addr uint16) (errCode uint16) {
	str := m.ReadZString(s.PZString)
	_, err := os.Stdout.WriteString(str)
	if err != nil {
		return ErrIOError
	}
	return ErrNoErr
}

type GetRandomHandler struct {
	IOHeader
	Random uint16
}

func (h *GetRandomHandler) Handle(m Memory, addr uint16) (errCode uint16) {
	m.PutWord(addr+4, uint16(rand.Intn(65536)))
	return ErrNoErr
}
