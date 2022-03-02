package asm

import (
	"fmt"
	"github.com/jsando/mpu/machine"
	"strconv"
)

type Parser struct {
	messages      *Messages
	lexer         TokenReader
	pc            int
	first         *Statement
	last          *Statement
	label         string   // most recent global label
	pendingLabels []string // labels to be assigned to the next fragment
}

func NewParser(lex TokenReader) *Parser {
	return &Parser{
		messages: &Messages{},
		lexer:    lex,
	}
}

func (p *Parser) syncNextStmt() TokenType {
	for p.lexer.Token() == TokEOL {
		p.lexer.Next()
	}
	return p.lexer.Token()
}

// skipToEOL skips all tokens up to the next EOL, useful in error recovery.
func (p *Parser) skipToEOL() {
	l := p.lexer
	for l.Token() != TokEOL && l.Token() != TokEOF {
		l.Next()
	}
}

func toKeyword(ident string) TokenType {
	for i, image := range tokenImage {
		if ident == image {
			return TokenType(i)
		}
	}
	return TokIdent
}

// Parse the input.
func (p *Parser) Parse() {
	lexer := p.lexer
	tok := lexer.Next()
loop:
	for {
		tok = p.syncNextStmt()
		switch tok {
		case TokEOF:
			break loop
		case TokDot:
			p.parseLocalLabel()
		case TokIdent:
			text := lexer.TokenText()
			tok = toKeyword(text)
			if tok == TokIdent {
				p.parseLabel()
			} else {
				fragment := p.newFragment(tok)
				p.lexer.Next()
				operands := p.parseOperands()
				fragment.operands = operands
			}
		default:
			p.errorf("unexpected: %s", p.lexer.Token())
			break loop
		}
	}
}

func (p *Parser) parseLabel() {
	text := p.lexer.TokenText()
	tok := p.lexer.Next()
	if tok == TokEquals {
		p.lexer.Next()
		fragment := p.newEquate(text)
		operands := p.parseOperands()
		fragment.operands = operands
	} else if tok == TokColon {
		p.defineLabel(text)
		p.lexer.Next()
	} else if tok == TokLeftParen {
		p.parseFunctionDecl(text)
	} else {
		p.errorf("expected =, :, or (): after identifier '%s'", text)
	}
}

func (p *Parser) parseFunctionDecl(fnName string) {
	p.defineLabel(fnName)
	p.lexer.Next()
	frag := p.newFragment(TokFunction)
	for p.lexer.Token() != TokRightParen {
		if p.lexer.Token() != TokIdent {
			p.errorf("expected: identifier, got: %s", p.lexer.Token())
			p.skipToEOL()
			return
		}
		id := fnName + "." + p.lexer.TokenText()
		tok := p.lexer.Next()
		if tok != TokIdent {
			p.errorf("expected: identifier, got: %s", p.lexer.Token())
			p.skipToEOL()
			return
		}
		var size int
		sizeText := p.lexer.TokenText()
		if sizeText == "word" {
			size = 2
		} else if sizeText == "byte" {
			size = 1
		} else {
			p.errorf("expected 'word' or 'byte', got: %s", sizeText)
			p.skipToEOL()
			return
		}
		// ok we have id, size ... wtf do we do now
		frag.AddFpArg(id, size)
		tok = p.lexer.Next()
		if tok != TokComma {
			break
		}
		p.lexer.Next()
	}
	tok := p.lexer.Next()
	if tok != TokColon {
		p.errorf("expected ':', got: %s", p.lexer.TokenText())
	}
	p.lexer.Next()
}

func (p *Parser) parseLocalLabel() {
	tok := p.lexer.Next()
	if tok != TokIdent {
		p.errorf("expected identifier, got: %s", tok)
		p.skipToEOL()
		return
	}
	tokenText := p.lexer.TokenText()
	p.lexer.Next()

	// If followed by '=', its a local equate
	if p.lexer.Token() == TokEquals {
		p.lexer.Next()
		fragment := p.newEquate(p.label + "." + tokenText)
		operands := p.parseOperands()
		fragment.operands = operands
	} else if p.lexer.Token() == TokIdent && p.lexer.TokenText() == "local" {
		// if followed by 'local' it's a local decl ie fp relative
		if p.last.operation != TokFunction {
			p.errorf("'local' can only be used immediately after function declaration")
			return
		}
		tok = p.lexer.Next()
		if tok != TokIdent {
			p.errorf("expected: identifier, got: %s", p.lexer.Token())
			p.skipToEOL()
			return
		}
		id := p.label + "." + tokenText
		var size int
		sizeText := p.lexer.TokenText()
		if sizeText == "word" {
			size = 2
		} else if sizeText == "byte" {
			size = 1
		} else {
			p.errorf("expected 'word' or 'byte', got: %s", sizeText)
			p.skipToEOL()
			return
		}
		// ok we have id, size ... wtf do we do now
		p.last.AddFpLocal(id, size)
		p.lexer.Next()
	} else {
		// define local label, associated to most recent global label
		p.defineLocalLabel(tokenText)
	}
}

// newFragment creates a new fragment with all pending labels and links it
// into the linked list.
func (p *Parser) newFragment(operation TokenType) *Statement {
	fragment := &Statement{
		file:      p.lexer.FileName(),
		line:      p.lexer.Line(),
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

// newFragment creates a new fragment with all pending labels and links it
// into the linked list.
func (p *Parser) newEquate(id string) *Statement {
	fragment := &Statement{
		file:      p.lexer.FileName(),
		line:      p.lexer.Line(),
		labels:    []string{id},
		operation: TokEquals,
	}
	if p.first == nil {
		p.first = fragment
		p.last = fragment
	} else {
		p.last.next = fragment
		p.last = fragment
	}
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
	p.messages.Error(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) warnf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Warn(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) infof(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Info(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) parseOperands() []*Operand {
	var operands []*Operand
	for {
		operand := p.parseOperand()
		if operand == nil {
			return operands
		}
		operands = append(operands, operand)
		if p.lexer.Token() != TokComma {
			break
		}
		p.lexer.Next()
	}
	return operands
}

func (p *Parser) parseOperand() *Operand {
	tok := p.lexer.Token()
	if tok == TokEOL {
		return nil
	}
	mode := machine.Absolute
	match := TokNone
	if tok == TokStar {
		tok = p.lexer.Next()
		if tok == TokLeftBracket {
			p.lexer.Next() // consumer 'fp'
			p.expect(TokIdent)
			text := p.lexer.TokenText()
			if text != "fp" {
				p.errorf("expected 'fp', got %s", text)
			}
			if p.messages.errors > 0 {
				return nil
			}
			p.lexer.Next()
			match = TokRightBracket
			mode = machine.RelativeIndirect
		} else {
			mode = machine.Indirect
		}
	} else if tok == TokHash {
		mode = machine.Immediate
		p.lexer.Next()
	} else if tok == TokLeftBracket {
		match = TokRightBracket
		mode = machine.Relative
		p.lexer.Next()
		p.expect(TokIdent)
		text := p.lexer.TokenText()
		if text != "fp" {
			p.errorf("expected 'fp', got %s", text)
		}
		if p.messages.errors > 0 {
			return nil
		}
		p.lexer.Next()
	}
	expr := p.parseExpr()
	if match != TokNone {
		p.expect(match)
		p.lexer.Next()
	}
	return &Operand{
		mode: mode,
		expr: expr,
	}
}

func (p *Parser) parseExpr() Expr {
	// MulExpr [ ('+' | '-' | '|' | '^') MulExpr]*
	expr := p.parseMulExpr()
	for isAddOp(p.lexer.Token()) {
		op := p.lexer.Token()
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
	for isMulOp(p.lexer.Token()) {
		op := p.lexer.Token()
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
	if isUnaryOp(p.lexer.Token()) {
		op := p.lexer.Token()
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
	switch p.lexer.Token() {
	case TokLeftParen:
		p.lexer.Next()
		expr = p.parseExpr()
		p.expect(TokRightParen)
		p.lexer.Next()
	case TokIdent:
		expr = ExprIdent{ident: p.lexer.TokenText(), activeLabel: p.label}
	case TokString:
		s, err := strconv.Unquote(p.lexer.TokenText())
		if err != nil {
			p.errorf("invalid string literal")
		}
		expr = BytesLiteral{value: []byte(s)}
	case TokInt:
		// strconv.ParseInt, if passed bitSize=0, will use Go's syntax for literals
		// such as 0b, 0x, underscores, etc.
		// ParseInt assumes signed types so have to pass bitSize=32 otherwise
		// would error on uint16 > 32767.
		val, err := strconv.ParseInt(p.lexer.TokenText(), 0, 32)
		if err != nil {
			p.errorf("invalid integer literal '%s'", p.lexer.TokenText())
		}
		expr = IntLiteral{value: int(val)}
	case TokChar:
		expr = IntLiteral{value: int(p.lexer.TokenText()[1])}
	default:
		// some kind of error
		p.errorf("expected (expr), identifier, or literal (got %s)", p.lexer.Token())
		expr = IntLiteral{value: 0}
		// todo skip to eol?
	}
	p.lexer.Next()
	return expr
}

func (p *Parser) expect(tokenType TokenType) {
	if p.lexer.Token() != tokenType {
		p.errorf("expected: %s, got: %s", tokenType, p.lexer.TokenText())
	}
}

func (p *Parser) PrintErrors() {
	p.messages.Print()
}

func (p *Parser) HasErrors() bool {
	return p.messages.errors > 0
}

func (p *Parser) Fragments() *Statement {
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
