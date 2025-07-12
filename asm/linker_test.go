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

package asm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugInfoCollection(t *testing.T) {
	source := `
		org 0x100
a:		dw 0
start:	clc
		add a, #5
		sea
		cmp a, #5
		hlt
`
	parser := NewParserFromReader("test.s", strings.NewReader(source))
	parser.Parse()
	assert.False(t, parser.HasErrors())

	linker := NewLinker(parser.Statements())
	linker.Link()
	if linker.HasErrors() {
		linker.messages.Print()
	}
	assert.False(t, linker.HasErrors())

	debugInfo := linker.DebugInfo()
	assert.NotEmpty(t, debugInfo)

	// Should have debug info for each instruction
	// clc, add, sea, cmp, hlt = 5 instructions
	assert.GreaterOrEqual(t, len(debugInfo), 5)

	// First instruction (clc) should be after the dw (at 0x102)
	assert.Equal(t, uint16(0x102), debugInfo[0].PC)
	assert.Equal(t, "test.s", debugInfo[0].File)
}

func TestDebugInfoAccuracy(t *testing.T) {
	source := `
		org 0x100
a:		dw 0
		clc         // Line 3
		sea         // Line 4
		cmp a, #10  // Line 5
`
	parser := NewParserFromReader("test.s", strings.NewReader(source))
	parser.Parse()
	assert.False(t, parser.HasErrors())

	linker := NewLinker(parser.Statements())
	linker.Link()
	if linker.HasErrors() {
		linker.messages.Print()
	}
	assert.False(t, linker.HasErrors())

	debugInfo := linker.DebugInfo()

	// Find debug info for each instruction
	var clcInfo, seaInfo, cmpInfo *DebugInfo
	for i := range debugInfo {
		switch debugInfo[i].Line {
		case 4:
			clcInfo = &debugInfo[i]
		case 5:
			seaInfo = &debugInfo[i]
		case 6:
			cmpInfo = &debugInfo[i]
		}
	}

	assert.NotNil(t, clcInfo, "Should have debug info for CLC")
	assert.NotNil(t, seaInfo, "Should have debug info for SEA")
	assert.NotNil(t, cmpInfo, "Should have debug info for CMP")

	// Verify PC values are sequential (each instruction takes at least 1 byte)
	assert.Less(t, clcInfo.PC, seaInfo.PC)
	assert.Less(t, seaInfo.PC, cmpInfo.PC)
}

func TestDebugInfoMultipleFiles(t *testing.T) {
	// This test would require import support, which is more complex
	// For now, just verify that file names are tracked correctly
	source := `
		org 0x100
a:		dw 0
b:		dw 0
mytest:	sea
		cmp a, b
		ret
`
	parser := NewParserFromReader("mytest.s", strings.NewReader(source))
	parser.Parse()
	if parser.HasErrors() {
		parser.Messages().Print()
	}
	assert.False(t, parser.HasErrors())

	linker := NewLinker(parser.Statements())
	linker.Link()
	if linker.HasErrors() {
		linker.messages.Print()
	}
	assert.False(t, linker.HasErrors())

	debugInfo := linker.DebugInfo()
	for _, info := range debugInfo {
		assert.Equal(t, "mytest.s", info.File)
	}
}
