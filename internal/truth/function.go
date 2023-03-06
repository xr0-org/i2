package truth

import (
	"fmt"
	"strings"
)

type function struct {
	name string
	vars []Variable
}

func (fn function) eval(_ state) bool {
	panic("cannot eval function")
}

func (fn function) free() []Variable {
	return []Variable(fn.vars)
}

func (fn function) replace(a, b Variable) Proposition {
	vars := make([]Variable, len(fn.vars))
	for i, v := range fn.vars {
		nv, ok := v.replace(a, b).(Variable)
		if !ok {
			panic("replace returned non-variable")
		}
		vars[i] = nv
	}
	return function{fn.name, vars}
}

func (fn function) needsBrackets(_ operator) bool {
	return false
}

func (fn function) equals(p Proposition) bool {
	fn2, ok := p.(function)
	if !ok {
		return false
	}
	if len(fn.vars) != len(fn2.vars) {
		return false
	}
	for i := range fn.vars {
		if !fn.vars[i].equals(fn2.vars[i]) {
			return false
		}
	}
	return fn.name == fn2.name
}

func (fn function) String() string {
	sarr := make([]string, len(fn.vars))
	for i := range fn.vars {
		sarr[i] = string(fn.vars[i])
	}
	return fmt.Sprintf("%s(%s)", fn.name, strings.Join(sarr, ", "))
}
