package parser

import "testing"

func TestParsing(t *testing.T) {
	input := `
type @nat(a, b, c);
type even(a);
auto @one: 1 nat;		
func @succ(x nat) nat;
`
	if ret := yyParse(&lexer{[]rune(input), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}

