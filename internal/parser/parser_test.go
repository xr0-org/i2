package parser

import "testing"

func TestExpression(t *testing.T) {
	if ret := yyParse(&lexer{[]rune("12*(6/2)/0\n"), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
