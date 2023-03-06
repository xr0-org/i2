package truth

import (
	"fmt"
)

type quantifier string

const (
	universal   quantifier = "∀"
	existential            = "∃"
)

func (q quantifier) flip() quantifier {
	return map[quantifier]quantifier{
		universal:   existential,
		existential: universal,
	}[q]
}

// operator returns the operator (for display purposes) corresponding to the
// quantifier.
func (q quantifier) operator() operator {
	switch q {
	case universal:
		return opUniversalQuantification
	case existential:
		return opExistentialQuantification
	default:
		panic(fmt.Sprintf("invalid quantifier %s", q))
	}
}

type lambda struct {
	q     quantifier
	v     Variable
	scope Proposition
}

func (λ lambda) free() []Variable {
	var vars []Variable
	for _, v := range λ.scope.free() {
		if v != λ.v {
			vars = append(vars, v)
		}
	}
	return vars
}

func (λ lambda) eval(_ state) bool {
	panic("lambdas cannot eval")
}

func occursFreely(a Variable, D Proposition) bool {
	for _, v := range D.free() {
		if v == a {
			return true
		}
	}
	return false
}

func (λ lambda) reductionVariable(D Proposition) Variable {
	if !occursFreely(λ.v, D) {
		return λ.v
	}
	i, v := 0, λ.v
	for occursFreely(v, λ.scope) || occursFreely(v, D) {
		// NB: the assumption here is that bound occurrences in nested
		// lambdas are irrelevant.
		v = Variable(fmt.Sprintf("%s%d", λ.v, i))
		i++
	}
	return v
}

func (λ lambda) replace(a, b Variable) Proposition {
	// FIXME: this is not working
	return λ
}

func (λ lambda) needsBrackets(_ operator) bool {
	return false
}

func (λ lambda) equals(p Proposition) bool {
	λ2, ok := p.(lambda)
	if !ok {
		return false
	}
	return λ.q == λ2.q && λ.v == λ2.v && λ.scope.equals(λ2.scope)
}

func (λ lambda) String() string {
	switch λ.q {
	case universal:
		return opUniversalQuantification.format(λ.v, λ.scope)
	case existential:
		return opExistentialQuantification.format(λ.v, λ.scope)
	default:
		panic(fmt.Sprintf("unknown quantifier %s", λ.q))
	}
}
