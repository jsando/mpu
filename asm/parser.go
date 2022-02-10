package asm

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	messages      *Messages
	lexer         *Lexer
	pc            int
	first         *Fragment
	last          *Fragment
	label         string   // most recent global label
	pendingLabels []string // labels to be assigned to the next fragment
}

func NewParser(lex *Lexer) *Parser {
	return &Parser{
		messages: &Messages{},
		lexer:    lex,
	}
}

// Parse the input.
func (p *Parser) Parse() {
	lexer := p.lexer
	tok := lexer.Next()
	for {
		// Skip blank lines or lines with only comments
		tok = lexer.syncNextStmt()
		if tok == TokEOF {
			break
		}

		// Local label definition?
		if tok == TokDot {
			tok = lexer.Next()
			if tok != TokIdent {
				p.errorf("expected identifier, got: %s", tok)
				lexer.skipToEOL()
				continue
			}
			// define local label, associated to most recent global label
			p.defineLocalLabel(lexer.s.TokenText())
			lexer.Next()
			continue
		}

		if tok == TokIdent {
			// label if followed by ':' (else it's a directive/opcode)
			if lexer.s.Peek() == ':' {
				p.defineLabel(lexer.s.TokenText())
				lexer.s.Next()
				lexer.Next()
				continue
			}

			// must be a keyword
			text := lexer.s.TokenText()
			tok = toKeyword(text)
			if tok == TokIdent {
				p.errorf("expected directive or opcode, got: %s", text)
				lexer.skipToEOL()
				continue
			}
			fragment := p.newFragment(tok)
			p.lexer.Next()
			operands := p.parseOperands()
			fragment.operands = operands
		}
	}
}

// newFragment creates a new fragment with all pending labels and links it
// into the linked list.
func (p *Parser) newFragment(operation TokenType) *Fragment {
	fragment := &Fragment{
		file:      p.lexer.s.Filename,
		line:      p.lexer.line,
		labels:    p.pendingLabels,
		operation: operation,
	}
	if p.first == nil {
		p.first = fragment
		p.last = fragment
	} else {
		p.last.next = fragment
		p.last = fragment
	}
	p.pendingLabels = []string{}
	return fragment
}

func (p *Parser) defineLabel(label string) {
	p.label = label
	p.pendingLabels = append(p.pendingLabels, label)
}

func (p *Parser) defineLocalLabel(text string) {
	if len(p.label) == 0 {
		p.errorf("can't define local '%s', no global in scope", text)
	}
	p.pendingLabels = append(p.pendingLabels, p.label+"."+text)
}

func (p *Parser) errorf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Error(p.lexer.s.Filename, p.lexer.s.Line, p.lexer.s.Column, s)
}

func (p *Parser) warnf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Warn(p.lexer.s.Filename, p.lexer.s.Line, p.lexer.s.Column, s)
}

func (p *Parser) infof(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Info(p.lexer.s.Filename, p.lexer.s.Line, p.lexer.s.Column, s)
}

func (p *Parser) parseOperands() []*Operand {
	var operands []*Operand
	for {
		operand := p.parseOperand()
		operands = append(operands, operand)
		if p.lexer.tok != TokComma {
			break
		}
		p.lexer.Next()
	}
	return operands
}

func (p *Parser) parseOperand() *Operand {
	tok := p.lexer.tok
	mode := AbsoluteMode
	if tok == TokStar {
		mode = IndirectMode
		p.lexer.Next()
	} else if tok == TokHash {
		mode = ImmediateMode
		p.lexer.Next()
	}
	expr := p.parseExpr()
	return &Operand{
		mode: mode,
		expr: expr,
	}
}

func (p *Parser) parseExpr() Expr {
	// MulExpr [ ('+' | '-' | '|' | '^') MulExpr]*
	expr := p.parseMulExpr()
	for isAddOp(p.lexer.tok) {
		op := p.lexer.tok
		p.lexer.Next()
		expr2 := p.parseMulExpr()
		binop := ExprBinary{
			op:    op,
			expr1: expr,
			expr2: expr2,
		}
		expr = binop
	}
	return expr
}

func (p *Parser) parseMulExpr() Expr {
	// UnaryExpr ['*' | '/' | '%' | '<<' | '>>'  UnaryExpr]*
	expr := p.parseUnaryExpr()
	for isMulOp(p.lexer.tok) {
		op := p.lexer.tok
		p.lexer.Next()
		expr2 := p.parseUnaryExpr()
		binop := ExprBinary{
			op:    op,
			expr1: expr,
			expr2: expr2,
		}
		expr = binop
	}
	return expr
}

func (p *Parser) parseUnaryExpr() Expr {
	// ['+' | '-'] PrimaryExpr
	if isUnaryOp(p.lexer.tok) {
		op := p.lexer.tok
		p.lexer.Next()
		expr1 := p.parsePrimaryExpr()
		expr := ExprUnary{
			op:   op,
			expr: expr1,
		}
		return expr
	} else {
		return p.parsePrimaryExpr()
	}
}

//PrimaryExpr :=
//      '(' expr ')'
//    | Identifier
//    | Literal (int, String, Char)
func (p *Parser) parsePrimaryExpr() Expr {
	var expr Expr
	switch p.lexer.tok {
	case TokLeftParen:
		expr = p.parseExpr()
		p.expect(TokRightParent, "expected ')'")
	case TokIdent:
		expr = ExprIdent{ident: p.lexer.s.TokenText()}
	case TokString:
		expr = BytesLiteral{value: []byte(p.lexer.s.TokenText())}
	case TokInt:
		text := p.lexer.s.TokenText()
		var val int64
		var err error
		if strings.HasPrefix(text, "0x") {
			// why the frick does text/scanner recognize ints in octal, hex, etc but not parse them for me????
			// todo: move this shit to the lexer.  also encapsulate the scanner so it don't do p.lexer.s.XXX anymore
			s := text[2:]
			val, err = strconv.ParseInt(s, 16, 32)
			if err != nil {
				p.errorf("invalid integer literal '%s'", text)
			}
		} else {
			val, err = strconv.ParseInt(text, 10, 32)
			if err != nil {
				p.errorf("invalid integer literal '%s'", text)
			}
		}
		expr = IntLiteral{value: int(val)}
	case TokChar:
		expr = IntLiteral{value: int(p.lexer.s.TokenText()[0])}
	default:
		// some kind of error
		p.errorf("expected (expr), identifier, or literal (got %s)", p.lexer.tok)
		expr = IntLiteral{value: 0}
		// todo skip to eol?
	}
	p.lexer.Next()
	return expr
}

func (p *Parser) expect(tokenType TokenType, msg string) {
	if p.lexer.tok != tokenType {
		p.errorf(msg)
	}
}

func (p *Parser) PrintErrors() {
	p.messages.Print()
}

func (p *Parser) HasErrors() bool {
	return p.messages.errors > 0
}

func (p *Parser) Fragments() *Fragment {
	return p.first
}

// todo: use iota marker values and simple range compares for these
func isAddOp(tok TokenType) bool {
	return tok == TokPlus || tok == TokMinus || tok == TokPipe || tok == TokCaret
}

func isMulOp(tok TokenType) bool {
	return tok == TokStar || tok == TokSlash || tok == TokPercent || tok == TokLeftShift || tok == TokRightShift
}

func isUnaryOp(tok TokenType) bool {
	return tok == TokPlus || tok == TokMinus
}
