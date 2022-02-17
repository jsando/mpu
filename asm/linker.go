package asm

import (
	"fmt"
	"github.com/jsando/mpu/machine"
)

type Linker struct {
	fragment *Statement
	symbols  *SymbolTable
	messages *Messages
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
		case TokSec, TokClc, TokSeb, TokClb, TokRet:
			l.doEmit0Operand(frag)
		case TokAdd, TokSub, TokMul, TokDiv, TokCmp, TokAnd, TokOr, TokXor, TokCpy:
			l.doEmit2Operand(frag)
		case TokJmp, TokJsr:
			l.doEmitAbsJump(frag)
		case TokJeq, TokJne, TokJge, TokJlt, TokJcc, TokJcs:
			l.doEmitRelJump(frag)
		case TokInc, TokDec, TokPsh, TokPop:
			l.doEmit1Operand(frag)
		}
		frag.pcEnd = l.pc
	}
	for _, patch := range l.patches {
		ival, _, res := patch.expr.computeValue(l.symbols)
		if res {
			if patch.size == 1 {
				if patch.fragment.operands[0].mode == machine.OffsetByte {
					ival = patch.pc - ival
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
	for _, label := range frag.labels {
		// todo check we aren't redefining global symbol
		l.symbols.AddSymbol(frag.file, frag.line, label)
		l.symbols.Define(frag.labels[0], l.pc)
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
				ival = l.pc - ival
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
	case TokSav:
		op = machine.Sav
	default:
		panic("unknown opcode")
	}
	return op
}
