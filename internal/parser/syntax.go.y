%{
	package parser

	import (
		"fmt"
	)

	type identifier struct {
		s string
		axiom bool
	}
%}

%union{
	s	string
	id	identifier
	idlst	[]identifier
}

%type <id> identifier
%type <idlst> identifier_list

%token <s> tkId tkType

%%
statements
	: statement ';' statements
	| /* empty */
	;

statement
	: tkType identifier '(' identifier_list ')'
		{ fmt.Printf("type: %v %v\n", $2, $4) }
	;

identifier_list
	: identifier_list ',' identifier	{ $$ = append($1, $3) }
	| identifier				{ $$ = []identifier{$1} }
	;

identifier
	: tkId					{ $$ = identifier{$1, false} }
	| '@' tkId				{ $$ = identifier{$2, true} }
	;
