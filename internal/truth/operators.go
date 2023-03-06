package truth

import (
	"fmt"
)

type precedence int

const (
	precConnective precedence = iota
	precDisjunction
	precConjunction
	precNegation
	precSingular
)

type operator int

const (
	// unary
	opIdentity operator = iota
	opNegation

	// binary
	opUniversalQuantification
	opExistentialQuantification
	opConjunction
	opDisjunction
	opImplication
	opEquivalence
)

func (op operator) unary() bool {
	return op < opUniversalQuantification
}

func (op operator) precedence() precedence {
	switch op {
	case opIdentity:
		return precSingular
	case opNegation, opUniversalQuantification, opExistentialQuantification:
		return precNegation
	case opConjunction:
		return precConjunction
	case opDisjunction:
		return precDisjunction
	case opImplication, opEquivalence:
		return precConnective
	default:
		panic(fmt.Sprintf("unknown operator %d", op))
	}
}

func (op operator) String() string {
	switch op {
	case opIdentity:
		return "id"
	case opNegation:
		return "!"
	case opUniversalQuantification:
		return "∀"
	case opExistentialQuantification:
		return "∃"
	case opConjunction:
		return "&&"
	case opDisjunction:
		return "||"
	case opImplication:
		return "==>"
	case opEquivalence:
		return "==="
	default:
		panic(fmt.Sprintf("unknown operator %d", op))
	}
}

func (op operator) fmtstring() string {
	switch op {
	case opIdentity:
		return "%s"
	case opNegation:
		return "!%s"
	case opUniversalQuantification:
		return "(∀%s)%s"
	case opExistentialQuantification:
		return "(∃%s)%s"
	case opConjunction:
		return "%s && %s"
	case opDisjunction:
		return "%s || %s"
	case opImplication:
		return "%s ==> %s"
	case opEquivalence:
		return "%s === %s"
	default:
		panic(fmt.Sprintf("unknown operator %d", op))
	}
}

func bracketed(p Proposition, op operator) string {
	if p.needsBrackets(op) {
		return fmt.Sprintf("(%s)", p)
	}
	return fmt.Sprintf("%s", p)
}

func (op operator) propsToAny(props []Proposition) []any {
	arr := make([]any, len(props))
	for i := range props {
		arr[i] = bracketed(props[i], op)
	}
	return arr
}

func (op operator) rightArgs(nargs int) bool {
	return (op.unary() && nargs == 1) || nargs == 2
}

func (op operator) format(arg ...Proposition) string {
	if l := len(arg); !op.rightArgs(l) {
		panic(fmt.Sprintf("%s cannot take %d arguments", op, l))
	}
	return fmt.Sprintf(op.fmtstring(), op.propsToAny(arg)...)
}

type operation struct {
	operator operator
	A, B     Proposition
}

func resolveNot(p Proposition) (*operation, bool) {
	impl, ok := p.(implication)
	if !ok {
		return nil, false
	}
	if !impl.consequent.equals(Constant(false)) {
		return nil, false
	}
	return &operation{operator: opNegation, A: impl.antecedent}, true
}

func resolveDblNot(p Proposition) (*operation, bool) {
	not, ok := resolveNot(p)
	if !ok {
		return nil, false
	}
	dblnot, ok := resolveNot(not.A)
	if !ok {
		return nil, false
	}
	return &operation{operator: opIdentity, A: dblnot.A}, true
}

func resolveCanonicalForm(p Proposition) (*operation, bool) {
	if dblnot, ok := resolveDblNot(p); ok {
		return dblnot, true
	}
	return nil, false
}

func resolveAnd(p Proposition) (*operation, bool) {
	// !(B ==> !A)
	not, ok := resolveNot(p)
	if !ok {
		return nil, false
	}
	// B ==> !A
	impl, ok := not.A.(implication)
	if !ok {
		return nil, false
	}
	// !A
	notA, ok := resolveNot(impl.consequent)
	if !ok {
		return nil, false
	}
	return &operation{opConjunction, notA.A, impl.antecedent}, true
}

func resolveOr(p Proposition) (*operation, bool) {
	// (!B ==> A)
	impl, ok := p.(implication)
	if !ok {
		return nil, false
	}
	// !B ==> A
	notB, ok := resolveNot(impl.antecedent)
	if !ok {
		return nil, false
	}
	return &operation{opDisjunction, impl.consequent, notB.A}, true
}

func resolveEqv(p Proposition) (*operation, bool) {
	// (A ==> B) && (B ==> A)
	and, ok := resolveAnd(p)
	if !ok {
		return nil, false
	}
	// forward (A ==> B)
	fwd, ok := and.A.(implication)
	if !ok {
		return nil, false
	}
	// backward (B ==> A)
	bwd, ok := and.B.(implication)
	if !ok {
		return nil, false
	}
	if !fwd.antecedent.equals(bwd.consequent) ||
		!fwd.consequent.equals(bwd.antecedent) {
		return nil, false
	}
	return &operation{opEquivalence, fwd.antecedent, fwd.consequent}, true
}

func resolveImplOp(impl implication) *operation {
	resolvers := []func(Proposition) (*operation, bool){
		resolveCanonicalForm,
		resolveEqv,
		resolveOr,
		resolveAnd,
		resolveNot,
	}
	for _, resolve := range resolvers {
		if r, ok := resolve(impl); ok {
			return r
		}
	}
	return &operation{opImplication, impl.antecedent, impl.consequent}
}

func formatImplOp(impl implication) string {
	switch r := resolveImplOp(impl); r.operator {
	case opIdentity, opNegation:
		return r.operator.format(r.A)
	default:
		return r.operator.format(r.A, r.B)
	}
}

func Impl(A, B Proposition) Proposition {
	return implication{A, B}
}

func Not(A Proposition) Proposition {
	return Impl(A, Constant(false))
}

func Fllw(A, B Proposition) Proposition {
	return Impl(B, A)
}

func Or(A, B Proposition) Proposition {
	return Impl(Not(B), A)
}

func And(A, B Proposition) Proposition {
	return Not(Impl(B, Not(A)))
}

func Eqv(A, B Proposition) Proposition {
	return And(Impl(A, B), Impl(B, A))
}

func buildLambda(q quantifier, v Variable, scope Proposition) Proposition {
	return lambda{q, v, scope}
}

func Universal(v Variable, p Proposition) Proposition {
	return buildLambda(universal, v, p)
}

func Existential(v Variable, p Proposition) Proposition {
	return buildLambda(existential, v, p)
}

func Func(name string, vars ...Variable) Proposition {
	return function{name, vars}
}
