%{
	package parser

	import (
		"fmt"
		"strings"

		"git.sr.ht/~lbnz/i2/internal/symbol"
	)

	var sigma = symbol.Table{}

	type labelledJust struct {
		expr symbol.Expr
		label string
	}
%}

%union{
	s		string
	sarr		[]string
	n		int
	b		bool
	proof		[]labelledJust

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
%type <sarr> value_list

%type <sym_op> connective 
%type <sym_tmpl> template
%type <sym_func> function

%type <sym_paramarr> type_assertion_list

%type <sym_expr> expression logical_and_expression logical_or_expression negated_expression simple_expression constant_expression
%type <proof> expression_list

%type <sym_exprarr> argument_list
%type <sym_pfexpr> postfix_expresion
%type <sym_pfexpr_ptr> justification

/* primary */
%token <s> tkIdentifier tkConstant tkFalse tkTrue

/* operators */
%token <s> tkLt tkGt tkEq tkNe tkAnd tkOr tkEqv tkImpl tkFllw

/* keywords */ 
%token <s> tkTmpl tkFunc tkTerm

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
		$4.Name = $3
		sigma[$3] = $4
		tbl, err := $4.Table()
		if err != nil {
			yylex.Error(err.Error())
		}
		fmt.Printf("%s: %s\n", $3, $4)
		for _, prf := range $4.Proofs {
			contextTbl := tbl.Nest(sigma)
			proven := []string{}
			for _, preprf := range prf.Preamble {
				if err := sound(
					preprf.Chain(), 
					contextTbl,
				); err != nil {
					yylex.Error(
						fmt.Sprintf(
						"preamble error: %s", err),
					)
				}
				burden, err := preprf.Burden()
				if err != nil {
					yylex.Error(
						fmt.Sprintf(
						"burden error: %s", err),
					)
				}
				if l := preprf.Label(); l != "" {
					contextTbl[l] = symbol.LocalProof{burden}
					proven = append(proven, l)
				}
			}
			err := examineProof(
				$4.E, prf.Proof.Chain(), proven, contextTbl,
			)
			if err != nil {
				yylex.Error(err.Error())
			}
		}
	}
	| axiom tkFunc tkIdentifier function	{
		$4.IsAxiom = $1
		$4.Name = $3
		sigma[$3] = $4
	}
	| tkTerm value type {
		sigma[$2] = symbol.Type($3)
	}
	;

template
	: '(' type_assertion_list ')' '{' expression '}'	{
		$$ = symbol.Template{
			Params:	$2,
			E:	$5,
			Proofs:	[]symbol.ProofChain{},
		}
	}
	| '(' ')' '{' expression '}'				{
		$$ = symbol.Template{
			Params:	[]symbol.Parameter{},
			E:	$4,
			Proofs:	[]symbol.ProofChain{},
		}
	}
	| template '{' expression_list '}'			{
		preambleLen := len($3) - 1
		if preambleLen < 0 {
			yylex.Error("empty proof")
		}
		preamble := make([]symbol.Proof, preambleLen)
		for i, just := range $3[:preambleLen] {
			var prf symbol.Proof
			if λ, ok := just.expr.(symbol.LambdaExpr); ok {
				prf = symbol.LambdaProof{λ, just.label}
			} else if expr, ok := just.expr.(symbol.JustifiableBinaryOpExpr); ok {
				prf = symbol.LabelledChain{
					expr.Quantise(), just.label,
				}
			}
			preamble[i] = prf
		}
		proper, ok := $3[len($3)-1].expr.(symbol.JustifiableBinaryOpExpr)
		if !ok {
			yylex.Error("non bop proof")
		}
		$$ = symbol.Template{
			Params:	$$.Params,
			E: 	$$.E,
			Proofs:	append($$.Proofs, symbol.ProofChain{
				preamble, proper.Quantise(),
			}),
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

type
	: tkIdentifier				{ $$ = $1 }
	| tkFunc '(' value_list ')' type	{ 
		$$ = fmt.Sprintf("func(%s) %s", strings.Join($3, ", "), $5)
	}
	;

value_list
	: value_list ',' value		{ $$ = append($1, $3) }
	| value				{ $$ = []string{$1} }
	;

value
	: tkConstant		{ $$ = $1 }
	| tkIdentifier		{ $$ = $1 }
	;

expression
	: expression connective justification expression { 
		$$ = symbol.JustifiableBinaryOpExpr{
			BinaryOpExpr:	symbol.BinaryOpExpr{Op: $2, E1: $1, E2: $4},
			Just: 		$3,
		}
	}
	| type_assertion_list	
		{ $$ = symbol.ConstantExpr(false) }
	| logical_or_expression 
	;

expression_list
	: expression_list expression_list
		{ $$ = append($1, $2...) }
	| tkIdentifier ':' expression ';' 
		{ $$ = []labelledJust{labelledJust{$3, $1}} }
	| expression ';' 
		{ $$ = []labelledJust{labelledJust{$1, ""}} }
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
	| '(' type_assertion_list ')' '{' expression '}' { 
		$$ = symbol.LambdaExpr{$2, $5} 
	}
	| simple_expression
	;

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
