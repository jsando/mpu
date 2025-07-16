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
	"strings"
	"testing"

	"github.com/jsando/mpu/asm"
	"github.com/stretchr/testify/assert"
)

func parseSource(t *testing.T, source string) asm.Statement {
	parser := asm.NewParserFromReader("test.s", strings.NewReader(source))
	parser.Parse()
	if parser.HasErrors() {
		parser.Messages().Print()
		t.Fatal("parse errors")
	}
	return parser.Statements()
}

func TestDiscoverSingleTest(t *testing.T) {
	source := `
test TestAdd():
    cpy a, #5
    cpy b, #3
    add a, b
    sea
    cmp a, #8
    ret
`
	statements := parseSource(t, source)
	suite, err := DiscoverTests(statements)

	assert.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Len(t, suite.Tests, 1)

	test := suite.Tests[0]
	assert.Equal(t, "TestAdd", test.Name)
	assert.Equal(t, "TestAdd", test.Function)
	assert.Equal(t, "test.s", test.File)
	assert.Equal(t, 2, test.Line)
}

func TestDiscoverMultipleTests(t *testing.T) {
	source := `
test TestOne():
    ret

test TestTwo():
    ret
    
test TestThree():
    ret
`
	statements := parseSource(t, source)
	suite, err := DiscoverTests(statements)

	assert.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Len(t, suite.Tests, 3)

	assert.Equal(t, "TestOne", suite.Tests[0].Name)
	assert.Equal(t, "TestTwo", suite.Tests[1].Name)
	assert.Equal(t, "TestThree", suite.Tests[2].Name)
}

func TestDiscoverNoTests(t *testing.T) {
	source := `
myFunc(a word):
    inc a
    ret
    
main:
    jsr myFunc
    hlt
`
	statements := parseSource(t, source)
	suite, err := DiscoverTests(statements)

	assert.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Empty(t, suite.Tests)
}

func TestDiscoverMixedFunctions(t *testing.T) {
	source := `
helper(x word):
    inc x
    ret

test TestHelper():
    cpy a, #5
    jsr helper
    sea
    cmp a, #6
    ret
    
normalFunc():
    ret
    
test TestAnother():
    ret
`
	statements := parseSource(t, source)
	suite, err := DiscoverTests(statements)

	assert.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Len(t, suite.Tests, 2)

	assert.Equal(t, "TestHelper", suite.Tests[0].Name)
	assert.Equal(t, "TestAnother", suite.Tests[1].Name)
}

func TestDiscoverTestSetupTeardown(t *testing.T) {
	source := `
test_setup:
    // Initialize test environment
    ret
    
test TestSomething():
    ret
    
test_teardown:
    // Clean up
    ret
`
	statements := parseSource(t, source)
	suite, err := DiscoverTests(statements)

	assert.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Len(t, suite.Tests, 1)
	assert.Equal(t, "test_setup", suite.SetupFn)
	assert.Equal(t, "test_teardown", suite.TeardownFn)
}
