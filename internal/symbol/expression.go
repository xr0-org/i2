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

	replace(map[string]Expr) Expr
}

func paramsToTypes(params []Parameter) []string {
	arr := make([]string, len(params))
	for i := range params {
		arr[i] = string(params[i].Type)
	}
	return arr
}

type SimpleExpr string

func (p SimpleExpr) analyseThis(tbl Table) (*AnalysedExpr, error) {
	if string(p) != "this" {
		// XXX: error
		return nil, fmt.Errorf("not this")
	}
	this, ok := tbl["this"]
	if !ok {
		// XXX: error
		return nil, fmt.Errorf("`this' not set")
	}
	tmpl, ok := this.(Template)
	if !ok {
		// XXX: error
		return nil, fmt.Errorf("`this' not template")
	}
	aExpr, err := tmpl.E.Analyse(tbl)
	if err != nil {
		// XXX: error
		return nil, err
	}
	return &AnalysedExpr{
		P: aExpr.P,
		arg: Parameter{
			"this", Type(fmt.Sprintf(
				// must be predicate because from template
				"func(%s) bool",
				strings.Join(paramsToTypes(tmpl.Params), ", "),
			)),
		},
	}, nil
}

func (p SimpleExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	if expr, err := p.analyseThis(tbl); err == nil {
		return expr, nil
	}
	sym, ok := tbl[string(p)]
	if !ok {
		return nil, fmt.Errorf(errVariableNotDefined, p)
	}
	if lp, ok := sym.(LocalProof); ok {
		return lp.Expr.Analyse(tbl)
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

func (p SimpleExpr) replace(m map[string]Expr) Expr {
	if repl, ok := m[string(p)]; ok {
		return repl
	}
	return p
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

func getinvocable(sym Scope, tbl Table) (invocable, error) {
	if f, ok := sym.(Function); ok {
		return f, nil
	} else if t, ok := sym.(Template); ok {
		return t, nil
	}
	return nil, errors.New("not invocable")
}

func invocabletype(sym Scope) Type {
	if f, ok := sym.(Function); ok {
		return f.Sig.Return
	}
	// template expressions must be boolean
	return Bool
}

func (p PostfixExpr) analyseThis(tbl Table) (*AnalysedExpr, error) {
	this, ok := tbl["this"]
	if !ok {
		// XXX: error
		return nil, fmt.Errorf("`this' not set")
	}
	tmpl, ok := this.(Template)
	if !ok {
		// XXX: error
		return nil, fmt.Errorf("`this' not template")
	}
	prop, err := tmpl.instantiate(p.Args, tbl)
	if err != nil {
		// XXX: error
		return nil, fmt.Errorf("instantiate error: %s", err)
	}
	return &AnalysedExpr{
		P:   prop,
		arg: Parameter{fmt.Sprintf("%s", prop), Bool},
	}, nil
}

func (p PostfixExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	if p.Name == "this" {
		expr, err := p.analyseThis(tbl)
		if err != nil {
			return nil, fmt.Errorf("this error: %s", err)
		}
		return expr, nil
	}
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
		return nil, fmt.Errorf("invocation error: %s", err)
	}
	name := fmt.Sprintf(
		"%s(%s)", p.Name, strings.Join(paramsToStrings(params), ", "),
	)
	return &AnalysedExpr{
		P:   truth.Variable(name),
		arg: Parameter{name, invocabletype(sym)},
	}, nil
}

func (p PostfixExpr) String() string {
	sarr := make([]string, len(p.Args))
	for i, arg := range p.Args {
		sarr[i] = arg.String()
	}
	return fmt.Sprintf("%s(%v)", p.Name, strings.Join(sarr, ", "))
}

func replaceName(name string, m map[string]Expr) string {
	if repl, ok := m[name]; ok {
		simp, ok := repl.(SimpleExpr)
		if !ok {
			panic(fmt.Sprintf("cannot replace name with %s", repl))
		}
		return string(simp)
	}
	return name
}

func (p PostfixExpr) replace(m map[string]Expr) Expr {
	args := make([]Expr, len(p.Args))
	for i := range p.Args {
		args[i] = p.Args[i].replace(m)
	}
	return PostfixExpr{Name: replaceName(p.Name, m), Args: args}
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

func (c ConstantExpr) replace(_ map[string]Expr) Expr {
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

func (br BracketedExpr) replace(m map[string]Expr) Expr {
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

func (n NegatedExpr) replace(m map[string]Expr) Expr {
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

type BinaryOp func(truth.Proposition, truth.Proposition) truth.Proposition

func (op Operator) SimpleTruthOp() BinaryOp {
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
		P:   b.Op.SimpleTruthOp()(aExpr1.P, aExpr2.P),
		arg: Parameter{name, Bool},
	}, nil
}

func (b BinaryOpExpr) String() string {
	return fmt.Sprintf("%s %s %s", b.E1, b.Op, b.E2)
}

func (b BinaryOpExpr) replace(m map[string]Expr) Expr {
	return BinaryOpExpr{b.Op, b.E1.replace(m), b.E2.replace(m)}
}

type JustifiableBinaryOpExpr struct {
	BinaryOpExpr
	Just *PostfixExpr
}

func (b JustifiableBinaryOpExpr) instantiateJustification(tbl Table) (truth.Proposition, error) {
	if b.Just == nil {
		return nil, fmt.Errorf("cannot justify with nil")
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
	if b.Just == nil {
		return b.BinaryOpExpr.Analyse(tbl)
	}
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
		return fmt.Sprintf("%s %s %s [UJ]", b.E1, b.Op, b.E2)
	}
	return fmt.Sprintf("%s %s %s by %s", b.E1, b.Op, b.E2, *b.Just)
}

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

// Quantise takes a JustifiableBinaryOpExpr that is a relation (i.e. has as its
// operation one of the links) and breaks it into a sequence of relations that
// can be evaluated one-by-one in order to compute the truth or falsehood of the
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

// RelationChain is a series of JustifiableBinaryOpExpr's each with the
// operation being one of the links, i.e. `==>`, `===`, or `<==`.
type RelationChain []JustifiableBinaryOpExpr

var ErrNoCommonOperator = errors.New("no common operator")

// GCF returns the greatest-common-factor-operator, if it exists, or an error
// otherwise.
func (rel RelationChain) GCFOperator() (Operator, error) {
	if len(rel) == 0 {
		return "", fmt.Errorf("zero width chain")
	}
	var gcf Operator = Eqv
	for _, b := range rel {
		switch b.Op {
		case Eqv:
			continue
		case Impl, Fllw:
			if gcf == Eqv {
				gcf = b.Op
			} else if gcf != b.Op {
				return "", ErrNoCommonOperator
			}
			continue
		default:
			return "", fmt.Errorf("invalid op")
		}
	}
	return gcf, nil
}

func (rel RelationChain) Chain() RelationChain {
	return rel
}

func (rel RelationChain) Burden() (Expr, error) {
	op, err := rel.GCFOperator()
	if err != nil {
		return nil, err
	}
	return BinaryOpExpr{
		Op: op, E1: rel[0].E1, E2: rel[len(rel)-1].E2,
	}, nil
}

func (rel RelationChain) Label() string {
	return ""
}

type Proof interface {
	Chain() RelationChain
	Burden() (Expr, error)
	Label() string
}

type LabelledChain struct {
	RelationChain
	L string
}

func (c LabelledChain) Label() string {
	return c.L
}

type ProofChain struct {
	Preamble []Proof
	Proof    Proof
}

type LambdaExpr struct {
	Params []Parameter
	Expr   Expr
}

func (λ LambdaExpr) Table() (Table, error) {
	T := Table{}
	for _, p := range λ.Params {
		T[p.Name] = p.Type
	}
	return T, nil
}

func (λ LambdaExpr) Analyse(tbl Table) (*AnalysedExpr, error) {
	λtbl, err := λ.Table()
	if err != nil {
		return nil, fmt.Errorf("table error: %s", err)
	}
	aExpr, err := λ.Expr.Analyse(λtbl.Nest(tbl))
	if err != nil {
		return nil, fmt.Errorf("expr analysis error: %s", err)
	}
	pseudolambda := fmt.Sprintf(
		"ψ(%s) { %s }",
		strings.Join(paramsToTypeAssertionList(λ.Params), ", "),
		aExpr.P.String(),
	)
	return &AnalysedExpr{
		P:   truth.Variable(pseudolambda),
		arg: Parameter{pseudolambda, Bool},
	}, nil
}

func paramsToTypeAssertionList(params []Parameter) []string {
	list := make([]string, len(params))
	for i, p := range params {
		list[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
	}
	return list
}

func (λ LambdaExpr) String() string {
	return fmt.Sprintf(
		"(%s) { %s }",
		strings.Join(paramsToTypeAssertionList(λ.Params), ", "),
		λ.Expr.String(),
	)
}

func (λ LambdaExpr) replace(m map[string]Expr) Expr {
	for _, p := range λ.Params {
		for k := range m {
			if p.Name == k {
				panic("bound replace NOT IMPLEMENTED")
			}
		}
	}
	return LambdaExpr{λ.Params, λ.Expr.replace(m)}
}

type LambdaProof struct {
	E LambdaExpr
	L string
}

func (prf LambdaProof) Chain() RelationChain {
	jop, ok := prf.E.Expr.(JustifiableBinaryOpExpr)
	if !ok {
		panic(fmt.Sprintf("not justifiable lambda proof"))
	}
	return jop.Quantise()
}

func (prf LambdaProof) Burden() (Expr, error) {
	relburden, err := prf.Chain().Burden()
	if err != nil {
		return nil, err
	}
	return LambdaExpr{prf.E.Params, relburden}, nil
}

func (prf LambdaProof) Label() string {
	return prf.L
}
