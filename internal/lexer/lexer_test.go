package lexer

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

const file = "xyz.i2"

func TestLexer(t *testing.T) {
	buf, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lex := &Lexer{bytes.Runes(buf), 0}
	for {
		tk, err := lex.ReadToken()
		switch {
		case tk == eof:
			return
		case err != nil:
			t.Fatalf("lex error: %s", err)
		}
	}
}

type tokenReader func() *Token

func newTokenSlice(tvs []tagval) tokenReader {
	tokens := make([]Token, len(tvs))
	for i := 0; i < len(tvs); i++ {
		tokens[i] = Token{i, tvs[i].Tag, tvs[i].Value}
	}
	pos := 0
	return func() *Token {
		if pos >= len(tokens) {
			panic("reading beyond expected tokens")
		}
		next := tokens[pos]
		pos++
		return &next
	}
}

func validateTokenization(lex *Lexer, exptokens tokenReader) error {
	equalWithoutPos := func(tk, tk2 *Token) bool {
		return tk.Tag == tk2.Tag && tk.Value == tk2.Value
	}
	for {
		tk, err := lex.ReadToken()
		switch {
		case tk == eof:
			return nil
		case err != nil:
			return fmt.Errorf("lex error: %s", err)
		}
		if exptk := exptokens(); !equalWithoutPos(tk, exptk) {
			return fmt.Errorf("expected %v got %v", exptk, tk)
		}
	}
}

func TestSimple(t *testing.T) {
	cases := map[string]tokenReader{
		"  x + y = z ⊣  assoc  (a, b, 1)": newTokenSlice([]tagval{
			tagval{TagId, "x"},
			tagval{TagPunct, "+"},
			tagval{TagId, "y"},
			tagval{TagPunct, "="},
			tagval{TagId, "z"},
			tagval{TagPunct, "⊣"},
			tagval{TagId, "assoc"},
			tagval{TagPunct, "("},
			tagval{TagId, "a"},
			tagval{TagPunct, ","},
			tagval{TagId, "b"},
			tagval{TagPunct, ","},
			tagval{TagNum, "1"},
			tagval{TagPunct, ")"},
		}),
	}
	for input, exptokens := range cases {
		lex := &Lexer{[]rune(input), 0}
		if err := validateTokenization(lex, exptokens); err != nil {
			t.Fatal(err)
		}
	}
}
