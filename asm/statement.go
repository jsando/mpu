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

import (
	"fmt"

	"github.com/jsando/mpu/machine"
)

type (
	// Node defines the common fields that all statements share.
	Node struct {
		next          Statement
		file          string
		line          int
		newlineBefore bool
		blockComment  []string
		eolComment    string
		pcStart       int // assigned by the linker
		pcEnd         int
	}

	// Statement is an interface of Node just so all my types can implement the same interface?
	Statement interface {
		Next() Statement
		SetNext(next Statement)
		File() string
		SetFile(file string)
		Line() int
		SetLine(line int)
		NewlineBefore() bool
		SetNewlineBefore(newlineBefore bool)
		BlockComment() []string
		SetBlockComment(blockComment []string)
		EolComment() string
		SetEolComment(eolComment string)
		PcStart() int
		SetPcStart(pcStart int)
		PcEnd() int
		SetPcEnd(pcEnd int)
	}

	LabelStatement struct {
		Node
		name string
	}

	EquateStatement struct {
		Node
		name  string
		value Expr
	}

	DefineByteStatement struct {
		Node
		values []Expr
	}

	DefineWordStatement struct {
		Node
		values []Expr
	}

	DefineSpaceStatement struct {
		Node
		size Expr
	}

	OrgStatement struct {
		Node
		origin Expr
	}

	IncludeStatement struct {
		Node
		path string
	}

	InstructionStatement struct {
		Node
		operation TokenType // TokXXX constant for a directive or opcode
		operands  []*Operand
	}

	FunctionStatement struct {
		Node
		name     string
		fpArgs   []*FpParam
		fpLocals []*FpParam
	}

	TestStatement struct {
		Node
		name string
	}

	VarStatement struct {
		Node
		name string
		size int
	}
)

func (n *Node) String() string {
	return fmt.Sprintf("%s: %d pc (%d-%d)\n", n.file, n.line, n.pcStart, n.pcEnd)
}

func (n *Node) Next() Statement {
	return n.next
}

func (n *Node) SetNext(next Statement) {
	n.next = next
}

func (n *Node) File() string {
	return n.file
}

func (n *Node) SetFile(file string) {
	n.file = file
}

func (n *Node) Line() int {
	return n.line
}

func (n *Node) SetLine(line int) {
	n.line = line
}

func (n *Node) NewlineBefore() bool {
	return n.newlineBefore
}

func (n *Node) SetNewlineBefore(newlineBefore bool) {
	n.newlineBefore = newlineBefore
}

func (n *Node) BlockComment() []string {
	return n.blockComment
}

func (n *Node) SetBlockComment(blockComment []string) {
	n.blockComment = blockComment
}

func (n *Node) EolComment() string {
	return n.eolComment
}

func (n *Node) SetEolComment(eolComment string) {
	n.eolComment = eolComment
}

func (n *Node) PcStart() int {
	return n.pcStart
}

func (n *Node) SetPcStart(pcStart int) {
	n.pcStart = pcStart
}

func (n *Node) PcEnd() int {
	return n.pcEnd
}

func (n *Node) SetPcEnd(pcEnd int) {
	n.pcEnd = pcEnd
}

type FpParam struct {
	id     string
	size   int
	offset int // offset once assigned
}

type Operand struct {
	mode machine.AddressMode
	expr Expr
}

type Expr interface {
	computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool)
	hasFramePointerSymbols(symbols *SymbolTable) bool
}

type IntLiteral struct {
	value int
	text  string // original text from source, ie 0xff, 123, 0b_1111
}

func (e IntLiteral) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return false
}

func (e IntLiteral) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	return e.value, nil, true
}

type CharLiteral struct {
	value int
	text  string // original text from source, ie 'n'
}

func (c CharLiteral) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	return c.value, nil, true
}

func (c CharLiteral) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return false
}

type BytesLiteral struct {
	value []byte
	text  string
}

func (b BytesLiteral) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return false
}

func (b BytesLiteral) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	return 0, b.value, true
}

type ExprIdent struct {
	activeLabel string
	ident       string
}

func (e ExprIdent) hasFramePointerSymbols(symbols *SymbolTable) bool {
	sym := symbols.GetSymbol(e.activeLabel + "." + e.ident)
	if sym != nil && sym.defined {
		return sym.fp
	}
	sym = symbols.GetSymbol(e.ident)
	if sym != nil && sym.defined {
		return sym.fp
	}
	return false
}

func (e ExprIdent) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	sym := symbols.GetSymbol(e.activeLabel + "." + e.ident)
	if sym != nil && sym.defined {
		return sym.value, nil, true
	}
	sym = symbols.GetSymbol(e.ident)
	if sym != nil && sym.defined {
		return sym.value, nil, true
	}
	return 0, nil, false
}

// ExprUnary is a unary operation (+, -) on an expression.
type ExprUnary struct {
	op   TokenType
	expr Expr
}

func (e ExprUnary) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return e.expr.hasFramePointerSymbols(symbols)
}

func (e ExprUnary) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	// + is a noop, so only -
	ival, bval, resolved = e.expr.computeValue(symbols)
	if resolved && e.op == TokMinus {
		// todo: if bval is defined this doesn't make any sense
		ival = -ival
	}
	return
}

type ExprBinary struct {
	op    TokenType
	expr1 Expr
	expr2 Expr
}

func (e ExprBinary) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return e.expr1.hasFramePointerSymbols(symbols) || e.expr2.hasFramePointerSymbols(symbols)
}

func (e ExprBinary) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	i1, _, r1 := e.expr1.computeValue(symbols)
	i2, _, r2 := e.expr2.computeValue(symbols)
	resolved = r1 && r2
	if resolved {
		// todo: shouldn't have b1/b2 values, can't do addition on byte arrays
		switch e.op {
		case TokPlus:
			ival = i1 + i2
		case TokMinus:
			ival = i1 - i2
		case TokPipe:
			ival = i1 | i2
		case TokCaret:
			ival = i1 ^ i2
		case TokStar:
			ival = i1 * i2
		case TokSlash:
			ival = i1 / i2
		case TokPercent:
			ival = i1 % i2
		case TokLeftShift:
			ival = i1 << i2
		case TokRightShift:
			ival = i1 >> i2
		default:
			// todo: report error somehow but don't have reference to messages
			panic("help!")
		}
	}
	return
}

// Name returns the name of the test function.
func (t *TestStatement) Name() string {
	return t.name
}

// Name returns the name of the label.
func (l *LabelStatement) Name() string {
	return l.name
}
