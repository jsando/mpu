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
