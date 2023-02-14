package parser

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	tkEof   int = 0
	tkError     = -1
)

type lexer struct {
	input []rune
	pos   int
}

func (l *lexer) Error(err string) {
	fmt.Fprintf(os.Stderr, "error: %s, at rune: '%c'\n", err, l.input[l.pos])
}

func (l *lexer) Lex(lval *yySymType) int {
	if l.pos += skipWhitespace(l.input[l.pos:]); l.pos >= len(l.input) {
		return tkEof
	}
	lexers := []func([]rune, *yySymType) (*token, error){
		lexPunct, lexString,
	}
	for _, lex := range lexers {
		tk, err := lex(l.input[l.pos:], lval)
		if err != nil {
			continue
		}
		l.pos += tk.length
		return tk.token
	}
	return tkError
}

func skipWhitespace(input []rune) int {
	n := 0
	for n < len(input) && unicode.IsSpace(input[n]) {
		n++
	}
	return n
}

type token struct {
	token, length int
}

func lexPunct(input []rune, lval *yySymType) (*token, error) {
	switch c := input[0]; c {
	case ';', '(', ')', '@', ',':
		return &token{int(c), 1}, nil
	default:
		return nil, fmt.Errorf("unknown char '%c'", c)
	}
}

var keyword = map[string]int{
	"type": tkType,
}

func stringtype(s string) int {
	if tk, ok := keyword[s]; ok {
		return tk
	}
	return tkId
}

func lexString(input []rune, lval *yySymType) (*token, error) {
	if c := input[0]; !(unicode.IsLetter(c) || c == '_') {
		return nil, fmt.Errorf("invalid id/keyword first char: '%c'", c)
	}
	cond := func(c rune) bool {
		return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_'
	}
	var b strings.Builder
	for n := 0; n < len(input) && cond(input[n]); n++ {
		b.WriteRune(input[n])
	}
	lval.s = b.String()
	return &token{stringtype(lval.s), len(lval.s)}, nil
}
