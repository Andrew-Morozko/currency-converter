grammar CurrencyConverter;
import CurrencyConverterSymbols;

root: line EOF;
line: implicit_src TO dst | explicit_src dst | implicit_src;
explicit_src: currency expr | expr currency;
implicit_src: explicit_src | expr;

dst: currency;

expr:
	num
	| '(' expr ')'
	| op = SUB right = expr
	| left = expr op = POW right = expr
	| left = expr op = (MUL | DIV) right = expr
	| left = expr op = (ADD | SUB) right = expr
	| left = expr op = (ADD | SUB) right_pct = persent_expr;

persent_expr: num PERCENT | '(' expr ')' PERCENT;

num: sepnum;

sepnum: sepnum sepnum | sepnum SEP+ sepnum | SEP+ sepnum | NUM;

currency: name = CURNAME | sym = CURSIGN;

CURNAME: [A-Za-z][A-Za-z][A-Za-z];

NUM: [0-9]+;

SEP: [.,`'];
TO: [tT][oO] | [iI][nN];

ADD: '+';
SUB: '-';
MUL: '*';
DIV: '/';
POW: '^';
PERCENT: '%';
WS: [ \t\n]+ -> skip;
