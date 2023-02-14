package parser

import "testing"

func TestType(t *testing.T) {
	input := `
type @nat(a, b, c);
type even(a);
`
	if ret := yyParse(&lexer{[]rune(input), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
