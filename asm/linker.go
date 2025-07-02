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
	"strings"

	"github.com/jsando/mpu/machine"
)

// DebugInfo maps PC addresses to source locations
type DebugInfo struct {
	PC     uint16
	File   string
	Line   int
	Column int
}

type Linker struct {
	statements Statement
	symbols    *SymbolTable
	messages   *Messages
	function   *FunctionStatement // Pointer to statement if in function with automatic fp/sav/rst handling
	pc         int
	code       []byte
	patches    []patch
	debugInfo  []DebugInfo
}

// patch is a expression with a forward reference to be resoled on pass 2
type patch struct {
	statement  Statement
	expr       Expr
	offsetByte bool
	pc         int
	size       int // 1 = byte, 2 = word
}

func NewLinker(stmt Statement) *Linker {
	return &Linker{
		statements: stmt,
		symbols:    NewSymbolTable(),
		messages:   &Messages{},
		code:       make([]byte, 65536),
	}
}

// Link uses two passes to try to resolve all references and generate code into l.code.
func (l *Linker) Link() {
	for stmt := l.statements; stmt != nil; stmt = stmt.Next() {
		stmt.SetPcStart(l.pc)
		switch t := stmt.(type) {
		case *EquateStatement:
			l.doEquate(t)
		case *DefineByteStatement:
			l.doDefineByte(t)
		case *DefineWordStatement:
			l.doDefineWord(t)
		case *DefineSpaceStatement:
			l.doDefineSpace(t)
		case *ImportStatement:
			// nothing to do
		case *OrgStatement:
			l.doOrg(t)
		case *LabelStatement:
			l.defineLabel(t, t.name)
		case *FunctionStatement:
			l.doFunction(t)
		case *TestStatement:
			l.doTest(t)
		case *VarStatement:
			// nothing to do, they are appended to the FunctionStatement
		case *InstructionStatement:
			l.overrideFramePointerSymbols(t)
			switch t.operation {
			case TokSec, TokClc, TokSeb, TokClb, TokRet, TokRst, TokHlt, TokSea:
				l.doEmit0Operand(t)
			case TokAdd, TokSub, TokMul, TokDiv, TokCmp, TokAnd, TokOr, TokXor, TokCpy:
				l.doEmit2Operand(t)
			case TokJmp, TokJsr:
				l.doEmitAbsJump(t)
			case TokJeq, TokJne, TokJge, TokJlt, TokJcc, TokJcs:
				l.doEmitRelJump(t)
			case TokInc, TokDec, TokPsh, TokPop, TokSav:
				l.doEmit1Operand(t)
			default:
				panic("illegal opcode")
			}
		default:
			panic("unknown statement type")
		}
		stmt.SetPcEnd(l.pc)
	}
	for _, patch := range l.patches {
		ival, _, res := patch.expr.computeValue(l.symbols)
		if res {
			if patch.size == 1 {
				if patch.offsetByte {
					ival = ival - patch.pc + 1
				}
				l.writeByteAt(ival, patch.pc)
			} else if patch.size == 2 {
				l.writeWordAt(ival, patch.pc)
			} else {
				panic("invalid patch size")
			}
		} else {
			l.errorf(patch.statement, "expression still unresolved after second pass")
		}
	}
	//fmt.Printf("ast dump:\n")
	//for stmt := l.statements; stmt != nil; stmt = stmt.Next() {
	//	fmt.Printf("%v\n", stmt)
	//}
}

func (l *Linker) addPatch(stmt Statement, expr Expr, size int, offsetByte bool) {
	p := patch{
		statement:  stmt,
		expr:       expr,
		pc:         l.pc,
		size:       size,
		offsetByte: offsetByte,
	}
	l.patches = append(l.patches, p)
	for i := 0; i < size; i++ {
		l.writeByte(0)
	}
}

func (l *Linker) defineLabel(s Statement, name string) {
	global := false
	//for _, s := range s.labels {
	if l.isErrorDefined(s, name) {
		return
	}
	l.symbols.AddSymbol(s.File(), s.Line(), name)
	l.symbols.Define(name, l.pc)
	if !strings.ContainsAny(name, ".") {
		global = true
	}
	//}
	if global {
		l.function = nil
	}
}

func (l *Linker) errorf(stmt Statement, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.messages.Error(stmt.File(), stmt.Line(), 0, msg)
}

func (l *Linker) writeByte(val int) {
	l.code[l.pc] = byte(val)
	l.pc++
}

func (l *Linker) writeBytes(val []byte) {
	n := copy(l.code[l.pc:], val)
	l.pc += n
	if n != len(val) {
		panic("need to validate and report this as error in caller")
	}
}

func (l *Linker) writeWord(val int) {
	lo := byte(val & 0xff)
	hi := byte(val >> 8)
	l.code[l.pc] = lo
	l.code[l.pc+1] = hi
	l.pc += 2
}

func (l *Linker) writeByteAt(val, pc int) {
	l.code[pc] = byte(val)
}

func (l *Linker) writeWordAt(val int, pc int) {
	lo := byte(val & 0xff)
	hi := byte(val >> 8)
	l.code[pc] = lo
	l.code[pc+1] = hi
}

// recordDebugInfo records the current PC and source location
func (l *Linker) recordDebugInfo(stmt Statement) {
	l.debugInfo = append(l.debugInfo, DebugInfo{
		PC:     uint16(l.pc),
		File:   stmt.File(),
		Line:   stmt.Line(),
		Column: 0, // Column tracking would require lexer changes
	})
}

func (l *Linker) doEquate(equate *EquateStatement) {
	if l.isErrorDefined(equate, equate.name) {
		return
	}
	l.symbols.AddSymbol(equate.file, equate.line, equate.name)
	ival, _, res := equate.value.computeValue(l.symbols)
	if !res {
		l.errorf(equate, "equate must have constant value")
	}
	l.symbols.Define(equate.name, ival)
}

func (l *Linker) doOrg(org *OrgStatement) {
	// org bumps to PC to a new location.  Error to go backwards.
	ival, _, res := org.origin.computeValue(l.symbols)
	if !res {
		l.errorf(org, "org must have constant value")
	} else {
		if l.pc > ival {
			l.errorf(org, "pc already at %d, can't set back to %d", l.pc, ival)
		} else {
			l.pc = ival
		}
	}
}

func (l *Linker) doDefineWord(dw *DefineWordStatement) {
	for _, expr := range dw.values {
		ival, bval, res := expr.computeValue(l.symbols)
		if res {
			if bval != nil {
				// todo: does it even make sense to write a string in lo/hi byte order???? I say no.
				l.errorf(dw, "can't use string with dw")
			} else {
				l.writeWord(ival)
			}
		} else {
			l.addPatch(dw, expr, 2, false)
		}
	}
}

func (l *Linker) doDefineByte(db *DefineByteStatement) {
	for _, expr := range db.values {
		ival, bval, res := expr.computeValue(l.symbols)
		if res {
			if bval != nil {
				l.writeBytes(bval)
			} else {
				l.writeByte(ival)
			}
		} else {
			l.addPatch(db, expr, 1, false)
		}
	}
}

func (l *Linker) doDefineSpace(ds *DefineSpaceStatement) {
	ival, bval, res := ds.size.computeValue(l.symbols)
	if !res {
		l.errorf(ds, "illegal forward reference in ds")
		return
	}
	if bval != nil {
		l.errorf(ds, "ds requires int param")
	}
	for i := 0; i < ival; i++ {
		l.writeByte(0)
	}
}

/*
 *			1st param   fp+8
 *			2nd param	fp+6
 *			3rd param   fp+4
 *			return addr	fp+2
 *   fp --> saved fp
 *	 		local-1		fp-2
 *	 sp --> local-2		fp-4
 */
func (l *Linker) doFunction(fn *FunctionStatement) {
	// define any labels, esp the global label for this function.
	// it will emit a SAV at this address to make space for locals.
	l.defineLabel(fn, fn.name)
	l.function = fn

	// compute the fp offset for all params and locals and
	// add them to the symbol table
	offset := 0
	localSize := 0
	for _, local := range fn.fpLocals {
		localSize += local.size
		offset -= local.size
		local.offset = offset
		if offset < -128 {
			l.errorf(fn, "fp local offset out of range (-1,-128): %d", offset)
		}
		l.symbols.AddFpSymbol(fn.file, fn.line, local.id, local.offset)
	}
	offset = 4
	for i := len(fn.fpArgs) - 1; i >= 0; i-- {
		arg := fn.fpArgs[i]
		arg.offset = offset
		if offset > 127 {
			l.errorf(fn, "fp arg offset out of range (0,127): %d", offset)
		}
		l.symbols.AddFpSymbol(fn.file, fn.line, arg.id, arg.offset)
		offset += arg.size
	}
	opCode := machine.EncodeOp(machine.Sav, machine.ImmediateByte, machine.Implied)
	l.writeByte(int(opCode))
	l.writeByte(localSize)
}

func (l *Linker) doTest(test *TestStatement) {
	// Test functions are just like regular functions but without parameters
	// Define the test function label
	l.defineLabel(test, test.name)
	l.function = nil  // Tests don't have automatic fp handling like functions
}

func (l *Linker) doEmit2Operand(stmt *InstructionStatement) {
	if len(stmt.operands) != 2 {
		l.errorf(stmt, "expected 2 operands")
		return
	}
	l.recordDebugInfo(stmt)
	op1 := stmt.operands[0]
	op2 := stmt.operands[1]
	op := tokToOp(stmt.operation)
	opCode := machine.EncodeOp(op, op1.mode, op2.mode)
	l.writeByte(int(opCode))
	l.resolveWordOperand(stmt, op1)
	l.resolveWordOperand(stmt, op2)
}

func (l *Linker) doEmit1Operand(ins *InstructionStatement) {
	if len(ins.operands) != 1 {
		l.errorf(ins, "expected 1 operand")
		return
	}
	l.recordDebugInfo(ins)
	op1 := ins.operands[0]
	op := tokToOp(ins.operation)
	if l.function != nil && op == machine.Sav {
		l.errorf(ins, "within functions, asm generates automatic SAV")
		return
	}
	if op == machine.Pop && op1.mode == machine.Immediate {
		op1.mode = machine.ImmediateByte
	}
	opCode := machine.EncodeOp(op, op1.mode, machine.Implied)
	l.writeByte(int(opCode))
	l.resolveWordOperand(ins, op1)
}

func (l *Linker) doEmitAbsJump(ins *InstructionStatement) {
	if len(ins.operands) != 1 {
		l.errorf(ins, "expected 1 operand")
		return
	}
	// override so we don't need # on all jumps
	ins.operands[0].mode = machine.Immediate
	l.doEmit1Operand(ins)
}

func (l *Linker) doEmitRelJump(stmt *InstructionStatement) {
	if len(stmt.operands) != 1 {
		l.errorf(stmt, "expected 1 operand")
		return
	}
	stmt.operands[0].mode = machine.OffsetByte
	l.doEmit1Operand(stmt)
}

func (l *Linker) doEmit0Operand(stmt *InstructionStatement) {
	if len(stmt.operands) != 0 {
		l.errorf(stmt, "expected 0 operands, got %d", len(stmt.operands))
		return
	}
	l.recordDebugInfo(stmt)
	op := tokToOp(stmt.operation)
	if l.function != nil && op == machine.Ret {
		op = machine.Rst
	}
	opCode := machine.EncodeOp(op, machine.Implied, machine.Implied)
	l.writeByte(int(opCode))
}

func (l *Linker) resolveWordOperand(ins *InstructionStatement, op *Operand) {
	// technically I want ImmediateByte to be an unsigned int but ... ok for now to make them all int8
	byteOp := false
	switch op.mode {
	case machine.ImmediateByte, machine.OffsetByte, machine.Relative, machine.RelativeIndirect:
		byteOp = true
	}
	ival, bval, res := op.expr.computeValue(l.symbols)
	if res {
		if bval != nil {
			l.errorf(ins, "expected int value, not []byte")
			return
		}
		if byteOp {
			if op.mode == machine.OffsetByte {
				ival = ival - l.pc + 1

				if ival > 127 || ival < -128 {
					l.errorf(ins, "relative offset out of range: %d", ival)
				}
			}
			l.writeByte(ival)
		} else {
			l.writeWord(ival)
		}
	} else {
		if byteOp {
			l.addPatch(ins, op.expr, 1, op.mode == machine.OffsetByte)
		} else {
			l.addPatch(ins, op.expr, 2, false)
		}
	}
}

func (l *Linker) BytesFor(stmt Statement) []byte {
	return l.code[stmt.PcStart():stmt.PcEnd()]
}

func (l *Linker) PrintMessages() {
	l.messages.Print()
}

func (l *Linker) HasErrors() bool {
	return l.messages.errors > 0
}

func (l *Linker) Code() []byte {
	return l.code[0 : l.pc+1]
}

func (l *Linker) DebugInfo() []DebugInfo {
	return l.debugInfo
}

func (l *Linker) Symbols() *SymbolTable {
	return l.symbols
}

func (l *Linker) Messages() *Messages {
	return l.messages
}

func (l *Linker) overrideFramePointerSymbols(ins *InstructionStatement) {
	for _, op := range ins.operands {
		if op.mode == machine.Absolute {
			if op.expr == nil {
				fmt.Printf("nil expr from %s:%d in op %s\n", ins.file, ins.line, ins.operation)
			}
			if op.expr.hasFramePointerSymbols(l.symbols) {
				op.mode = machine.Relative
			}
		} else if op.mode == machine.Indirect {
			if op.expr.hasFramePointerSymbols(l.symbols) {
				op.mode = machine.RelativeIndirect
			}
		}
	}
}

func (l *Linker) isErrorDefined(stmt Statement, id string) bool {
	s := l.symbols.GetSymbol(id)
	if s != nil {
		l.errorf(stmt, "attempt to redefine '%s', already defined by %s:%d", s.text, s.file, s.line)
		return true
	}
	return false
}

func tokToOp(tok TokenType) machine.OpCode {
	var op machine.OpCode
	switch tok {
	case TokAdd:
		op = machine.Add
	case TokSub:
		op = machine.Sub
	case TokMul:
		op = machine.Mul
	case TokDiv:
		op = machine.Div
	case TokCmp:
		op = machine.Cmp
	case TokAnd:
		op = machine.And
	case TokOr:
		op = machine.Or
	case TokXor:
		op = machine.Xor
	case TokCpy:
		op = machine.Cpy
	case TokPsh:
		op = machine.Psh
	case TokPop:
		op = machine.Pop
	case TokInc:
		op = machine.Inc
	case TokDec:
		op = machine.Dec
	case TokJmp:
		op = machine.Jmp
	case TokJeq:
		op = machine.Jeq
	case TokJne:
		op = machine.Jne
	case TokJge:
		op = machine.Jge
	case TokJlt:
		op = machine.Jlt
	case TokJcc:
		op = machine.Jcc
	case TokJcs:
		op = machine.Jcs
	case TokJsr:
		op = machine.Jsr
	case TokSec:
		op = machine.Sec
	case TokClc:
		op = machine.Clc
	case TokSeb:
		op = machine.Seb
	case TokClb:
		op = machine.Clb
	case TokRet:
		op = machine.Ret
	case TokRst:
		op = machine.Rst
	case TokHlt:
		op = machine.Hlt
	case TokSav:
		op = machine.Sav
	case TokSea:
		op = machine.Sea
	default:
		panic("unknown opcode")
	}
	return op
}
