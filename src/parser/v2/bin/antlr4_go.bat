@echo off

if NOT "%~dp0" == "%CD%\" (
	echo "Please start script from directly inside the directory"
	PAUSE
	EXIT
)

set CLASSPATH=;%CD%\antlr-4.7.1-complete.jar%CLASSPATH%
java org.antlr.v4.Tool -package imagemonkeyquerylang -Dlanguage=Go ..\grammar\ImagemonkeyQueryLang.g4
copy "..\grammar\imagemonkeyquerylang_base_listener.go" ..\
copy "..\grammar\imagemonkeyquerylang_lexer.go" ..\
copy "..\grammar\imagemonkeyquerylang_listener.go" ..\
copy "..\grammar\imagemonkeyquerylang_parser.go" ..\
