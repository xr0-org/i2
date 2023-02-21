%{
	package parser

	import (
		"git.sr.ht/~lbnz/i2/internal/symbol"
	)

	func installAny(T symbol.Table, values []string) {
		for _, s := range values {
			T[s] = symbol.Any
		}
	}

	type typedlist struct {
		vars []string
		pred symbol.Type
	}

	func installTypedlist(T symbol.Table, list typedlist) {
		for _, v := range list.vars {
			T[v] = list.pred
		}
	}

	var sigma = symbol.Table{}
%}

%union{
	s		string
	sarr		[]string
	n		int
	b		bool
	sym_tmpl	symbol.TemplateSymbol
	sym_func	symbol.FunctionSymbol
	sym_type	symbol.Type
	typed_list	typedlist
}

%type <b> axiom
%type <s> value
%type <sarr> value_list
%type <sym_tmpl> template
%type <sym_func> function
%type <sym_type> predicate
%type <typed_list> type_assertion_list

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
	}
	| axiom tkFunc tkIdentifier function	{
		$4.IsAxiom = $1
		sigma[$3] = $4
	}
	;

template
	: '(' value_list ')' '{' proposition '}'	{
		t := symbol.Table{}
		installAny(t, $2)
		$$ = symbol.TemplateSymbol{T: t}
	}
	| '(' ')' '{' proposition '}'			{
		$$ = symbol.TemplateSymbol{T: symbol.Table{}}
	}
	| template '{' proposition '}'
	;

function
	: '(' type_assertion_list ')' predicate	{
		t := symbol.Table{}
		installTypedlist(t, $2)
		$$ = symbol.FunctionSymbol{
			T:	t,
			Return: $4,
		}
	}
	| '(' value_list ')' predicate {
		t := symbol.Table{}
		installAny(t, $2)
		$$ = symbol.FunctionSymbol{
			T:	t,
			Return: $4,
		}
	}
	;

type_assertion_list
	: value_list predicate	{ $$ = typedlist{$1, $2} }
	;

predicate
	: tkIdentifier		{ $$ = symbol.Type($1) }
	| tkFunc function	{ $$ = symbol.Type("func") }
	;

value_list
	: value_list ',' value	{ $$ = append($1, $3) }
	| value			{ $$ = []string{$1} }
	;

value
	: tkConstant
	| tkIdentifier
	;

justification
	: '{' tkIdentifier '(' argument_list ')' '}' '~' '[' value_list ']'
	| /* empty */
	;

proposition
	: proposition connective justification logical_or_proposition
	| logical_or_proposition
	;

connective
	: tkEqv
	| tkImpl
	| tkFllw
	;

logical_or_proposition
	: logical_or_proposition tkOr logical_and_proposition
	| logical_and_proposition
	;

logical_and_proposition
	: logical_and_proposition tkAnd negated_proposition
	| negated_proposition
	;

negated_proposition
	: '!' negated_proposition
	| postfix_proposition
	;

argument_list
	: argument_list ',' proposition
	| proposition 
	;

postfix_proposition
	: postfix_proposition '(' argument_list ')'
	| primary_proposition
	;

primary_proposition
	: '(' value_list ')' '{' proposition '}'
	| type_assertion_list
	| value_list
	| tkThis '[' tkConstant ']'
	| tkTrue
	| tkFalse
	| '(' proposition ')'
	;
