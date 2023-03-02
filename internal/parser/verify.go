package parser

import (
	"fmt"
	"strings"

	"git.sr.ht/~lbnz/i2/internal/symbol"
)

var result, errors strings.Builder

func Verify(input string) (string, error) {
	sigma = symbol.Table{"1": symbol.Any}
	result.Reset()
	errors.Reset()
	if ret := yyParse(&lexer{[]rune(string(input)), 0}); ret != 0 {
		return result.String(), fmt.Errorf("parser errors:\n%s", errors.String())
	}
	return result.String(), fmt.Errorf("%s", errors.String())
}
