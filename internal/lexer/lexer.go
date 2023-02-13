package lexer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

var (
	errEOF   error = errors.New("lexer at EOF")
	tagError       = errors.New("cannot tag")
)

type tagval struct {
	Tag
	Value string
}

func (l *Lexer) ReadToken() (*Token, error) {
	n, err := skipWhitespace(l.input[l.pos:])
	if err != nil {
		if err == errEOF {
			return eof, nil
		}
		return nil, err
	}
	l.pos += n
	taggers := []func(input []rune) (*tagval, error){
		parseNum, parsePunct, parseId,
	}
	for _, tagger := range taggers {
		tv, err := tagger(l.input[l.pos:])
		if err == nil {
			tk := &Token{l.pos, tv.Tag, tv.Value}
			l.pos += len(tv.Value)
			return tk, nil
		}
		if err != tagError {
			panic(fmt.Sprintf("unknown error: %s", err))
		}
	}
	return nil, fmt.Errorf("cannot recognize rune: %c", l.input[l.pos])
}

func skipWhitespace(input []rune) (int, error) {
	for n := 0; n < len(input); n++ {
		if !unicode.IsSpace(input[n]) {
			return n, nil
		}
	}
	return 0, errEOF
}

func parsePunct(input []rune) (*tagval, error) {
	if r := input[0]; isPunctuationSymbol[r] {
		return &tagval{TagPunct, fmt.Sprintf("%c", r)}, nil
	}
	return nil, tagError
}

func parseNum(input []rune) (*tagval, error) {
	if !unicode.IsDigit(input[0]) {
		return nil, tagError
	}
	var b strings.Builder
	for n := 0; n < len(input) && unicode.IsDigit(input[n]); n++ {
		b.WriteRune(input[n])
	}
	return &tagval{TagNum, b.String()}, nil
}

func parseId(input []rune) (*tagval, error) {
	if !unicode.IsLetter(input[0]) {
		return nil, tagError
	}
	cond := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	var b strings.Builder
	for n := 0; n < len(input) && cond(input[n]); n++ {
		b.WriteRune(input[n])
	}
	return &tagval{TagId, b.String()}, nil
}
