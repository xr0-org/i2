package parser

import "testing"

func TestParsing(t *testing.T) {
	input := `
@func eq(x, y) bool;
@func nat(x) bool;
@tmpl one() { 1 nat };
@func succ(x nat) nat;
@tmpl succ_notone(x) { x nat ==> !eq(succ(x), 1) };
@tmpl injectivity(x, y) { x, y nat && eq(succ(x), succ(y)) ==> eq(x, y) };
@tmpl induction(y) {
	y func(nat) bool ==> (
		y(1) && (x){x nat ==> ( y(x) ==> y(succ(x)) )}
	==> 	(x){x nat ==> y(x)}
	)
};

tmpl thm1(x, y) { x, y nat && !eq(x, y) ==> !eq(succ(x), succ(y)) } {
	!this[0]
===	x, y nat && !eq(x, y) && eq(succ(x), succ(y))
==> { injectivity(x, y) } ~ [0, 2] 
	eq(x, y) && !eq(x, y)
===	false
};
`
	if ret := yyParse(&lexer{[]rune(input), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
