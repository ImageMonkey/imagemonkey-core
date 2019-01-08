grammar ImagemonkeyQueryLang;

/*
 * Parser Rules
 */

expression:
    exp EOF;

exp 
    :   LPAR exp RPAR            							# parenthesesExpression
    |   NOT exp                  							# notExpression
    |   exp AND exp       							# andExpression
    |   exp OR exp        							# orExpression
    |   ANNOTATION_COVERAGE_PREFIX OPERATOR VAL PERCENT             # annotationCoverageExpression
    |   IMAGE_HEIGHT_PREFIX OPERATOR VAL PIXEL                      # imageHeightExpression
    |   IMAGE_WIDTH_PREFIX OPERATOR VAL PIXEL                       # imageWidthExpression
    |   ASSIGNMENT                                                  # assignmentExpression
    |   LABEL                                                       # labelExpression
    |   UUID                                                        # uuidExpression
    ;



/*
 * Lexer Rules
 */
fragment LOWERCASE  		: [a-z];
fragment UPPERCASE  		: [A-Z];
fragment UPPERLOWERCASE     : [a-zA-Z];
fragment UPPERLOWERCASEWS   : [a-zA-Z ];
fragment UUIDBLOCK  		: [A-Za-z0-9] [A-Za-z0-9] [A-Za-z0-9] [A-Za-z0-9];
fragment WS                 : ' ';
ANNOTATION_COVERAGE_PREFIX	: 'annotation.coverage';
IMAGE_WIDTH_PREFIX			: 'image.width';
IMAGE_HEIGHT_PREFIX			: 'image.height';
PERCENT						: '%';
PIXEL						: 'px';
OPERATOR					: '>' | '<' | '>=' | '=' | '<=';
ASSIGNMENT 					: (UPPERLOWERCASE)+ ('.' UPPERLOWERCASE+)? WS* '=' WS* '\'' (UPPERLOWERCASEWS)+ '\'';
LABEL               		: UPPERLOWERCASE | (UPPERLOWERCASE (WS | UPPERLOWERCASE)* UPPERLOWERCASE);
UUID 						: UUIDBLOCK UUIDBLOCK '-' UUIDBLOCK '-' UUIDBLOCK '-' UUIDBLOCK '-' UUIDBLOCK UUIDBLOCK UUIDBLOCK;
VAL							: [0-9]+;
AND							: '&';
OR							: '|';
NOT							: '~'; 	
LPAR						: '(';
RPAR						: ')';

SKIPPED_TOKENS              : [ \t\r\n]+ -> skip ; // skip spaces, tabs, newlines