package asm

import (
	"strings"
	"testing"
)

func TestParseLineNumbers(t *testing.T) {
	str := `pc: dw add_numbers
 a:   dw 0	
    dw 0

	org 0x10


    org 100

// add_numbers will add two numbers and return the result on the stack.
add_numbers(result word, a word, b word):
    var c word
 
	cpy result, a
    clc
    add result, b
	jeq loop
    ret
`
	parser := NewParserFromReader("test", strings.NewReader(str))
	parser.Parse()
	parser.messages.Print()
	if parser.messages.errors != 0 {
		t.Errorf("expected 0 errors, got %d", parser.messages.errors)
	}
	s := parser.Statements()
	lines := []int{
		1, 1, 2, 2, 3, 5, 8, 11, 12, 14, 15, 16, 17,
	}
	for _, line := range lines {
		if s.Line() != line {
			t.Errorf("wanted: %d, got: %d", line, s.Line())
		}
		s = s.Next()
	}
}

func TestErrorLineNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLine int
		wantErr  string
	}{
		{
			name: "invalid instruction on line 3",
			input: `org 0x100
clc
invalid_instruction
hlt`,
			wantLine: 3,
			wantErr:  "expected =, :, or ():",
		},
		{
			name: "invalid instruction after blank line",
			input: `org 0x100
clc

invalid_instruction
hlt`,
			wantLine: 4,
			wantErr:  "expected =, :, or ():",
		},
		{
			name: "invalid instruction on first line",
			input: `invalid_instruction
clc`,
			wantLine: 1,
			wantErr:  "expected =, :, or ():",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParserFromReader("test", strings.NewReader(tt.input))
			parser.Parse()

			if parser.messages.errors == 0 {
				t.Fatal("expected an error but got none")
			}

			// Check that the error is on the correct line
			foundCorrectLine := false
			for _, msg := range parser.messages.messages {
				if msg.messageType == MessageError && strings.Contains(msg.message, tt.wantErr) {
					// msg.line is 1-based line number
					if msg.line == tt.wantLine {
						foundCorrectLine = true
						break
					} else {
						t.Errorf("error reported on line %d, want line %d", msg.line, tt.wantLine)
					}
				}
			}

			if !foundCorrectLine {
				t.Errorf("did not find error on line %d", tt.wantLine)
			}
		})
	}
}

func TestParseTestFunction(t *testing.T) {
	// Test parsing a simple test function
	source := `
test MyTest():
    sea
    cmp a, #5
    ret
`
	parser := NewParserFromReader("test", strings.NewReader(source))
	parser.Parse()

	if parser.messages.errors != 0 {
		t.Fatalf("expected no errors, got %d", parser.messages.errors)
	}

	// Find the test statement
	var testStmt *TestStatement
	for s := parser.Statements(); s != nil; s = s.Next() {
		if ts, ok := s.(*TestStatement); ok {
			testStmt = ts
			break
		}
	}

	if testStmt == nil {
		t.Fatal("expected to find TestStatement")
	}

	if testStmt.name != "MyTest" {
		t.Errorf("expected test name 'MyTest', got '%s'", testStmt.name)
	}
}

func TestParseNormalFunction(t *testing.T) {
	// Ensure normal functions still work
	source := `
myFunc(a word, b word):
    add a, b
    ret
`
	parser := NewParserFromReader("test", strings.NewReader(source))
	parser.Parse()

	if parser.messages.errors != 0 {
		t.Fatalf("expected no errors, got %d", parser.messages.errors)
	}

	// Find the function statement
	var fnStmt *FunctionStatement
	for s := parser.Statements(); s != nil; s = s.Next() {
		if fs, ok := s.(*FunctionStatement); ok {
			fnStmt = fs
			break
		}
	}

	if fnStmt == nil {
		t.Fatal("expected to find FunctionStatement")
	}

	if fnStmt.name != "myFunc" {
		t.Errorf("expected function name 'myFunc', got '%s'", fnStmt.name)
	}
}

func TestParseTestAndNormalMixed(t *testing.T) {
	// Test parsing both test and normal functions in same file
	source := `
myFunc(a word):
    inc a
    ret

test TestMyFunc():
    cpy a, #5
    jsr myFunc
    sea
    cmp a, #6
    ret
`
	parser := NewParserFromReader("test", strings.NewReader(source))
	parser.Parse()

	if parser.messages.errors != 0 {
		t.Fatalf("expected no errors, got %d", parser.messages.errors)
	}

	// Count statements
	var fnCount, testCount int
	for s := parser.Statements(); s != nil; s = s.Next() {
		switch s.(type) {
		case *FunctionStatement:
			fnCount++
		case *TestStatement:
			testCount++
		}
	}

	if fnCount != 1 {
		t.Errorf("expected 1 function, got %d", fnCount)
	}

	if testCount != 1 {
		t.Errorf("expected 1 test, got %d", testCount)
	}
}
