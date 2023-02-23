package truth

import (
	"fmt"
	"reflect"
	"strings"
)

type Proposition interface {
	Variables() []Variable
	String() string
}

type Constant bool

func (b Constant) Variables() []Variable {
	return []Variable{}
}

func (b Constant) String() string {
	return fmt.Sprintf("%t", b)
}

type Variable string

func (v Variable) Variables() []Variable {
	return []Variable{v}
}

func (v Variable) String() string {
	return string(v)
}

type implication struct {
	antecedent, consequent Proposition
}

func (impl implication) Variables() []Variable {
	vars := []Variable{}
	m := map[Variable]bool{}
	for _, v := range append(
		impl.antecedent.Variables(),
		impl.consequent.Variables()...,
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

type assignment map[Variable]bool

func (m assignment) String() string {
	parts := []string{}
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("`%s` := %t", k, v))
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

func eval(p Proposition, m assignment) bool {
	if b, ok := p.(Constant); ok {
		return bool(b)
	} else if v, ok := p.(Variable); ok {
		return m[v]
	}
	impl, ok := p.(implication)
	if !ok {
		panic(fmt.Sprintf("panic: of type %s", reflect.TypeOf(p)))
	}
	return !eval(impl.antecedent, m) || eval(impl.consequent, m)
}

func assignments(vars []Variable) []assignment {
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

func Decide(p Proposition) (bool, error) {
	asms := assignments(p.Variables())
	v0 := eval(p, asms[0])
	for _, asm := range asms[1:] {
		if v0 != eval(p, asm) {
			return false, &conflict{A: asms[0], B: asm, aval: v0}
		}
	}
	return v0, nil
}
