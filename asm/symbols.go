package asm

import "fmt"

type SymbolTable struct {
	symbols map[string]*Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{symbols: make(map[string]*Symbol)}
}

type Symbol struct {
	text    string
	file    string
	line    int
	value   int
	defined bool
	fp      bool
}

func (s *SymbolTable) GetSymbol(text string) *Symbol {
	return s.symbols[text]
}

func (s *SymbolTable) AddSymbol(file string, line int, text string) {
	symbol := &Symbol{
		text:  text,
		file:  file,
		line:  line,
		value: 0,
	}
	s.symbols[text] = symbol
	fmt.Printf("add %s\n", text)
}

func (s *SymbolTable) Define(text string, value int) {
	sym := s.GetSymbol(text)
	fmt.Printf("define %s=%d\n", text, value)
	if sym == nil {
		// todo some kinda error
	}
	sym.defined = true
	sym.value = value
}

func (s *SymbolTable) AddFpSymbol(file string, line int, text string, offset int) {
	symbol := &Symbol{
		text:    text,
		file:    file,
		line:    line,
		value:   offset,
		defined: true,
		fp:      true,
	}
	s.symbols[text] = symbol
	fmt.Printf("addfp %s=%d\n", text, offset)
}
