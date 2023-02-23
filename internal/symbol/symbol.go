package symbol

import (
	"errors"
	"fmt"

	"git.sr.ht/~lbnz/i2/internal/truth"
)

var (
	ErrNoTable = errors.New("symbol has no table")

	errIsInvocationLengthMismatch    = "length of parameters does not match parameters in function abstract"
	errIsInvocationParameterMismatch = "parameter %s is of type %s but function definition requires type %s in position %d"

	errFunctionOrTemplateNotDefined = "function/template: %s not defined"
	errVariableNotDefined           = "variable %s not defined"
	errNonSimpleExpr                = "%s is not a primary expression"
	errNonInvocableInvoked          = "%s cannot be invoked"
	errOpOnNonBoolExpr              = "op %s cannot be applied to expr %s of type %s"
	errInvalidBinaryOp              = "op %s is an invalid binary op"
)

type Symbol interface {
	Table() (Table, error)
}

type Table map[string]Symbol

func (T Table) Nest(Parent Table) Table {
	newT := Table{}
	for k, v := range Parent {
		newT[k] = v
	}
	for k, v := range T {
		newT[k] = v
	}
	return newT
}

type Function struct {
	IsAxiom bool
	Sig     FunctionSignature
}

func (f Function) isPredicate() bool {
	return f.Sig.Return == Bool
}

func (f Function) Table() (Table, error) {
	T := Table{}
	for _, p := range f.Sig.Params {
		T[p.Name] = p.Type
	}
	return T, nil
}

func optionalat(at bool) string {
	if at {
		return "@"
	}
	return ""
}

type Parameter struct {
	Name string
	Type
}

func (p Parameter) isBool() bool {
	return p.Type == Bool
}

func (f Function) IsInvocation(params []Parameter) error {
	if len(params) != len(f.Sig.Params) {
		return fmt.Errorf(errIsInvocationLengthMismatch)
	}
	for i := range params {
		p, fp := params[i], f.Sig.Params[i]
		if !(fp.Type == Any || fp.Type == p.Type) {
			return fmt.Errorf(
				errIsInvocationParameterMismatch,
				p.Name, p.Type, fp.Type, i,
			)
		}
	}
	return nil
}

func (f Function) String() string {
	return fmt.Sprintf("%sfunc %s", optionalat(f.IsAxiom), f.Sig)
}

type Template struct {
	IsAxiom bool
	Params  []Parameter
	E       Expr
	Proofs  []RelationChain
}

func (t Template) instantiate(args []Expr, tbl Table) (truth.Proposition, error) {
	m := map[Expr]Expr{}
	// by assumption the type checking for args & the template has already
	// been done, so we can map directly
	for i, param := range t.Params {
		m[SimpleExpr(param.Name)] = args[i]
	}
	aExpr, err := t.E.replace(m).Analyse(tbl)
	return aExpr.P, err
}

func (t Template) Table() (Table, error) {
	T := Table{}
	for _, p := range t.Params {
		T[p.Name] = p.Type
	}
	return T, nil
}

func (t Template) IsInvocation(params []Parameter) error {
	if len(params) != len(t.Params) {
		return fmt.Errorf(errIsInvocationLengthMismatch)
	}
	for i := range params {
		p, tp := params[i], t.Params[i]
		if !(tp.Type == Any || tp.Type == p.Type) {
			return fmt.Errorf(
				errIsInvocationParameterMismatch,
				p.Name, p.Type, tp.Type, i,
			)
		}
	}
	return nil
}

func (t Template) String() string {
	return fmt.Sprintf("%stmpl(%s)", optionalat(t.IsAxiom), t.Params)
}

type FunctionSignature struct {
	Params []Parameter
	Return Type
}

func (f FunctionSignature) Table() (Table, error) {
	return nil, ErrNoTable
}

type Type string

const (
	Any  Type = "any"
	Bool      = "bool"
)

func (p Type) Table() (Table, error) {
	return nil, ErrNoTable
}
