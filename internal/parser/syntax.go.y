%{
	package parser

	import (
		"fmt"

		"git.sr.ht/~lbnz/i2/internal/symbol"
		"git.sr.ht/~lbnz/i2/internal/truth"
	)

	var sigma = symbol.Table{"1": symbol.Any}
%}

%union{
	s		string
	sarr		[]string
	n		int
	b		bool

	sym_op		symbol.Operator
	sym_tmpl	symbol.Template
	sym_func	symbol.Function
	sym_type	symbol.Type
	sym_paramarr	[]symbol.Parameter

	sym_expr	symbol.Expr
	sym_exprarr	[]symbol.Expr
	sym_pfexpr	symbol.PostfixExpr
	sym_pfexpr_ptr	*symbol.PostfixExpr
}

%type <b> axiom
%type <s> value type

%type <sym_op> connective 
%type <sym_tmpl> template
%type <sym_func> function

%type <sym_paramarr> type_assertion_list

%type <sym_expr> expression logical_and_expression logical_or_expression negated_expression simple_expression constant_expression

%type <sym_exprarr> argument_list
%type <sym_pfexpr> postfix_expresion
%type <sym_pfexpr_ptr> justification

/* primary */
%token <s> tkIdentifier tkConstant tkFalse tkTrue

/* operators */
%token <s> tkLt tkGt tkEq tkNe tkAnd tkOr tkEqv tkImpl tkFllw

/* keywords */ 
%token <s> tkTmpl tkFunc tkThis

%%
statement_list
	: statement ';' statement_list
	| /* empty */
	;

axiom
	: '@'		{ $$ = true }
	| /* empty */	{ $$ = false }
	;

statement
	: axiom tkTmpl tkIdentifier template	{
		$4.IsAxiom = $1
		sigma[$3] = $4

		tbl, err := $4.Table()
		if err != nil {
			yylex.Error(err.Error())
		}
		aExpr, err := $4.E.Analyse(tbl.Nest(sigma))
		if err != nil {
			yylex.Error(err.Error())
		}
		result.WriteString(fmt.Sprintf("tmpl %s %v\n", $3, aExpr))
		for _, prf := range $4.Proofs {
			result.WriteString(fmt.Sprintf("proof:\n"))
			for _, expr := range prf {
				result.WriteString(fmt.Sprintf("\t%s\n", expr))
				aExpr, err := expr.Analyse(tbl.Nest(sigma))
				if err != nil {
					yylex.Error(err.Error())
				}
				outcome, err := truth.Decide(aExpr.P)
				if err != nil {
					yylex.Error(err.Error())
				}
				if !outcome {
					yylex.Error("contradiction")
				}
			}
			result.WriteString(fmt.Sprintf("qed\n"))
		}
	}
	| axiom tkFunc tkIdentifier function	{
		$4.IsAxiom = $1
		sigma[$3] = $4
	}
	;

template
	: '(' type_assertion_list ')' '{' expression '}'	{
		$$ = symbol.Template{
			Params:	$2,
			E:	$5,
			Proofs:	[]symbol.RelationChain{},
		}
	}
	| '(' ')' '{' expression '}'			{
		$$ = symbol.Template{
			Params:	[]symbol.Parameter{},
			E:	$4,
			Proofs:	[]symbol.RelationChain{},
		}
	}
	| template '{' expression '}'			{
		prf, ok := $3.(symbol.JustifiableBinaryOpExpr)
		if !ok {
			yylex.Error("non bop proof")
		}
		$$ = symbol.Template{
			Params:	$$.Params,
			E: 	$$.E,
			Proofs:	append($$.Proofs, prf.Quantise()),
		}
	}
	;

function
	: '(' type_assertion_list ')' type {
		$$ = symbol.Function{
			Sig: symbol.FunctionSignature{
				Params: $2,
				Return: symbol.Type($4),
			},
		}
	}
	;

type_assertion_list
	: type_assertion_list ',' value type
		{ $$ = append($1, symbol.Parameter{ Name: $3, Type: symbol.Type($4) }) }
	| value type
		{ $$ = []symbol.Parameter{symbol.Parameter{Name: $1, Type: symbol.Type($2)}}}
	;

value
	: tkConstant		{ $$ = $1 }
	| tkIdentifier		{ $$ = $1 }
	;

type
	: tkIdentifier		{ $$ = $1 }
	| tkFunc function	{ $$ = "func" }
	;

expression
	: expression connective justification logical_or_expression { 
		$$ = symbol.JustifiableBinaryOpExpr{
			BinaryOpExpr:	symbol.BinaryOpExpr{Op: $2, E1: $1, E2: $4},
			Just: 		$3,
		}
	}
	| type_assertion_list	{
		// TODO reconcile this with Expr
		$$ = symbol.ConstantExpr(false)
	}
	| logical_or_expression 
	;

justification
	: '{' postfix_expresion '}'
		{ var x = $2; $$ = &x }
	| /* empty */ 
		{ $$ = nil }
	;

connective
	: tkEqv		{ $$ = symbol.Eqv }
	| tkImpl	{ $$ = symbol.Impl }
	| tkFllw	{ $$ = symbol.Fllw }
	;

logical_or_expression
	: logical_or_expression tkOr logical_and_expression
		{ $$ = symbol.BinaryOpExpr{Op: symbol.Or, E1: $1, E2: $3} }
	| logical_and_expression
	;

logical_and_expression
	: logical_and_expression tkAnd negated_expression
		{ $$ = symbol.BinaryOpExpr{Op: symbol.And, E1: $1, E2: $3} }
	| negated_expression
	;

negated_expression
	: '!' negated_expression	{ $$ = symbol.NegatedExpr{$2} }
	| constant_expression
	;

constant_expression
	: tkTrue			{ $$ = symbol.ConstantExpr(true) }
	| tkFalse			{ $$ = symbol.ConstantExpr(false) }
	| '(' expression ')'		{ $$ = symbol.BracketedExpr{$2} }
	| simple_expression
	;

/* `tkThis` disabled for now */
simple_expression
	: tkIdentifier			{ $$ = symbol.SimpleExpr($1) }
	| tkConstant			{ $$ = symbol.SimpleExpr($1) }
	| postfix_expresion		{ $$ = symbol.Expr($1) }
	;

argument_list
	: argument_list ',' expression	{ $$ = append($1, $3) }
	| expression			{ $$ = []symbol.Expr{$1} }
	;

postfix_expresion
	: tkIdentifier '(' argument_list ')'
		{ $$ = symbol.PostfixExpr{$1, $3} }
	| tkIdentifier '(' ')'
		{ $$ = symbol.PostfixExpr{$1, []symbol.Expr{}} }
	;
%%
