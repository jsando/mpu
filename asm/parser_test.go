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
