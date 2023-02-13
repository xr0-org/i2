package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	tkEOF   int = 0
	tkError     = -1
)

type lexer struct {
	input []rune
	pos   int
}

func (l *lexer) Error(err string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
}

func div(a, b int) int {
	if b == 0 {
		panic("division by zero")
	}
	return a / b
}

func (l *lexer) Lex(lval *yySymType) int {
	if l.pos >= len(l.input) {
		return tkEOF
	}
	switch c := l.input[l.pos]; c {
	case '+', '*', '-', '/', '(', ')', '\n':
		l.pos++
		return int(c)
	}
	i, delta := parseNum(l.input[l.pos:])
	if i == tkError {
		return tkError
	}
	l.pos += delta
	lval.val = i
	return tkNumber
}

func parseNum(input []rune) (int, int) {
	if !unicode.IsDigit(input[0]) {
		return tkError, 0
	}
	var b strings.Builder
	for n := 0; n < len(input) && unicode.IsDigit(input[n]); n++ {
		b.WriteRune(input[n])
	}
	num := b.String()
	i, err := strconv.Atoi(num)
	if err != nil {
		return tkError, 0
	}
	return i, len(num)
}
