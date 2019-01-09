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
	version int
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
        version: 2,
    } 
}

func (p *QueryParser) AllowStaticQueryAttributes(allow bool) {
    p.allowStaticQueryAttributes = allow
}

func (p *QueryParser) SetOffset(offset int) {
	p.offset = offset
}

func (p *QueryParser) SetVersion(version int) {
	p.version = version
}

func (p *QueryParser) Parse() (ParseResult, error) {
	is := antlr.NewInputStream(p.query)

	lexer := NewImagemonkeyQueryLangLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	listener := imagemonkeyQueryLangListener{
		pos: p.offset,
		allowStaticQueryAttributes: p.allowStaticQueryAttributes,
		numOfLabels: 0,
		version: p.version,
		isUuidQuery: true,
		typeOfQueryKnown: false,
	}
	errorListener := NewCustomErrorListener() 
	errorListener.query = p.query
	parser := NewImagemonkeyQueryLangParser(stream)
	parser.RemoveErrorListeners() //remove default error listeners
	parser.AddErrorListener(errorListener)
	antlr.ParseTreeWalkerDefault.Walk(&listener, parser.Expression())

	parseResult := listener.pop()
	parseResult.IsUuidQuery = listener.isUuidQuery
	parseResult.Input = p.query

	var err error = nil
	if errorListener.err != nil {
		err = errorListener.err
	} else if listener.err != nil {
		err = listener.err
	}

	return parseResult, err
}