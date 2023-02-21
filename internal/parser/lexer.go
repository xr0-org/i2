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

type lineinfo struct {
	lines   [][]rune
	linepos int // the position on the line
	n       int // the line number
}

func getlineinfo(input []rune, pos int) *lineinfo {
	info := &lineinfo{lines: [][]rune{}}
	last := 0
	for i := 0; i < len(input); i++ {
		if input[i] == '\n' {
			if last <= pos && pos <= i {
				info.linepos = pos - last
				info.n = len(info.lines)
			}
			info.lines = append(info.lines, input[last:i])
			last = i + 1
		}
	}
	return info
}

type colour string

const (
	colourRed colour = "\x1B[31m"
	colourGrn        = "\x1B[32m"
	colourYel        = "\x1B[33m"
	colourBlu        = "\x1B[34m"
	colourMag        = "\x1B[35m"
	colourCyn        = "\x1B[36m"
	colourWht        = "\x1B[37m"
	colourOff        = "\x1B[0m"
)

func (l *lexer) Error(err string) {
	info := getlineinfo(l.input, l.pos)
	fmt.Fprint(os.Stderr, colourRed)
	fmt.Fprintf(os.Stderr, ">>> %s\n    ", string(info.lines[info.n]))
	fmt.Fprint(os.Stderr, colourOff)
	for i := 0; i < info.linepos; i++ {
		fmt.Fprintf(os.Stderr, " ")
	}
	fmt.Fprintf(os.Stderr, "^\n")
	fmt.Fprintf(os.Stderr, "error: %s at position %d on line %d\n",
		err, info.linepos, info.n+1)
	os.Exit(1)
}

func (l *lexer) Lex(lval *yySymType) int {
	if l.pos += skipNPCs(l.input[l.pos:], l); l.pos >= len(l.input) {
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
	l.Error("unrecoginized sequence")
	return tkError
}

func skipNPCs(input []rune, l *lexer) int {
	n := 0
	for n < len(input) && unicode.IsSpace(input[n]) {
		n++
	}
	if n+1 < len(input) && string(input[n:n+2]) == "/*" {
		n += skipComments(input[n:], l)
		return n + skipNPCs(input[n:], l)
	}
	return n
}

func skipComments(input []rune, l *lexer) int {
	n := 0
	for n+1 < len(input) {
		if string(input[n:n+2]) == "*/" {
			return n + 2
		}
		n++
	}
	l.Error("file ends in comment")
	return -1
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
	case ';', '(', ')', '[', ']', '{', '}', '@', ',', ':', '~':
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
					return &token{tkFllw, 3}, nil
				}
			}
		}
		d := input[1]
		switch tk {
		case tkAnd:
			if d == '&' {
				return &token{tk, 2}, nil
			}
			break
		default:
			if d == '=' {
				return &token{tk, 2}, nil
			}
			if tk == tkNe {
				return &token{'!', 1}, nil
			}
			break
		}
		return nil, fmt.Errorf("unknown symbol '%s'", string(input[:2]))
	default:
		return nil, fmt.Errorf("unknown char '%c'", c)
	}
}

var keyword = map[string]int{
	"tmpl": tkTmpl,
	"func": tkFunc,

	"this": tkThis,
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
	lval.s = b.String()
	return &token{tkConstant, len(lval.s)}, nil
}
