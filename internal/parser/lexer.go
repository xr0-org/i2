package parser

import (
	"fmt"
	"os"
	"strconv"
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
	fmt.Fprintf(os.Stderr, "error: %s, near: '%.*s'\n", err, 10, string(l.input[l.pos-10:]))
}

func (l *lexer) Lex(lval *yySymType) int {
	if l.pos += skipWhitespace(l.input[l.pos:]); l.pos >= len(l.input) {
		return tkEof
	}
	lexers := []func([]rune, *yySymType) (*token, error){
		lexPunct, lexString, lexNum,
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

var dblpunct = map[rune]int{
	'!': tkNe,
	'=': tkEq,
	'>': tkGt,
	'<': tkLt,
	'&': tkAnd,
}

func lexPunct(input []rune, lval *yySymType) (*token, error) {
	switch c := input[0]; c {
	case ';', '(', ')', '{', '}', '@', ',', ':':
		return &token{int(c), 1}, nil
	case '!', '=', '>', '<', '&':
		if len(input) < 2 {
			return &token{int(c), 1}, nil
		}
		tk, ok := dblpunct[c]
		if !ok {
			return nil, fmt.Errorf("unknown symbol '%s'", string(input[:2]))
		}
		if len(input) > 2 {
			d := input[2]
			switch tk {
			case tkEq:
				switch d {
				case '=':
					return &token{tkEqv, 3}, nil
				case '>':
					return &token{tkImpl, 3}, nil
				}
			case tkLt:
				if d == '=' {
					return &token{tkRevImpl, 3}, nil
				}
			}
		}
		return &token{tk, 2}, nil
	default:
		return nil, fmt.Errorf("unknown char '%c'", c)
	}
}

var keyword = map[string]int{
	"type": tkType,
	"func": tkFunc,

	"auto": tkAuto,

	"for": tkFor,
}

func stringtype(s string) int {
	if tk, ok := keyword[s]; ok {
		return tk
	}
	return tkIdentifier
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

func lexNum(input []rune, lval *yySymType) (*token, error) {
	if c := input[0]; !(unicode.IsDigit(c)) {
		return nil, fmt.Errorf("invalid number first char: '%c'", c)
	}
	var b strings.Builder
	for n := 0; n < len(input) && unicode.IsDigit(input[n]); n++ {
		b.WriteRune(input[n])
	}
	val, err := strconv.Atoi(b.String())
	if err != nil {
		return nil, fmt.Errorf("could not parse int: '%s'", b.String())
	}
	lval.n = val
	return &token{tkConstant, len(fmt.Sprint(lval.n))}, nil
}
