package asm

import (
	"fmt"
	"github.com/jsando/lilac/machine"
)

type Linker struct {
	fragment *Fragment
	symbols  *SymbolTable
	messages *Messages
	pc       int
	code     []byte
	patches  []patch
}

// patch is a expression with a forward reference to be resoled on pass 2
type patch struct {
	fragment *Fragment
	expr     Expr
	pc       int
	size     int // 1 = byte, 2 = word
}

func NewLinker(fragment *Fragment) *Linker {
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
		case TokAdd, TokSub, TokMul, TokDiv, TokCmp, TokAnd, TokOr, TokXor, TokCpy:
			l.doEmit2Operand(frag)
		case TokPsh, TokPop, TokJmp, TokJeq, TokJne, TokJge, TokJlt:
			l.doEmit1Operand(frag)
		}
		frag.pcEnd = l.pc
	}
	for _, patch := range l.patches {
		ival, _, res := patch.expr.computeValue(l.symbols)
		if res {
			if patch.size == 1 {
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

func (l *Linker) addPatch(frag *Fragment, expr Expr, size int) {
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

func (l *Linker) defineLabels(frag *Fragment) {
	for _, label := range frag.labels {
		// todo check we aren't redefining global symbol
		l.symbols.AddSymbol(frag.file, frag.line, label)
		l.symbols.Define(frag.labels[0], l.pc)
	}
}

func (l *Linker) errorf(frag *Fragment, format string, args ...interface{}) {
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

func (l *Linker) doEquate(frag *Fragment) {
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

func (l *Linker) doOrg(frag *Fragment) {
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

func (l *Linker) doDefineWord(frag *Fragment) {
	l.defineLabels(frag)
	for _, operand := range frag.operands {
		if operand.mode != AbsoluteMode {
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

func (l *Linker) doDefineByte(frag *Fragment) {
	l.defineLabels(frag)
	for _, operand := range frag.operands {
		if operand.mode != AbsoluteMode {
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

func (l *Linker) doDefineSpace(frag *Fragment) {
	l.defineLabels(frag)
	if len(frag.operands) != 1 {
		l.errorf(frag, "ds requires a single operand")
		return
	}
	operand := frag.operands[0]
	if operand.mode != AbsoluteMode {
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

func (l *Linker) doEmit2Operand(frag *Fragment) {
	l.defineLabels(frag)
	if len(frag.operands) != 2 {
		l.errorf(frag, "expected 2 operands")
		return
	}
	op1 := frag.operands[0]
	op2 := frag.operands[1]
	size := machine.SizeWord
	mode := getMachineMode(op1.mode, op2.mode)
	op := tokToOp(frag.operation)
	opCode := machine.EncodeOp(op, mode, size)
	l.writeByte(int(opCode))
	l.resolveWordOperand(frag, op1)
	l.resolveWordOperand(frag, op2)
}

func (l *Linker) doEmit1Operand(frag *Fragment) {
	l.defineLabels(frag)
	if len(frag.operands) != 1 {
		l.errorf(frag, "expected 1 operand")
		return
	}
	op1 := frag.operands[0]
	size := machine.SizeWord
	mode := getMachineMode(op1.mode, AbsoluteMode)
	op := tokToOp(frag.operation)
	opCode := machine.EncodeOp(op, mode, size)
	l.writeByte(int(opCode))
	l.resolveWordOperand(frag, op1)
}

func (l *Linker) resolveWordOperand(frag *Fragment, op *Operand) {
	ival, bval, res := op.expr.computeValue(l.symbols)
	if res {
		if bval != nil {
			l.errorf(frag, "expected int value, not []byte")
			return
		}
		l.writeWord(ival)
	} else {
		l.addPatch(frag, op.expr, 2)
	}
}

func (l *Linker) BytesFor(frag *Fragment) []byte {
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

func tokToOp(tok TokenType) machine.Opcode {
	var op machine.Opcode
	switch tok {
	case TokAdd:
		op = machine.OpAdd
	case TokSub:
		op = machine.OpSub
	case TokMul:
		op = machine.OpMul
	case TokDiv:
		op = machine.OpDiv
	case TokCmp:
		op = machine.OpCmp
	case TokAnd:
		op = machine.OpAnd
	case TokOr:
		op = machine.OpOr
	case TokXor:
		op = machine.OpXor
	case TokCpy:
		op = machine.OpCpy
	case TokPsh:
		op = machine.OpPsh
	case TokPop:
		op = machine.OpPop
	case TokJmp:
		op = machine.OpJmp
	case TokJeq:
		op = machine.OpJeq
	case TokJne:
		op = machine.OpJne
	case TokJge:
		op = machine.OpJge
	case TokJlt:
		op = machine.OpJlt
	default:
		panic("unknown opcode")
	}
	return op
}

func getMachineMode(op1, op2 AddressMode) machine.OperandMode {
	switch op1 {
	case ImmediateMode:
		return machine.ParamImm
	case AbsoluteMode:
		switch op2 {
		case AbsoluteMode:
			return machine.ParamAbsAbs
		case ImmediateMode:
			return machine.ParamAbsImm
		case IndirectMode:
			return machine.ParamAbsInd
		}
	case IndirectMode:
		switch op2 {
		case AbsoluteMode:
			return machine.ParamIndAbs
		case ImmediateMode:
			return machine.ParamIndImm
		case IndirectMode:
			return machine.ParamIndInd
		}
	}
	panic("invalid mode")
}
