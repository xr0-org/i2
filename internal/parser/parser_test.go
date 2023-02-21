package parser

import (
	"os"
	"testing"
)

func TestParsing(t *testing.T) {
	input, err := os.ReadFile("../../examples/landau/addition.i2")
	if err != nil {
		t.Fatal(err)
	}
	if ret := yyParse(&lexer{[]rune(string(input)), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
