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
	"os"
)

const (
	StdoutDeviceId     = 0x0100
	StdoutCommandWrite = 1
)

type StdoutWriteHandler struct {
	Id       uint16 // 0x0101
	PZString uint16 // pointer to zero-terminated string
}

func (s *StdoutWriteHandler) Handle(m Memory, addr uint16) (errCode uint16) {
	// This could use copy to avoid creating a string just to print it, but this
	// was simpler to code want the Memory interface for now.  For a toy 16 bit project
	// I doubt anything writing to stdout is going to be a bottleneck.
	str := m.ReadZString(s.PZString)
	_, err := os.Stdout.WriteString(str)
	if err != nil {
		return ErrIOError
	}
	return ErrNoErr
}
