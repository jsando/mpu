package asm

import (
	"fmt"
	"github.com/jsando/mpu/machine"
	"io"
	"strings"
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

func (p *Printer) printf(format string, a ...interface{}) {
	n, err := fmt.Fprintf(p.w, format, a...)
	if err != nil {
		panic(err)
	}
	p.column += n
}

func (p *Printer) newline() {
	fmt.Fprintln(p.w)
	p.column = 0
}

func (p *Printer) Print(stmt *Statement) {
	for stmt != nil {
		if p.column != 0 {
			p.newline()
		}
		if stmt.newlineBefore {
			p.newline()
		}

		// block comments on functions or global equates or labels are aligned to left margin
		// otherwise in the op column
		commentIndent := !isGlobal(stmt)
		for _, c := range stmt.blockComment {
			if commentIndent {
				p.tab(OpColumn)
			}
			p.printf(";%s", c)
			p.newline()
		}
		for _, lbl := range stmt.labels {
			// todo why is this in a for loop??? equates too?
			if stmt.operation == TokFunction {
				p.printf("%s(", lbl)
				count := 0
				for _, arg := range stmt.fpArgs {
					if count != 0 {
						p.printf(", ")
					}
					p.printf("%s %s", toLocal(arg.id), wordOrByte(arg.size))
					count++
				}
				p.printf("):") // todo could have line comment
				p.newline()
				for _, arg := range stmt.fpLocals {
					p.tab(OpColumn)
					p.printf("var %s %s", toLocal(arg.id), wordOrByte(arg.size))
					p.newline()
				}
			} else if stmt.operation == TokEquals {
				local := toLocal(lbl)
				if local != lbl {
					local = "." + local
				}
				p.printf("%s", local)
			} else {
				local := toLocal(lbl)
				if local != lbl {
					p.printf(".%s", local)
				} else {
					p.printf("%s:", local)
				}
				if stmt.newlineBefore {
					p.newline()
				}
				if p.column >= OpColumn-1 {
					p.newline()
				}
			}
		}
		if stmt.operation != TokFunction {
			p.tab(OpColumn)
			p.stmt(stmt)
			if len(stmt.eolComment) > 0 {
				p.tab(CommentColumn)
				p.printf(";%s", stmt.eolComment)
			}
			p.newline()
		}
		stmt = stmt.next
	}
}

func isGlobal(stmt *Statement) bool {
	if stmt.operation == TokFunction {
		return true
	}
	for _, lbl := range stmt.labels {
		loc := toLocal(lbl)
		if loc == lbl {
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

func (p *Printer) stmt(stmt *Statement) {
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
		if strings.HasPrefix(text, "0x") {
			text = "$" + text[2:]
		}
		if strings.HasPrefix(text, "0b") {
			text = "%" + text[2:]
		}
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
