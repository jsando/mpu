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
	"bytes"
	"math/rand"
)

// RNG exposes a random number generator as a Memory unit.
// Writes are ignored.  Both bytes and words can be read.
type RNG struct {
	gen *rand.Rand
}

func (r *RNG) BytesReaderAt(addr uint16) *bytes.Reader {
	panic("not supported")
}

func (r *RNG) ReadZString(addr uint16) string {
	panic("not supported")
}

func NewRNG(seed int64) *RNG {
	return &RNG{
		gen: rand.New(rand.NewSource(seed)),
	}
}

func (r *RNG) PutByte(addr uint16, b byte) {
	// nop
}

func (r *RNG) GetByte(addr uint16) byte {
	return byte(r.gen.Intn(256))
}

func (r *RNG) PutWord(addr uint16, w uint16) {
	// nop
}

func (r *RNG) GetWord(addr uint16) uint16 {
	return uint16(r.gen.Intn(65536))
}
