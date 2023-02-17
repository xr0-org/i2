package truth

import (
	"fmt"
	"strings"
)

type proposition interface {
	variables() []variable
	String() string
}

type constant bool

func (b constant) variables() []variable {
	return []variable{}
}

func (b constant) String() string {
	return fmt.Sprintf("%t", b)
}

type variable string

func (v variable) variables() []variable {
	return []variable{v}
}

func (v variable) String() string {
	return string(v)
}

type implication struct {
	antecedent, consequent proposition
}

func (impl implication) variables() []variable {
	vars := []variable{}
	m := map[variable]bool{}
	for _, v := range append(
		impl.antecedent.variables(),
		impl.consequent.variables()...,
	) {
		if m[v] {
			continue
		}
		m[v] = true
		vars = append(vars, v)
	}
	return vars
}

func (impl implication) String() string {
	return fmt.Sprintf("(%s ==> %s)", impl.antecedent, impl.consequent)
}

type assignment map[variable]bool

func (m assignment) String() string {
	parts := []string{}
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s := %t", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

type conflict struct {
	A, B assignment
	aval bool
}

func (c *conflict) Error() string {
	return fmt.Sprintf("case %s yields %t; %s yields %t",
		c.A, c.aval, c.B, !c.aval)
}

func eval(p proposition, m assignment) bool {
	if b, ok := p.(constant); ok {
		return bool(b)
	} else if v, ok := p.(variable); ok {
		return m[v]
	}
	impl, _ := p.(implication)
	return !eval(impl.antecedent, m) || eval(impl.consequent, m)
}

func assignments(vars []variable) []assignment {
	if len(vars) == 0 {
		return []assignment{assignment{}}
	}
	var asms []assignment
	for _, subasm := range assignments(vars[1:]) {
		f, t := assignment{vars[0]: false},
			assignment{vars[0]: true}
		for k, v := range subasm {
			f[k], t[k] = v, v
		}
		asms = append(asms, f, t)
	}
	return asms
}

func decide(p proposition) (bool, error) {
	asms := assignments(p.variables())
	v0 := eval(p, asms[0])
	for _, asm := range asms[1:] {
		if v0 != eval(p, asm) {
			return false, &conflict{A: asms[0], B: asm, aval: v0}
		}
	}
	return v0, nil
}
