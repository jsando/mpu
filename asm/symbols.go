package asm

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
}

func (s *SymbolTable) Define(text string, value int) {
	sym := s.GetSymbol(text)
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
}
