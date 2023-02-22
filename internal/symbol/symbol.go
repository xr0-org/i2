package symbol

import (
	"errors"
	"fmt"
)

var (
	ErrNoTable = errors.New("symbol has no table")

	errIsInvocationLengthMismatch    = "length of parameters does not match parameters in function abstract"
	errIsInvocationParameterMismatch = "parameter %s is of type %s but function definition requires type %s in position %d"
)

type Symbol interface {
	Table() (Table, error)
}

type Table map[string]Symbol

type Function struct {
	IsAxiom bool
	Type    FunctionType
}

func (f Function) Table() (Table, error) {
	T := Table{}
	for _, p := range f.Type.Params {
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

func (f Function) IsInvocation(params []Parameter) error {
	if len(params) != len(f.Type.Params) {
		return fmt.Errorf(errIsInvocationLengthMismatch)
	}
	for i := range params {
		p, fp := params[i], f.Type.Params[i]
		if !(fp.Type == Any || fp.Type == p.Type) {
			return fmt.Errorf(
				errIsInvocationLengthMismatch,
				p.Name, p.Type, fp.Type, i,
			)
		}
	}
	return nil
}

func (f Function) String() string {
	return fmt.Sprintf("%sfunc %s", optionalat(f.IsAxiom), f.Type)
}

type Template struct {
	IsAxiom bool
	Params  []Parameter
}

func (t Template) Table() (Table, error) {
	T := Table{}
	for _, p := range t.Params {
		T[p.Name] = p.Type
	}
	return T, nil
}

func (t Template) String() string {
	return fmt.Sprintf("%stmpl(%s)", optionalat(t.IsAxiom), t.Params)
}

type FunctionType struct {
	Params []Parameter
	Return Type
}

func (f FunctionType) Table() (Table, error) {
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
