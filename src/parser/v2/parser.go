package imagemonkeyquerylang

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type Parser interface {
	Parse() error
}

type QueryParser struct {
	query string
	offset int
	allowStaticQueryAttributes bool
}

type ParseResult struct {
	Input string
	Query string
    Subquery string
    IsUuidQuery bool
	QueryValues []interface{}
}

func NewQueryParser(query string) *QueryParser {
    return &QueryParser {
        query: query,
        offset: 1,
        allowStaticQueryAttributes: true,
    } 
}

func (p *QueryParser) AllowStaticQueryAttributes(allow bool) {
    p.allowStaticQueryAttributes = allow
}

func (p *QueryParser) SetOffset(offset int) {
	p.offset = offset
}

func (p *QueryParser) Parse() (ParseResult, error) {
	is := antlr.NewInputStream(p.query)

	lexer := NewImagemonkeyQueryLangLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	listener := imagemonkeyQueryLangListener{
		pos: p.offset,
		allowStaticQueryAttributes: p.allowStaticQueryAttributes,
		numOfLabels: 0,
	}
	errorListener := NewCustomErrorListener() 
	parser := NewImagemonkeyQueryLangParser(stream)
	parser.RemoveErrorListeners() //remove default error listeners
	parser.AddErrorListener(errorListener)
	antlr.ParseTreeWalkerDefault.Walk(&listener, parser.Expression())

	return listener.pop(), errorListener.Error
}