package truth

import (
	"errors"
	"fmt"
	"strings"
)

type Proposition interface {
	// eval returns the value of the Proposition in the given state.
	eval(state) bool

	// free returns the free variables contained in the Proposition.
	free() []Variable

	// replace returns the result of changing free occurrences of a to b.
	replace(a, b Variable) Proposition

	// needsBrackets indicates whether the Proposition should be bracketed
	// when the given operator is applied to it.
	needsBrackets(operator) bool

	equals(Proposition) bool

	fmt.Stringer
}

var errAlreadyPrenex = errors.New("already prenex")

func innexReduce(p Proposition) (Proposition, error) {
	if impl, ok := p.(implication); ok {
		return impl.innexReduce()
	} else if λ, ok := p.(lambda); ok {
		if red, err := innexReduce(λ.scope); err == nil {
			return buildLambda(λ.q, λ.v, red), nil
		}
	}
	return nil, errAlreadyPrenex
}

func prenex(p Proposition) Proposition {
	switch red, err := innexReduce(p); {
	case err == nil:
		return prenex(red)
	case err == errAlreadyPrenex:
		return p
	default:
		panic(fmt.Sprintf("unknown prenex error: %s", err))
	}
}

type Constant bool

func (b Constant) free() []Variable {
	return []Variable{}
}

func (b Constant) eval(_ state) bool {
	return bool(b)
}

func (b Constant) replace(_, __ Variable) Proposition {
	return b
}

func (b Constant) needsBrackets(_ operator) bool {
	return false
}

func (b Constant) equals(p Proposition) bool {
	if b1, ok := p.(Constant); ok {
		return b1 == b
	}
	return false
}

func (b Constant) String() string {
	return fmt.Sprintf("%t", b)
}

type Variable string

func (v Variable) free() []Variable {
	return []Variable{v}
}

func (v Variable) eval(m state) bool {
	return m[v]
}

func (v Variable) replace(a, b Variable) Proposition {
	if a == v {
		return b
	}
	return v
}

func (v Variable) needsBrackets(op operator) bool {
	return false
}

func (v Variable) equals(p Proposition) bool {
	if v1, ok := p.(Variable); ok {
		return v1 == v
	}
	return false
}

func (v Variable) String() string {
	return string(v)
}

type state map[Variable]bool

func (m state) String() string {
	parts := []string{}
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("`%s' := %t", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

type conflict struct {
	A, B state
	aval bool
}

func (c *conflict) Error() string {
	return fmt.Sprintf("state %s yields %t; %s yields %t",
		c.A, c.aval, c.B, !c.aval)
}

func states(vars []Variable) []state {
	if len(vars) == 0 {
		return []state{state{}}
	}
	var asms []state
	for _, subasm := range states(vars[1:]) {
		f, t := state{vars[0]: false},
			state{vars[0]: true}
		for k, v := range subasm {
			f[k], t[k] = v, v
		}
		asms = append(asms, f, t)
	}
	return asms
}

func Decide(p Proposition) (bool, error) {
	asms := states(p.free())
	v0 := p.eval(asms[0])
	for _, asm := range asms[1:] {
		if v0 != p.eval(asm) {
			return false, &conflict{A: asms[0], B: asm, aval: v0}
		}
	}
	return v0, nil
}
