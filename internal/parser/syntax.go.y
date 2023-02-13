%{
	package parser

	import (
		"fmt"
	)
%}

%union{
	val int
}

%type <val> expr term factor

%token <val> tkNumber

%%
line
	: expr '\n'		{ fmt.Printf("%d\n", $1) }
	;

expr
	: expr '+' term		{ $$ = $1 + $3 }
	| expr '-' term		{ $$ = $1 - $3 }
	| term
	;

term
	: term '*' factor	{ $$ = $1 * $3 }
	| term '/' factor	{ $$ = div($1, $3) }
	| factor
	;

factor
	: '(' expr ')'		{ $$ = $2 }
	| tkNumber
	;
