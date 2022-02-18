package asm

import (
	"fmt"
	"github.com/jsando/mpu/machine"
	"strings"
)

// Statement represents a single directive or instruction.
// They form a linked list.
// The parser uses the lexer to create the list of statements,
// which are then fed to the linker to generate the machine code.
type Statement struct {
	next      *Statement
	file      string
	line      int
	labels    []string  //optional labels
	operation TokenType // TokXXX constant for a directive or opcode
	operands  []*Operand
	pcStart   int
	pcEnd     int
	fpArgs    []*FpParam
	fpLocals  []*FpParam
}

type FpParam struct {
	id     string
	size   int
	offset int // offset once assigned
}

func (f *Statement) AddFpArg(id string, size int) {
	f.fpArgs = append(f.fpArgs, &FpParam{
		id:   id,
		size: size,
	})
}

func (f *Statement) AddFpLocal(id string, size int) {
	f.fpLocals = append(f.fpLocals, &FpParam{
		id:   id,
		size: size,
	})
}

func (f *Statement) String() string {
	return fmt.Sprintf("%s: %s", strings.Join(f.labels, ","), f.operation)
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
}

func (e IntLiteral) hasFramePointerSymbols(symbols *SymbolTable) bool {
	return false
}

func (e IntLiteral) computeValue(symbols *SymbolTable) (ival int, bval []byte, resolved bool) {
	return e.value, nil, true
}

type BytesLiteral struct {
	value []byte
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
