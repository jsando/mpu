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
	"io"
	"strings"

	"github.com/jsando/mpu/machine"
)

// Printer pretty-prints statements to a writer.
type Printer struct {
	w      io.Writer
	column int
}

const (
	OpColumn      = 16
	CommentColumn = 32
)

func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

func (p *Printer) Print(stmt Statement) {
	emitln := true
	header := true
	for stmt != nil {
		if p.column != 0 && emitln {
			p.newline()
		}
		emitln = true
		if stmt.NewlineBefore() {
			p.newline()
		}
		// block comments on functions or global equates or labels are aligned to left margin
		// otherwise in the op column
		commentIndent := !isGlobal(stmt) && !header
		for _, c := range stmt.BlockComment() {
			if commentIndent {
				p.tab(OpColumn)
			}
			p.printf("//%s", c)
			p.newline()
		}
		header = false
		switch t := stmt.(type) {
		case *EquateStatement:
			id := toLocal(t.name)
			if id != t.name {
				id = "." + id
			}
			p.print(id)
			p.tab(OpColumn)
			p.print("= ")
			p.expr(t.value)
		case *DefineByteStatement:
			p.tab(OpColumn)
			p.print("db ")
			p.exprList(t.values)
		case *DefineWordStatement:
			p.tab(OpColumn)
			p.print("dw ")
			p.exprList(t.values)
		case *DefineSpaceStatement:
			p.tab(OpColumn)
			p.print("ds ")
			p.expr(t.size)
		case *ImportStatement:
			p.tab(OpColumn)
			p.printf("import \"%s\"", t.path)
		case *OrgStatement:
			p.tab(OpColumn)
			p.print("org ")
			p.expr(t.origin)
		case *LabelStatement:
			next := stmt.Next()
			_, ok := next.(*InstructionStatement)
			if !ok {
				emitln = false
			}
			p.label(t.name)
		case *FunctionStatement:
			p.printf("%s(", t.name)
			count := 0
			for _, arg := range t.fpArgs {
				if count != 0 {
					p.printf(", ")
				}
				p.printf("%s %s", toLocal(arg.id), wordOrByte(arg.size))
				count++
			}
			p.printf("):")
		case *VarStatement:
			p.tab(OpColumn)
			p.printf("var %s %s", toLocal(t.name), wordOrByte(t.size))
		case *InstructionStatement:
			p.tab(OpColumn)
			p.stmt(t)
		default:
			panic("unknown statement type")
		}
		p.comment(stmt)
		stmt = stmt.Next()
	}
	p.newline()
}

func (p *Printer) printf(format string, a ...interface{}) {
	n, err := fmt.Fprintf(p.w, format, a...)
	if err != nil {
		panic(err)
	}
	p.column += n
}

func (p *Printer) print(s string) {
	n, err := fmt.Fprint(p.w, s)
	if err != nil {
		panic(err)
	}
	p.column += n
}

func (p *Printer) label(label string) {
	if p.column != 0 {
		p.newline()
	}
	loc := toLocal(label)
	if loc == label {
		loc += ":"
	} else {
		loc = "." + loc
	}
	p.print(loc)
}

func (p *Printer) labelOp(label, op string) {
	p.label(label)
	p.tab(OpColumn)
	p.print(op)
	p.print(" ")
}

func (p *Printer) newline() {
	fmt.Fprintln(p.w)
	p.column = 0
}

func (p *Printer) comment(s Statement) {
	c := s.EolComment()
	if len(c) > 0 {
		p.tab(CommentColumn)
		p.printf("//%s", c)
	}
}

func isGlobal(stmt Statement) bool {
	switch t := stmt.(type) {
	case *FunctionStatement:
		return true
	case *EquateStatement:
		return true
	case *LabelStatement:
		loc := toLocal(t.name)
		if loc == t.name {
			return true
		}
	}
	return false
}

func wordOrByte(size int) string {
	switch size {
	case 1:
		return "byte"
	case 2:
		return "word"
	default:
		panic(fmt.Sprintf("invalid arg size %d", size))
	}
}

func toLocal(id string) string {
	s := strings.Split(id, ".")
	if len(s) == 2 {
		return s[1]
	}
	return id
}

func (p *Printer) tab(column int) {
	for {
		p.printf(" ")
		if p.column >= column {
			break
		}
	}
}

func (p *Printer) stmt(stmt *InstructionStatement) {
	p.printf("%s ", stmt.operation)
	count := 0
	for _, op := range stmt.operands {
		if count > 0 {
			p.printf(",")
		}
		switch op.mode {
		case machine.Implied:
		case machine.Absolute:
		case machine.Immediate:
			p.printf("#")
		case machine.ImmediateByte:
		case machine.OffsetByte:
		case machine.Indirect:
			p.printf("*")
		case machine.Relative:
		case machine.RelativeIndirect:
			p.printf("*")
		}
		p.expr(op.expr)
		count++
	}
}

func (p *Printer) expr(expr Expr) {
	switch e := expr.(type) {
	case CharLiteral:
		p.printf("%s", e.text)
	case IntLiteral:
		text := e.text
		//if strings.HasPrefix(text, "0x") {
		//	text = "$" + text[2:]
		//}
		//if strings.HasPrefix(text, "0b") {
		//	text = "%" + text[2:]
		//}
		p.printf("%s", text)
	case ExprUnary:
		p.printf("%s", e.op)
		p.expr(e.expr)
	case ExprBinary:
		p.expr(e.expr1)
		p.printf("%s", e.op)
		p.expr(e.expr2)
	case BytesLiteral:
		p.printf("%s", e.text)
	case ExprIdent:
		p.printf("%s", e.ident)
	}
}

func (p *Printer) exprList(values []Expr) {
	count := 0
	for _, expr := range values {
		if count > 0 {
			p.print(",")
		}
		p.expr(expr)
		count++
	}
}
