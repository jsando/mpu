package asm

import (
	"io"
	"strings"
	"text/scanner"
)

type TokenType int

const (
	TokNone TokenType = iota
	TokEOF
	TokIdent
	TokInt
	TokChar
	TokString
	TokDot
	TokColon
	TokHash
	TokStar
	TokComma
	TokPlus
	TokMinus
	TokPipe
	TokCaret
	TokSlash
	TokPercent
	TokLeftShift
	TokRightShift
	TokLeftParen
	TokRightParen
	TokLeftBracket
	TokRightBracket
	TokEquals
	TokOrg
	TokDw
	TokDb
	TokDs
	TokLocal
	TokAdd
	TokSub
	TokMul
	TokDiv
	TokCmp
	TokAnd
	TokOr
	TokXor
	TokCpy
	TokPsh
	TokPop
	TokJmp
	TokJeq
	TokJne
	TokJge
	TokJlt
	TokInc
	TokDec
	TokJsr
	TokRet
	TokClc
	TokSec
	TokClb
	TokSeb
	TokJcc
	TokJcs
	TokSav
	TokRst
	TokHlt
	TokFunction
	TokEOL
)

// tokenImage has the image of the token in the same order as the Tok* constants.
var tokenImage = []string{
	"<none>",
	"<eof>", "<ident>", "<int>", "<char>", "<string>",
	".", ":", "#", "*", ",",
	"+", "-", "|", "^", "/",
	"%", "<<", ">>", "(", ")", "[", "]",
	"=", "org", "dw", "db", "ds", "local",
	"add", "sub", "mul", "div", "cmp",
	"and", "or", "xor", "cpy", "psh",
	"pop", "jmp", "jeq", "jne", "jge",
	"jlt", "inc", "dec", "jsr", "ret",
	"clc", "sec", "clb", "seb", "jcc",
	"jcs", "sav", "rst", "hlt", "function()",
	"<eol>",
}

func (t TokenType) String() string {
	return tokenImage[t]
}

type Lexer struct {
	s    *scanner.Scanner
	line int
	tok  TokenType
}

func NewLexer(name string, r io.Reader) *Lexer {
	var s scanner.Scanner
	s.Init(r)
	// Newline is like a semicolon; other space characters are fine.
	s.Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '
	// Don't skip comments: we need to count newlines.
	s.Mode = scanner.ScanChars |
		scanner.ScanIdents |
		scanner.ScanInts |
		scanner.ScanStrings |
		scanner.ScanComments
	s.Position.Filename = name
	return &Lexer{
		s:    &s,
		line: 1,
	}
}

func (l *Lexer) FileName() string {
	return l.s.Filename
}

func (l *Lexer) Line() int {
	return l.s.Line
}

func (l *Lexer) Column() int {
	return l.s.Column
}

func (l *Lexer) Next() TokenType {
	s := l.s
	var scanToken rune
	for {
		scanToken = s.Scan()
		if scanToken != scanner.Comment {
			break
		}
		text := s.TokenText()
		l.line += strings.Count(text, "\n")
	}
	l.tok = TokNone
	switch scanToken {
	case '\n':
		l.line++
		l.tok = TokEOL
	case ':':
		l.tok = TokColon
	case '.':
		l.tok = TokDot
	case '=':
		l.tok = TokEquals
	case '#':
		l.tok = TokHash
	case '*':
		l.tok = TokStar
	case ',':
		l.tok = TokComma
	case '+':
		l.tok = TokPlus
	case '-':
		l.tok = TokMinus
	case '|':
		l.tok = TokPipe
	case '^':
		l.tok = TokCaret
	case '/':
		l.tok = TokSlash
	case '%':
		l.tok = TokPercent
	case '(':
		l.tok = TokLeftParen
	case ')':
		l.tok = TokRightParen
	case '[':
		l.tok = TokLeftBracket
	case ']':
		l.tok = TokRightBracket
	case '<':
		if s.Peek() == '<' {
			l.tok = TokLeftShift
		}
	case '>':
		if s.Peek() == '>' {
			l.tok = TokRightShift
		}
	case scanner.EOF:
		l.tok = TokEOF
	case scanner.String:
		l.tok = TokString
	case scanner.Char:
		l.tok = TokChar
	case scanner.Int:
		l.tok = TokInt
	case scanner.Ident:
		l.tok = TokIdent
	}
	//fmt.Printf("> type: %s, text: '%s'\n", l.tok, l.s.TokenText())
	return l.tok
}

func (l *Lexer) syncNextStmt() TokenType {
	for l.tok == TokEOL {
		l.Next()
	}
	return l.tok
}

// skipToEOL skips all tokens up to the next EOL, useful in error recovery.
func (l *Lexer) skipToEOL() {
	for l.tok != TokEOL && l.tok != TokEOF {
		l.Next()
	}
}

func (l *Lexer) TokenText() string {
	return l.s.TokenText()
}

func toKeyword(ident string) TokenType {
	for i, image := range tokenImage {
		if ident == image {
			return TokenType(i)
		}
	}
	return TokIdent
}
