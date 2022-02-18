package asm

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseFunctions(t *testing.T) {
	str := `
    dw add_numbers
    dw 0
    dw 0

    org 100
add_numbers(result word, a word, b word):
    .c local word

    cpy result, a
    clc
    add result, b
    ret
`
	lexer := NewLexer("test", strings.NewReader(str))
	parser := NewParser(lexer)
	parser.Parse()
	parser.messages.Print()
	if parser.messages.errors != 0 {
		t.Errorf("expected 0 errors, got %d", parser.messages.errors)
	}
	linker := NewLinker(parser.Fragments())
	linker.Link()
	linker.messages.Print()
	fmt.Printf("Code size: %d\n", linker.pc)
	fmt.Println(hex.Dump(linker.code[0:linker.pc]))
	fmt.Println()

	WriteListing(strings.NewReader(str), os.Stdout, linker)
	//machine := machine.NewMachineFromSlice(linker.Code())
	//machine.Run()
}
