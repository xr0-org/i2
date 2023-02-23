package symbol

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"git.sr.ht/~lbnz/i2/internal/truth"
)

type AnalysedExpr struct {
	P   truth.Proposition
	arg Parameter
}

// Expr is the interface describing the AST nodes that are processed into
// "pure" first-order logic expressions evaluated by the truth package.
type Expr interface {
	Analyse(Table) (*AnalysedExpr, error)
	String() string

	replace(map[Expr]Expr) Expr
}

type SimpleExpr string

func (p SimpleExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	sym, ok := tbl[string(p)]
	if !ok {
		return nil, fmt.Errorf(errVariableNotDefined, p)
	}
	typ, ok := sym.(Type)
	if !ok {
		return nil, fmt.Errorf(errNonSimpleExpr, string(p))
	}
	return &AnalysedExpr{
		P:   truth.Variable(p),
		arg: Parameter{string(p), typ},
	}, nil
}

func (p SimpleExpr) String() string {
	return string(p)
}

func (p SimpleExpr) replace(m map[Expr]Expr) Expr {
	return m[p]
}

type PostfixExpr struct {
	Name string
	Args []Expr
}

func argsToParams(args []Expr, tbl Table) ([]Parameter, error) {
	params := make([]Parameter, len(args))
	for i := range params {
		expr, err := args[i].Analyse(tbl)
		if err != nil {
			return nil, err
		}
		params[i] = expr.arg
	}
	return params, nil
}

func paramsToStrings(params []Parameter) []string {
	sarr := make([]string, len(params))
	for i := range params {
		sarr[i] = params[i].Name
	}
	return sarr
}

type invocable interface {
	IsInvocation([]Parameter) error
}

func getinvocable(sym Symbol, tbl Table) (invocable, error) {
	if f, ok := sym.(Function); ok {
		return f, nil
	} else if t, ok := sym.(Template); ok {
		return t, nil
	}
	return nil, errors.New("not invocable")
}

func (p PostfixExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	sym, ok := tbl[p.Name]
	if !ok {
		return nil, fmt.Errorf(errFunctionOrTemplateNotDefined, p.Name)
	}
	inv, err := getinvocable(sym, tbl)
	if err != nil {
		return nil, fmt.Errorf(errNonInvocableInvoked, p.Name)
	}
	params, err := argsToParams(p.Args, tbl)
	if err != nil {
		return nil, err
	}
	if err := inv.IsInvocation(params); err != nil {
		return nil, err
	}
	name := fmt.Sprintf(
		"%s(%s)", p.Name, strings.Join(paramsToStrings(params), ", "),
	)
	return &AnalysedExpr{
		P:   truth.Variable(name),
		arg: Parameter{name, Bool},
	}, nil
}

func (p PostfixExpr) String() string {
	sarr := make([]string, len(p.Args))
	for i, arg := range p.Args {
		sarr[i] = arg.String()
	}
	return fmt.Sprintf("%s(%v)", p.Name, strings.Join(sarr, ", "))
}

func (p PostfixExpr) replace(m map[Expr]Expr) Expr {
	args := make([]Expr, len(p.Args))
	for i := range p.Args {
		args[i] = p.Args[i].replace(m)
	}
	return PostfixExpr{Name: p.Name, Args: args}
}

type ConstantExpr bool

func (c ConstantExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	return &AnalysedExpr{
		P:   truth.Constant(c),
		arg: Parameter{fmt.Sprintf("%t", c), Bool},
	}, nil
}

func (c ConstantExpr) String() string {
	return fmt.Sprintf("%t", c)
}

func (c ConstantExpr) replace(_ map[Expr]Expr) Expr {
	return c
}

type BracketedExpr struct {
	Expr
}

func (br BracketedExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	return br.Expr.Analyse(tbl)
}

func (br BracketedExpr) String() string {
	return fmt.Sprintf("(%s)", br.Expr)
}

func (br BracketedExpr) replace(m map[Expr]Expr) Expr {
	return BracketedExpr{br.Expr.replace(m)}
}

type NegatedExpr struct {
	Expr
}

func analyseProp(e Expr, tbl Table) (*AnalysedExpr, error) {
	aExpr, err := e.Analyse(tbl)
	if err != nil {
		return nil, err
	}
	if !aExpr.arg.isBool() {
		return nil, fmt.Errorf(
			errOpOnNonBoolExpr,
			reflect.TypeOf(e), aExpr.arg.Name, aExpr.arg.Type,
		)
	}
	return aExpr, nil
}

func (n NegatedExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	aExpr, err := analyseProp(n.Expr, tbl)
	if err != nil {
		return nil, err
	}
	return &AnalysedExpr{
		P:   truth.Not(aExpr.P),
		arg: Parameter{fmt.Sprintf("!(%s)", aExpr.arg.Name), Bool},
	}, nil
}

func (n NegatedExpr) String() string {
	return fmt.Sprintf("!%s", n.Expr)
}

func (n NegatedExpr) replace(m map[Expr]Expr) Expr {
	return NegatedExpr{n.Expr.replace(m)}
}

type Operator string

const (
	And  Operator = "&&"
	Or            = "||"
	Eqv           = "==="
	Impl          = "==>"
	Fllw          = "<=="
)

type BinaryOpExpr struct {
	Op     Operator
	E1, E2 Expr
}

type binaryOp func(truth.Proposition, truth.Proposition) truth.Proposition

func (op Operator) simpleTruthOp() binaryOp {
	switch op {
	case And:
		return truth.And
	case Or:
		return truth.Or
	case Eqv:
		return truth.Eqv
	case Impl:
		return truth.Impl
	case Fllw:
		return truth.Fllw
	default:
		panic(fmt.Sprintf("invalid simple truth op %s", op))
	}
}

func (b BinaryOpExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	aExpr1, err := analyseProp(b.E1, tbl)
	if err != nil {
		return nil, err
	}
	aExpr2, err := analyseProp(b.E2, tbl)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s %s %s", aExpr1.arg.Name, b.Op, aExpr2.arg.Name)
	return &AnalysedExpr{
		P:   b.Op.simpleTruthOp()(aExpr1.P, aExpr2.P),
		arg: Parameter{name, Bool},
	}, nil
}

func (b BinaryOpExpr) String() string {
	return fmt.Sprintf("%s %s %s", b.E1, b.Op, b.E2)
}

func (b BinaryOpExpr) replace(m map[Expr]Expr) Expr {
	return BinaryOpExpr{b.Op, b.E1.replace(m), b.E2.replace(m)}
}

type JustifiableBinaryOpExpr struct {
	BinaryOpExpr
	Just *PostfixExpr
}

func (b JustifiableBinaryOpExpr) instantiateJustification(tbl Table) (truth.Proposition, error) {
	if b.Just == nil {
		return truth.Constant(true), nil
	}
	if _, err := b.Just.Analyse(tbl); err != nil {
		// TODO: error
		return nil, err
	}
	sym, ok := tbl[b.Just.Name]
	if !ok {
		// TODO: error
		return nil, fmt.Errorf(errFunctionOrTemplateNotDefined, b.Just.Name)
	}
	tmpl, ok := sym.(Template)
	if !ok {
		// TODO: nonTemplateInvoked
		return nil, fmt.Errorf(errNonInvocableInvoked, b.Just.Name)
	}
	just, err := tmpl.instantiate(b.Just.Args, tbl)
	if err != nil {
		// TODO: error
		return nil, err
	}
	return just, nil
}

func justify(just, P, Q truth.Proposition, op Operator) truth.Proposition {
	switch op {
	case Impl:
		return truth.Impl(truth.And(just, P), Q)
	case Fllw:
		return truth.Fllw(P, truth.And(just, Q))
	case Eqv:
		return truth.Eqv(truth.And(just, P), truth.And(just, Q))
	default:
		panic(fmt.Sprintf("cannot justify %s", op))
	}
}

func (b JustifiableBinaryOpExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	aExpr1, err := analyseProp(b.E1, tbl)
	if err != nil {
		return nil, err
	}
	aExpr2, err := analyseProp(b.E2, tbl)
	if err != nil {
		return nil, err
	}
	just, err := b.instantiateJustification(tbl)
	if err != nil {
		// TODO error
		return nil, err
	}
	name := fmt.Sprintf("%s %s %s", aExpr1.arg.Name, b.Op, aExpr2.arg.Name)
	return &AnalysedExpr{
		P:   justify(just, aExpr1.P, aExpr2.P, b.Op),
		arg: Parameter{name, Bool},
	}, nil
}

func (b JustifiableBinaryOpExpr) String() string {
	if b.Just == nil {
		return fmt.Sprintf("%s %s %s", b.E1, b.Op, b.E2)
	}
	return fmt.Sprintf("%s %s %s by %s", b.E1, b.Op, b.E2, *b.Just)
}

// RelationChain is a series of JustifiableBinaryOpExpr's each with the operation being one
// of the links, i.e. `==>`, `===`, or `<==`.
type RelationChain []JustifiableBinaryOpExpr

// quantiseWithSides is a utility function used by Quantise to break an Expr
// into its component (quantised) subrelations and its zeroth and final term.
func quantiseWithSides(E Expr) (RelationChain, Expr, Expr) {
	if r, ok := E.(JustifiableBinaryOpExpr); ok {
		quant := r.Quantise()
		if len(quant) == 0 {
			panic("empty quantisation")
		}
		return quant, quant[0].E1, quant[len(quant)-1].E2
	}
	return RelationChain{}, E, E
}

// Quantise takes a JustifiableBinaryOpExpr that is a relation (i.e. has as its operation
// one of the links) and breaks it into a sequence of relations that can be
// evaluated one-by-one in order to compute the truth or falsehood of the
// original relation.
func (b JustifiableBinaryOpExpr) Quantise() RelationChain {
	// the chain is the concatenation of E1's and E2's quantisations with
	// an adjoining (pivot) element formed by relating the final item in
	// E1's chain with the initial one in E2's chain with b.Op
	previous, _, lhs := quantiseWithSides(b.E1)
	following, rhs, _ := quantiseWithSides(b.E2)
	pivot := JustifiableBinaryOpExpr{
		BinaryOpExpr: BinaryOpExpr{Op: b.Op, E1: lhs, E2: rhs},
		Just:         b.Just,
	}
	return append(append(previous, pivot), following...)
}
