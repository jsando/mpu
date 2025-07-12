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

package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jsando/mpu/asm"
	"github.com/jsando/mpu/machine"
	"github.com/stretchr/testify/assert"
)

func compileAndLoad(t *testing.T, source string) (*machine.Machine, *asm.SymbolTable, []asm.DebugInfo) {
	parser := asm.NewParserFromReader("test.s", strings.NewReader(source))
	parser.Parse()
	if parser.HasErrors() {
		parser.Messages().Print()
		t.Fatal("parse errors")
	}

	linker := asm.NewLinker(parser.Statements())
	linker.Link()
	if linker.HasErrors() {
		linker.Messages().Print()
		t.Fatal("link errors")
	}

	code := linker.Code()
	// Pad to 64KB
	if len(code) < 65536 {
		padded := make([]byte, 65536)
		copy(padded, code)
		code = padded
	}

	m := machine.NewMachine(code)
	return m, linker.Symbols(), linker.DebugInfo()
}

func TestExecutorPassingTest(t *testing.T) {
	source := `
		org 0x100
a:		dw 5

test TestPassing():
		sea
		cmp a, #5   // This should pass
		ret
`
	m, symbols, debugInfo := compileAndLoad(t, source)

	suite := &TestSuite{
		Tests: []TestInfo{{
			Name:     "TestPassing",
			Function: "TestPassing",
			File:     "test.s",
			Line:     5,
		}},
	}

	executor := NewTestExecutor(m, suite, symbols, debugInfo)
	err := executor.Run()
	assert.NoError(t, err)

	results := executor.Results()
	assert.Len(t, results, 1)
	assert.True(t, results[0].Passed)
	assert.Equal(t, "TestPassing", results[0].Name)
}

func TestExecutorFailingTest(t *testing.T) {
	source := `
		org 0x100
a:		dw 5

test TestFailing():
		sea
		cmp a, #10   // This should fail (5 != 10)
		ret
`
	m, symbols, debugInfo := compileAndLoad(t, source)

	suite := &TestSuite{
		Tests: []TestInfo{{
			Name:     "TestFailing",
			Function: "TestFailing",
			File:     "test.s",
			Line:     5,
		}},
	}

	executor := NewTestExecutor(m, suite, symbols, debugInfo)
	err := executor.Run()
	assert.NoError(t, err)

	results := executor.Results()
	assert.Len(t, results, 1)
	assert.False(t, results[0].Passed)
	assert.Contains(t, results[0].Message, "assertion(s) failed")
}

func TestExecutorMultipleTests(t *testing.T) {
	source := `
		org 0x100
a:		dw 5

test TestPass1():
		sea
		cmp a, #5
		ret

test TestPass2():
		sea
		cmp a, #5
		ret

test TestFail():
		sea
		cmp a, #10
		ret
`
	m, symbols, debugInfo := compileAndLoad(t, source)

	// Discover tests from the source
	parser := asm.NewParserFromReader("test.s", strings.NewReader(source))
	parser.Parse()
	suite, _ := DiscoverTests(parser.Statements())

	executor := NewTestExecutor(m, suite, symbols, debugInfo)
	err := executor.Run()
	assert.NoError(t, err)

	results := executor.Results()
	assert.Len(t, results, 3)

	passed, failed := executor.Summary()
	assert.Equal(t, 2, passed)
	assert.Equal(t, 1, failed)
}

func TestTerminalFormatter(t *testing.T) {
	results := []TestResult{
		{Name: "TestPass", Passed: true},
		{Name: "TestFail", Passed: false, Message: "assertion failed"},
	}

	// Test without color
	formatter := NewTerminalFormatter(true, false)
	var buf bytes.Buffer
	err := formatter.Format(results, &buf)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "✓ TestPass")
	assert.Contains(t, output, "✗ TestFail")
	assert.Contains(t, output, "assertion failed")
	assert.Contains(t, output, "1 passed")
	assert.Contains(t, output, "1 failed")
	assert.Contains(t, output, "2 total")
}
