package asm

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	str := `
pc:	    dw main		// entry point
sp:	    dw 0xffff	// stack pointer (grows down)

	    org 0x0100
main:
.ptr 	= 2
	psh #0x2000
.loop
    cpy *[fp+ptr], #0xd0d0
    add [fp+ptr], #2
    cmp [fp+ptr], #0x3000
    jlt loop
    db 0 // halt

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

func dumpTokens(in string) {
	lexer := NewLexer("test", strings.NewReader(in))
	parser := NewParser(lexer)
	parser.Parse()
	parser.messages.Print()
	f := parser.first
	for f != nil {
		fmt.Println(f)
		f = f.next
	}
	//for {
	//	tok := lexer.Next()
	//	if tok == TokEOF {
	//		break
	//	}
	//}
}
