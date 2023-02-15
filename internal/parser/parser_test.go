package parser

import "testing"

func TestParsing(t *testing.T) {
	input := `
@func nat(x) bool;
@func succ(x nat) nat;
@succ_notone{x}: x nat ==> succ(x) != 1;
@injectivity{x, y}: x, y nat && succ(x) == succ(y) ==> x == y;
`
	if ret := yyParse(&lexer{[]rune(input), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
