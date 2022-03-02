package asm

import (
	"fmt"
	"strings"

	"github.com/jsando/mpu/machine"
)

type Linker struct {
	fragment *Statement
	symbols  *SymbolTable
	messages *Messages
	function *Statement // Pointer to statement if in function with automatic fp/sav/rst handling
	pc       int
	code     []byte
	patches  []patch
}

// patch is a expression with a forward reference to be resoled on pass 2
type patch struct {
	fragment *Statement
	expr     Expr
	pc       int
	size     int // 1 = byte, 2 = word
}

func NewLinker(fragment *Statement) *Linker {
	return &Linker{
		fragment: fragment,
		symbols:  NewSymbolTable(),
		messages: &Messages{},
		code:     make([]byte, 65536),
	}
}

// Link uses two passes to try to resolve all references and generate code into l.code.
func (l *Linker) Link() {
	for frag := l.fragment; frag != nil; frag = frag.next {
		frag.pcStart = l.pc
		l.overrideFramePointerSymbols(frag)
		switch frag.operation {
		case TokEquals:
			l.doEquate(frag)
		case TokOrg:
			l.doOrg(frag)
		case TokDw:
			l.doDefineWord(frag)
		case TokDb:
			l.doDefineByte(frag)
		case TokDs:
			l.doDefineSpace(frag)
		case TokSec, TokClc, TokSeb, TokClb, TokRet, TokRst, TokHlt:
			l.doEmit0Operand(frag)
		case TokAdd, TokSub, TokMul, TokDiv, TokCmp, TokAnd, TokOr, TokXor, TokCpy:
			l.doEmit2Operand(frag)
		case TokJmp, TokJsr:
			l.doEmitAbsJump(frag)
		case TokJeq, TokJne, TokJge, TokJlt, TokJcc, TokJcs:
			l.doEmitRelJump(frag)
		case TokInc, TokDec, TokPsh, TokPop, TokSav:
			l.doEmit1Operand(frag)
		case TokFunction:
			l.doFunction(frag)
		}
		frag.pcEnd = l.pc
	}
	for _, patch := range l.patches {
		ival, _, res := patch.expr.computeValue(l.symbols)
		if res {
			if patch.size == 1 {
				if patch.fragment.operands[0].mode == machine.OffsetByte {
					ival = ival - patch.pc + 1
				}
				l.writeByteAt(ival, patch.pc)
			} else if patch.size == 2 {
				l.writeWordAt(ival, patch.pc)
			} else {
				panic("invalid patch size")
			}
		} else {
			l.errorf(patch.fragment, "expression still unresolved after second pass")
		}
	}
}

func (l *Linker) addPatch(frag *Statement, expr Expr, size int) {
	p := patch{
		fragment: frag,
		expr:     expr,
		pc:       l.pc,
		size:     size,
	}
	l.patches = append(l.patches, p)
	for i := 0; i < size; i++ {
		l.writeByte(0)
	}
}

func (l *Linker) defineLabels(frag *Statement) {
	// todo check we aren't redefining symbols
	global := false
	for _, label := range frag.labels {
		if l.isErrorDefined(frag, label) {
			return
		}
		l.symbols.AddSymbol(frag.file, frag.line, label)
		l.symbols.Define(label, l.pc)
		if !strings.ContainsAny(label, ".") {
			global = true
		}
	}
	if global {
		l.function = nil
	}
}

func (l *Linker) errorf(frag *Statement, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.messages.Error(frag.file, frag.line, 0, msg)
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

func (l *Linker) doEquate(frag *Statement) {
	if len(frag.labels) != 1 {
		l.errorf(frag, "equate must have exactly one label")
	}
	if l.isErrorDefined(frag, frag.labels[0]) {
		return
	}
	l.symbols.AddSymbol(frag.file, frag.line, frag.labels[0])
	ival, _, res := frag.operands[0].expr.computeValue(l.symbols) // todo equate must have 1 arg
	if !res {
		l.errorf(frag, "equate must have constant value")
	}
	l.symbols.Define(frag.labels[0], ival)
}

func (l *Linker) doOrg(frag *Statement) {
	// org bumps to PC to a new location.  Error to go backwards.
	ival, _, res := frag.operands[0].expr.computeValue(l.symbols) // todo equate must have 1 arg
	if !res {
		l.errorf(frag, "org must have constant value")
	} else {
		if l.pc > ival {
			l.errorf(frag, "pc already at %d, can't set back to %d", l.pc, ival)
		} else {
			l.pc = ival
		}
	}
}

func (l *Linker) doDefineWord(frag *Statement) {
	l.defineLabels(frag)
	for _, operand := range frag.operands {
		if operand.mode != machine.Absolute {
			l.errorf(frag, "illegal operand mode for dw")
			return
		}
		ival, bval, res := operand.expr.computeValue(l.symbols)
		if res {
			if bval != nil {
				// todo: does it even make sense to write a string in lo/hi byte order???? I say no.
				l.errorf(frag, "can't use string with dw")
			} else {
				l.writeWord(ival)
			}
		} else {
			l.addPatch(frag, operand.expr, 2)
		}
	}
}

func (l *Linker) doDefineByte(frag *Statement) {
	l.defineLabels(frag)
	for _, operand := range frag.operands {
		if operand.mode != machine.Absolute {
			l.errorf(frag, "illegal operand mode for db")
			return
		}
		ival, bval, res := operand.expr.computeValue(l.symbols)
		if res {
			if bval != nil {
				l.writeBytes(bval)
			} else {
				l.writeByte(ival)
			}
		} else {
			l.addPatch(frag, operand.expr, 1)
		}
	}
}

func (l *Linker) doDefineSpace(frag *Statement) {
	l.defineLabels(frag)
	if len(frag.operands) != 1 {
		l.errorf(frag, "ds requires a single operand")
		return
	}
	operand := frag.operands[0]
	if operand.mode != machine.Absolute {
		l.errorf(frag, "invalid operand for ds")
		return
	}
	ival, bval, res := operand.expr.computeValue(l.symbols)
	if !res {
		l.errorf(frag, "illegal forward reference in ds")
		return
	}
	if bval != nil {
		l.errorf(frag, "ds requires int param")
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
func (l *Linker) doFunction(frag *Statement) {
	// define any labels, esp the global label for this function.
	// it will emit a SAV at this address to make space for locals.
	l.defineLabels(frag)
	l.function = frag

	if len(frag.labels) != 1 {
		l.errorf(frag, "function must have only one associated label: %v", frag.labels)
	}

	// compute the fp offset for all params and locals and
	// add them to the symbol table
	offset := 0
	localSize := 0
	for _, local := range frag.fpLocals {
		localSize += local.size
		offset -= local.size
		local.offset = offset
		if offset < -128 {
			l.errorf(frag, "fp local offset out of range (-1,-128): %d", offset)
		}
		l.symbols.AddFpSymbol(frag.file, frag.line, local.id, local.offset)
	}
	offset = 4
	for i := len(frag.fpArgs) - 1; i >= 0; i-- {
		arg := frag.fpArgs[i]
		arg.offset = offset
		if offset > 127 {
			l.errorf(frag, "fp arg offset out of range (0,127): %d", offset)
		}
		l.symbols.AddFpSymbol(frag.file, frag.line, arg.id, arg.offset)
		offset += arg.size
	}
	opCode := machine.EncodeOp(machine.Sav, machine.ImmediateByte, machine.Implied)
	l.writeByte(int(opCode))
	l.writeByte(localSize)
}

func (l *Linker) doEmit2Operand(frag *Statement) {
	l.defineLabels(frag)
	if len(frag.operands) != 2 {
		l.errorf(frag, "expected 2 operands")
		return
	}
	op1 := frag.operands[0]
	op2 := frag.operands[1]
	op := tokToOp(frag.operation)
	opCode := machine.EncodeOp(op, op1.mode, op2.mode)
	l.writeByte(int(opCode))
	l.resolveWordOperand(frag, op1)
	l.resolveWordOperand(frag, op2)
}

func (l *Linker) doEmit1Operand(frag *Statement) {
	l.defineLabels(frag)
	if len(frag.operands) != 1 {
		l.errorf(frag, "expected 1 operand")
		return
	}
	op1 := frag.operands[0]
	op := tokToOp(frag.operation)
	if l.function != nil && op == machine.Sav {
		l.errorf(frag, "within functions, asm generates automatic SAV")
		return
	}
	if op == machine.Pop && op1.mode == machine.Immediate {
		op1.mode = machine.ImmediateByte
	}
	opCode := machine.EncodeOp(op, op1.mode, machine.Implied)
	l.writeByte(int(opCode))
	l.resolveWordOperand(frag, op1)
}

func (l *Linker) doEmitAbsJump(frag *Statement) {
	if len(frag.operands) != 1 {
		l.errorf(frag, "expected 1 operand")
		return
	}
	// override so we don't need # on all jumps
	frag.operands[0].mode = machine.Immediate
	l.doEmit1Operand(frag)
}

func (l *Linker) doEmitRelJump(frag *Statement) {
	if len(frag.operands) != 1 {
		l.errorf(frag, "expected 1 operand")
		return
	}
	frag.operands[0].mode = machine.OffsetByte
	l.doEmit1Operand(frag)
}

func (l *Linker) doEmit0Operand(frag *Statement) {
	l.defineLabels(frag)
	if len(frag.operands) != 0 {
		l.errorf(frag, "expected 0 operands, got %d", len(frag.operands))
		return
	}
	op := tokToOp(frag.operation)
	if l.function != nil && op == machine.Ret {
		op = machine.Rst
	}
	opCode := machine.EncodeOp(op, machine.Implied, machine.Implied)
	l.writeByte(int(opCode))
}

func (l *Linker) resolveWordOperand(frag *Statement, op *Operand) {
	// technically I want ImmediateByte to be an unsigned int but ... ok for now to make them all int8
	byteOp := false
	switch op.mode {
	case machine.ImmediateByte, machine.OffsetByte, machine.Relative, machine.RelativeIndirect:
		byteOp = true
	}
	ival, bval, res := op.expr.computeValue(l.symbols)
	if res {
		if bval != nil {
			l.errorf(frag, "expected int value, not []byte")
			return
		}
		if byteOp {
			if op.mode == machine.OffsetByte {
				ival = ival - l.pc + 1
			}
			l.writeByte(ival)
		} else {
			l.writeWord(ival)
		}
	} else {
		if byteOp {
			l.addPatch(frag, op.expr, 1)
		} else {
			l.addPatch(frag, op.expr, 2)
		}
	}
}

func (l *Linker) BytesFor(frag *Statement) []byte {
	return l.code[frag.pcStart:frag.pcEnd]
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

func (l *Linker) overrideFramePointerSymbols(frag *Statement) {
	for _, op := range frag.operands {
		if op.mode == machine.Absolute {
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

func (l *Linker) isErrorDefined(stmt *Statement, id string) bool {
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
	default:
		panic("unknown opcode")
	}
	return op
}
