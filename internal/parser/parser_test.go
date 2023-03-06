package parser

import (
	"os"
	"testing"
)

const additionFile = "../../examples/landau/addition-induction.i2"

func TestParsing(t *testing.T) {
	input, err := os.ReadFile(additionFile)
	if err != nil {
		t.Fatal(err)
	}
	if ret := yyParse(&lexer{[]rune(string(input)), 0}); ret != 0 {
		t.Fatal("returned", ret)
	}
}
