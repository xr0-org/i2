package symbol

import "errors"

type Symbol interface {
	Table() (Table, error)
}

type Table map[string]Symbol

type FunctionSymbol struct {
	IsAxiom bool
	T       Table
	Return  Type
}

func (f FunctionSymbol) Table() (Table, error) {
	return f.T, nil
}

type TemplateSymbol struct {
	IsAxiom bool
	T       Table
}

func (t TemplateSymbol) Table() (Table, error) {
	return t.T, nil
}

type FunctionVariable struct {
	Params []Type
	Return Type
}

var ErrNoTable = errors.New("symbol has no table ")

func (f FunctionVariable) Table() (Table, error) {
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
