#!/bin/bash

export CLASSPATH=antlr-4.7.1-complete.jar:$CLASSPATH

java org.antlr.v4.Tool -package imagemonkeyquerylang -Dlanguage=Go ../grammar/ImagemonkeyQueryLang.g4

cp ../grammar/imagemonkeyquerylang_base_listener.go ../
cp ../grammar/imagemonkeyquerylang_lexer.go ../
cp ../grammar/imagemonkeyquerylang_listener.go ../
cp ../grammar/imagemonkeyquerylang_parser.go ../
