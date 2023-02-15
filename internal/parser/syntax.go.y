%{
	package parser

	import (
		"fmt"
	)

	type statement struct {
	}
%}

%union{
	s	string
	sarr	[]string
	n	int
	narr	[]int
	stmt	statement
}

%type <s> modifier
%type <sarr> identifier_list modifier_list
%type <narr> constant_list

/* primary */
%token <s> tkIdentifier tkFalse tkTrue
%token <n> tkConstant

/* operators */
%token <s> tkLt tkGt tkEq tkNe tkAnd tkOr tkEqv tkImpl tkRevImpl

/* keywords */ 
%token <s> tkType tkAuto tkFunc tkFor

%%
statements
	: modifier_list statement ';' statements
	| /* empty */
	;

statement
	: tkIdentifier '{' identifier_list '}' ':' proposition
	| function
	;

function
	: tkFunc tkIdentifier '(' identifier_list ')' predicate 
	| tkFunc tkIdentifier '(' type_assertion_list ')' predicate 
	;

type_assertion_list
	: identifier_list tkIdentifier
		{ fmt.Printf("type: %v %v\n", $1, $2) }
	| constant_list tkIdentifier
		{ fmt.Printf("type: %d %v\n", $1, $2) }
	;

constant_list
	: constant_list ',' tkConstant		{ $$ = append($1, $3) }
	| tkConstant				{ $$ = []int{$1} }
	;

modifier_list
	: modifier_list modifier		{ $$ = append($1, $2) }
	| /* empty */				{ $$ = []string{} }
	;

modifier
	: '@'					{ $$ = "@" }
	| tkAuto
	;

identifier_list
	: identifier_list ',' tkIdentifier	{ $$ = append($1, $3) }
	| tkIdentifier				{ $$ = []string{$1} }
	;

predicate
	: tkIdentifier
	;

proposition
	: proposition tkEqv binding_proposition
	| proposition tkImpl binding_proposition
	| proposition tkRevImpl binding_proposition
	| binding_proposition
	;

binding_proposition
	: tkFor '(' type_assertion_list ')' binding_proposition
		{ fmt.Println("for") }
	| logical_or_proposition
	| '{' logical_or_proposition '}'
	;

logical_or_proposition
	: logical_or_proposition tkOr logical_and_proposition
	| logical_and_proposition
	;

logical_and_proposition
	: logical_and_proposition tkAnd unary_proposition
	| unary_proposition
	;

unary_proposition
	: '!' unary_proposition
	| primary_proposition
	;

primary_proposition
	: relational_proposition
	| type_assertion_list
	| tkTrue
	| tkFalse
	| '(' proposition ')'
	;

relational_proposition
	: relational_proposition order expression
	| expression order expression
	;

order
	: '<'
	| '>'
	| tkEq
	| tkNe
	| tkGt
	| tkLt
	;

expression
	: expression '+' multiplicative_expression
	| expression '-' multiplicative_expression
	| multiplicative_expression
	;

multiplicative_expression
	: multiplicative_expression '*' postfix_expression
	| multiplicative_expression '/' postfix_expression
	| multiplicative_expression '%' postfix_expression
	| postfix_expression
	;

argument_list
	: argument_list ',' proposition
	| proposition
	;

postfix_expression
	: postfix_expression '(' argument_list ')'
	| postfix_expression '(' identifier_list ')'
	| primary_expression
	;

primary_expression
	: tkIdentifier
	| tkConstant
	| '(' expression ')'
	;
