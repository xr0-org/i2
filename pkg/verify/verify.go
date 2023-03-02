package verify

import "git.sr.ht/~lbnz/i2/internal/parser"

func Verify(input string) (string, error) {
	result, err := parser.Verify(input)
	if err != nil {
		return "", err
	}
	return result, nil
}
