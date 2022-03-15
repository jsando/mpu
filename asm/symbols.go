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
	//fmt.Printf("add %s\n", text)
}

func (s *SymbolTable) Define(text string, value int) {
	sym := s.GetSymbol(text)
	//fmt.Printf("define %s=%d\n", text, value)
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
	//fmt.Printf("addfp %s=%d\n", text, offset)
}
