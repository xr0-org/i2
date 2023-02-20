%{
	package parser

	import (
		/*"fmt"*/
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

/* primary */
%token <s> tkIdentifier tkFalse tkTrue
%token <n> tkConstant

/* operators */
%token <s> tkLt tkGt tkEq tkNe tkAnd tkOr tkEqv tkImpl tkFllw

/* keywords */ 
%token <s> tkTmpl tkFunc tkThis

%%
statement_list
	: modifiers statement ';' statement_list
	| /* empty */
	;

modifiers
	: '@' /* axiom */
	| /* empty */
	;

statement
	: tkTmpl tkIdentifier template
	| tkFunc tkIdentifier function
	;

template
	: '(' value_list ')' '{' proposition '}'
	| '(' ')' '{' proposition '}'
	| template '{' proposition '}'
	;

function
	: '(' type_assertion_list ')' predicate 
	| '(' value_list ')' predicate 
	;

type_assertion_list
	: value_list predicate
	;

predicate
	: tkIdentifier
	| tkTrue
	| tkFalse
	| tkFunc function

value_list
	: value_list ',' value
	| value
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
