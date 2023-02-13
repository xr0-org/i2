package lexer

import "fmt"

type Tag string

const (
	TagId    Tag = "identifier"
	TagNum       = "number"
	TagPunct     = "punctuation"
)

const EOF string = "$"

var eof *Token = &Token{Tag: TagPunct, Value: EOF}

type Token struct {
	position int
	Tag
	Value string
}

func (tk Token) String() string {
	return fmt.Sprintf("{%q}", tk.Value)
}

var punctuationSymbols = []rune{
	'(', ')', '[', ']', '{', '}', // brackets
	'.',      // subaccessor
	',',      // parameter concatenator
	';',      // statement concatenator
	':',      // labeller
	'+', '-', // operation
	'=', // equals sign
	'⊣', // turnstile
}

var isPunctuationSymbol = map[rune]bool{}

func init() {
	for _, r := range punctuationSymbols {
		isPunctuationSymbol[r] = true
	}
}
