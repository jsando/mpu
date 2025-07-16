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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jsando/mpu/machine"
)

type Parser struct {
	messages       *Messages
	lexer          *Input
	pc             int
	first          Statement
	last           Statement
	eolCount       int                // to remember intentional blank lines for reformat
	comments       []string           // pending block comments for next statement
	processInclude bool               // If just formatting don't automatically append includes
	function       *FunctionStatement // Current function to append vars to or nil
	global         string             // Most recent global (function or label), to define locals against
}

func NewParser(lex *Input) *Parser {
	return &Parser{
		messages:       &Messages{},
		lexer:          lex,
		processInclude: true,
	}
}

func NewParserFromReader(name string, r io.Reader) *Parser {
	return NewParser(NewInput([]TokenReader{NewLexer(name, r)}))
}

func (p *Parser) SetProcessInclude(process bool) {
	p.processInclude = process
}

// Files returns the list of all files processed.
func (p *Parser) Files() []string {
	return p.lexer.Files()
}

func (p *Parser) PrintErrors() {
	p.messages.Print()
}

func (p *Parser) HasErrors() bool {
	return p.messages.errors > 0
}

func (p *Parser) Statements() Statement {
	return p.first
}

func (p *Parser) Messages() *Messages {
	return p.messages
}

func (p *Parser) addStatement(s Statement) {
	s.SetFile(p.lexer.FileName())
	s.SetLine(p.lexer.Line())
	comments := cleanComments(p.comments)
	s.SetBlockComment(comments)
	commentLines := len(comments)
	eolCount := p.eolCount - commentLines
	s.SetNewlineBefore(eolCount > 1)

	p.comments = nil
	p.eolCount = 0
	if p.first == nil {
		p.first = s
		p.last = s
	} else {
		p.last.SetNext(s)
		p.last = s
	}
}

func cleanComments(s []string) []string {
	clean := []string{}
	for _, c := range s {
		c = strings.TrimPrefix(c, "//")
		c = strings.TrimPrefix(c, "/*")
		c = strings.TrimSuffix(c, "*/")
		sa := strings.Split(c, "\n")
		for _, l := range sa {
			clean = append(clean, l)
		}
	}
	return clean
}

func (p *Parser) errorf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Error(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) errorAtf(line int, format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Error(p.lexer.FileName(), line, 0, s)
}

func (p *Parser) warnf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Warn(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) infof(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	p.messages.Info(p.lexer.FileName(), p.lexer.Line(), p.lexer.Column(), s)
}

func (p *Parser) expect(tokenType TokenType) {
	if p.lexer.Token() != tokenType {
		p.errorf("expected: %s, got: %s", tokenType, p.lexer.TokenText())
	}
}

func (p *Parser) syncNextStmt() TokenType {
skipLoop:
	for {
		switch p.lexer.Token() {
		case TokEOL:
			p.eolCount++
			p.lexer.Next()
		case TokComment:
			if p.eolCount == 0 && p.last != nil {
				block := cleanComments([]string{p.lexer.TokenText()})
				if len(block) > 1 {
					panic("wanna hear the most annoying sound in the world")
				}
				p.last.SetEolComment(block[0])
			} else {
				p.comments = append(p.comments, p.lexer.TokenText())
			}
			p.lexer.Next()
		default:
			break skipLoop
		}
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
		// Skip of EOL and Comment but buffer them up to attach to
		// the following statement.  Line comments are attached to
		// the last statement.
		tok = p.syncNextStmt()

		// Parse based on token.
		switch tok {
		case TokEOF:
			break loop
		case TokDot:
			tok = p.lexer.Next()
			p.parseLocalSymbol()
		case TokIdent:
			tok = toKeyword(lexer.TokenText())
			switch tok {
			case TokIdent:
				p.parseGlobalSymbol() // global label, equate, or function
			case TokInclude:
				p.parseIncludes()
			case TokTest:
				p.parseTestDecl()
			case TokDb:
				p.parseDb()
			case TokDw:
				p.parseDw()
			case TokDs:
				p.parseDs()
			case TokOrg:
				p.parseOrg()
			case TokVar:
				p.parseVar()
			case TokSea:
				p.parseInstruction(tok)
			default:
				// assume it's an instruction
				p.parseInstruction(tok)
			}
		default:
			p.errorf("unexpected: %s", p.lexer.Token())
			break loop
		}
	}
}

func (p *Parser) parseIncludes() {
	stmt := &IncludeStatement{}
	p.addStatement(stmt)
	tok := p.lexer.Next()
	if tok != TokString {
		p.errorf("includes requires string as argument")
		return
	}
	name, err := strconv.Unquote(p.lexer.TokenText())
	if err != nil {
		p.errorf("bad import name '%s'", p.lexer.TokenText())
		return
	}
	stmt.path = name

	// When running fmt we don't want to append all included files to the one original file :)
	// Just need the AST to reformat the code.
	if p.processInclude {
		dir := filepath.Dir(p.lexer.FileName())
		name = filepath.Join(dir, name)
		file, err := os.Open(name)
		if err != nil {
			p.errorf("error opening file '%s': %s\n", name, err.Error())
			return
		}
		r := NewLexer(file.Name(), file)
		p.lexer.Append(r)
	}
	p.lexer.Next()
}

func (p *Parser) parseGlobalSymbol() {
	p.function = nil
	text := p.lexer.TokenText()
	// Save line number before advancing lexer
	line := p.lexer.Line()
	tok := p.lexer.Next()
	if tok == TokEquals {
		p.global = ""
		stmt := &EquateStatement{
			name: text,
		}
		p.addStatement(stmt)
		p.lexer.Next()
		stmt.value = p.parseExpr()
	} else if tok == TokColon {
		p.global = text
		stmt := &LabelStatement{
			name: text,
		}
		p.addStatement(stmt)
		p.lexer.Next()
	} else if tok == TokLeftParen {
		p.global = text
		p.parseFunctionDecl(text)
	} else {
		// Report error at the saved line number
		p.errorAtf(line, "expected =, :, or (): after identifier '%s'", text)
		// Skip to end of line for error recovery
		p.skipToEOL()
	}
}

func (p *Parser) parseFunctionDecl(fnName string) {
	fn := &FunctionStatement{
		name: fnName,
	}
	p.addStatement(fn)
	p.lexer.Next()
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
		fn.fpArgs = append(fn.fpArgs, &FpParam{
			id:     id,
			size:   size,
			offset: 0,
		})
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
	p.function = fn
	p.lexer.Next()
}

func (p *Parser) parseTestDecl() {
	// test keyword already consumed
	if p.lexer.Next() != TokIdent {
		p.errorf("expected test function name after 'test'")
		p.skipToEOL()
		return
	}

	testName := p.lexer.TokenText()

	if p.lexer.Next() != TokLeftParen {
		p.errorf("expected '(' after test function name")
		p.skipToEOL()
		return
	}

	if p.lexer.Next() != TokRightParen {
		p.errorf("expected ')' - test functions take no parameters")
		p.skipToEOL()
		return
	}

	if p.lexer.Next() != TokColon {
		p.errorf("expected ':' after test function declaration")
		p.skipToEOL()
		return
	}

	// Create test statement
	test := &TestStatement{
		name: testName,
	}
	p.addStatement(test)

	// Set this as the current global context
	p.global = testName
	p.function = nil

	p.lexer.Next()
}

func (p *Parser) parseLocalSymbol() {
	if p.lexer.Token() != TokIdent {
		p.errorf("expected identifier, got: %s", p.lexer.Token())
		p.skipToEOL()
		return
	}
	if len(p.global) == 0 {
		p.errorf("locals are local to the nearest global but none is in scope")
		p.skipToEOL()
		return
	}
	text := p.lexer.TokenText()
	line := p.lexer.Line()
	id := p.global + "." + p.lexer.TokenText()
	p.lexer.Next()

	if p.lexer.Token() == TokEquals {
		stmt := &EquateStatement{
			name: id,
		}
		p.addStatement(stmt)
		p.lexer.Next()
		stmt.value = p.parseExpr()
	} else if p.lexer.Token() == TokColon {
		stmt := &LabelStatement{
			name: id,
		}
		p.addStatement(stmt)
		p.lexer.Next()
	} else {
		// Report error at the saved line number
		p.errorAtf(line, "expected '=' or ':' after identifier '%s'", text)
		// Skip to end of line for error recovery
		p.skipToEOL()
	}
}

func (p *Parser) parseDb() {
	stmt := &DefineByteStatement{}
	p.addStatement(stmt)
	p.lexer.Next()
	stmt.values = p.parseExpressionList()
}

func (p *Parser) parseDw() {
	stmt := &DefineWordStatement{}
	p.addStatement(stmt)
	p.lexer.Next()
	stmt.values = p.parseExpressionList()
}

func (p *Parser) parseDs() {
	stmt := &DefineSpaceStatement{}
	p.addStatement(stmt)
	p.lexer.Next()
	stmt.size = p.parseExpr()
}

func (p *Parser) parseOrg() {
	stmt := &OrgStatement{}
	p.addStatement(stmt)
	p.lexer.Next()
	stmt.origin = p.parseExpr()
}

func (p *Parser) parseVar() {
	if p.function == nil {
		p.errorf("var is only valid within function but none is in scope")
		p.skipToEOL()
		return
	}
	stmt := &VarStatement{}
	p.addStatement(stmt)
	tok := p.lexer.Next()
	if tok != TokIdent {
		p.errorf("expected: identifier, got: %s", p.lexer.Token())
		p.skipToEOL()
		return
	}
	stmt.name = p.function.name + "." + p.lexer.TokenText()
	p.lexer.Next()
	sizeText := p.lexer.TokenText()
	if sizeText == "word" {
		stmt.size = 2
	} else if sizeText == "byte" {
		stmt.size = 1
	} else {
		p.errorf("expected 'word' or 'byte', got: %s", sizeText)
		p.skipToEOL()
		return
	}
	p.function.fpLocals = append(p.function.fpLocals, &FpParam{
		id:   stmt.name,
		size: stmt.size,
	})
	p.lexer.Next()
}

func (p *Parser) parseInstruction(tok TokenType) {
	stmt := &InstructionStatement{
		operation: tok,
	}
	p.addStatement(stmt)
	p.lexer.Next()
	stmt.operands = p.parseOperands()
}

func (p *Parser) parseExpressionList() []Expr {
	var list []Expr
	for {
		expr := p.parseExpr()
		if expr == nil {
			return list
		}
		list = append(list, expr)
		if p.lexer.Token() != TokComma {
			break
		}
		p.lexer.Next()
	}
	return list
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
	if expr == nil {
		return nil
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

// PrimaryExpr :=
//
//	  '(' origin ')'
//	| Identifier
//	| Literal (int, String, Char)
func (p *Parser) parsePrimaryExpr() Expr {
	var expr Expr
	switch p.lexer.Token() {
	case TokLeftParen:
		p.lexer.Next()
		expr = p.parseExpr()
		p.expect(TokRightParen)
		p.lexer.Next()
	case TokIdent:
		expr = ExprIdent{ident: p.lexer.TokenText(), activeLabel: p.global}
	case TokString:
		s, err := strconv.Unquote(p.lexer.TokenText())
		if err != nil {
			p.errorf("invalid string literal")
		}
		expr = BytesLiteral{value: []byte(s), text: p.lexer.TokenText()}
	case TokInt:
		// strconv.ParseInt, if passed bitSize=0, will use Go's syntax for literals
		// such as 0b, 0x, underscores, etc.
		// ParseInt assumes signed types so have to pass bitSize=32 otherwise
		// would error on uint16 > 32767.
		val, err := strconv.ParseInt(p.lexer.TokenText(), 0, 32)
		if err != nil {
			p.errorf("invalid integer literal '%s'", p.lexer.TokenText())
		}
		expr = IntLiteral{value: int(val), text: p.lexer.TokenText()}
	case TokChar:
		expr = CharLiteral{value: int(p.lexer.TokenText()[1]), text: p.lexer.TokenText()}
	default:
		// its not part of an expression, just leave it be
		return expr
		// some kind of error
		//p.errorf("expected (origin), identifier, or literal (got %s)", p.lexer.Token())
		//origin = IntLiteral{value: 0}
		// todo skip to eol?
	}
	p.lexer.Next()
	return expr
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
