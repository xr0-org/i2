%{
	package parser
%}

/* X_IDENTIFIER is axiom */
%token IDENTIFIER X_IDENTIFIER CONSTANT LITERAL

/* value keywords */
%token THIS ABSURD

/* module keywords */
%token MOD IMPORT EXPORT

/* binders */
%token FOR SOME LAMBDA

/* modifiers */
%token AUTO

/* block types */
%token TYPE FUNC

/* conventions */
%token THIS CONTRADICTION

/* boolean connectives: ===, ==>, <== */
%token EQUIV_OP IMPLY_OP FOLLOW_OP

%token LE_OP GE_OP EQ_OP NE_OP AND_OP OR_OP

%start file
%%
file
	: header stmts footer
     	;

header
	: MOD IDENTIFIER IMPORT '{' imports '}' ';'
       	;

imports	
	: /* empty */
	| imports '\n' LITERAL
	| imports ',' LITERAL
	;

footer
	: /* empty */
       	| EXPORT opbang '{' exports '}' ';'
	;

exports
	: /* empty */
	| exports '\n' IDENTIFIER
	| exports ',' IDENTIFIER
	;

opbang
	: /* empty */
       	| ';'
	;

stmts
	: /* empty */
      	| stmt ';' stmts
	;

stmt
	: TYPE identifier br_type_list type_body 
     	| FUNC identifier br_type_list predicate func_body
	| identifier ':' theorem_body 
	;

type_body
	: /* empty */
	| ':' proposition
	;

br_type_list
	: '(' type_list ')'
	| '(' ')'
	;

type_list
	: id_list
	| id_list IDENTIFIER
	| type_list ',' type_list
	;

id_list
	: id_list ',' IDENTIFIER
	| IDENTIFIER
	;

identifier
	: IDENTIFIER
	| X_IDENTIFIER
	;

/* TODO */
func_body
	: /* empty */
	;

theorem_body
	: proposition
	| proposition '{' proof '}'
	;

proof
	: theorem_body
	| deductive_seq
	;

deductive_seq
	: deductive_seq triple_op justification proposition
	| proposition
	;

triple_op
	: EQUIV_OP
	| IMPLY_OP
	| FOLLOW_OP
	;

justification
	: /* empty */
	| '{' appeal '}' justification_application
	;

appeal
	: IDENTIFIER
	| IDENTIFIER '(' argument_expression_list ')'
	;

justification_application
	: /* empty */
	| '~' '[' application_list ']'
	;

application_list
	: IDENTIFIER
	| constant_range
	| application_list ',' application_list
	;

constant_range
	: CONSTANT
	| CONSTANT ':'
	| CONSTANT ':' CONSTANT
	| ':' CONSTANT
	;

predicate
	: IDENTIFIER
	| predicate '*' predicate
	| '(' predicate ')'
	;

binder
	: FOR
	| SOME
	| LAMBDA
	;

proposition
	: logical_or_expression
	| binder '(' type_list ')' logical_or_expression
	;

logical_or_expression
	: logical_and_expression
	| logical_or_expression OR_OP logical_and_expression

logical_and_expression
	: equality_expression
	| logical_and_expression AND_OP equality_expression
	;

equality_expression
	: relational_expression
	| equality_expression EQ_OP relational_expression
	| equality_expression NE_OP relational_expression
	;

relational_expression
	: additive_expression
	| relational_expression '<' additive_expression
	| relational_expression '>' additive_expression
	| relational_expression LE_OP additive_expression
	| relational_expression GE_OP additive_expression
	;

additive_expression
	: multiplicative_expression
	| additive_expression '+' multiplicative_expression
	| additive_expression '-' multiplicative_expression
	;

multiplicative_expression
	: unary_expression
	| multiplicative_expression '*' unary_expression
	| multiplicative_expression '/' unary_expression
	| multiplicative_expression '%' unary_expression
	;

unary_expression
	: '!' unary_expression
	| postfix_expression
	;

argument_expression_list
	: argument_expression_list ',' proposition
	| proposition
	;

postfix_expression
	: primary_expression
	| primary_expression '(' argument_expression_list ')'
	;

primary_expression
	: IDENTIFIER
	| CONSTANT
	| LITERAL
	| ABSURD
	| this_expression
	| '(' proposition ')'
	;

this_expression
	: THIS
	| THIS '(' ')'
	| THIS '{' '}'
	;
