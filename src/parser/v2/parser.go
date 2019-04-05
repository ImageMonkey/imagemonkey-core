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
	allowOrderByValidation bool
}

type ResultOrderType int
const (
	OrderByNumOfExistingValidations ResultOrderType = 1 << iota
	OrderByDefault
)

type ResultOrderDirection int
const (
	ResultOrderAscDirection ResultOrderDirection = 1 << iota
	ResultOrderDescDirection
	ResultOrderDefaultDirection
)


type ResultOrder struct {
	Type ResultOrderType
	Direction ResultOrderDirection
}

type ParseResult struct {
	Input string
	Query string
    Subquery string
    IsUuidQuery bool
	QueryValues []interface{}
	OrderBy ResultOrder
}

func NewQueryParser(query string) *QueryParser {
    return &QueryParser {
        query: query,
        offset: 1,
        allowStaticQueryAttributes: true,
        version: 2,
        allowOrderByValidation: false,
    } 
}

func (p *QueryParser) AllowStaticQueryAttributes(allow bool) {
    p.allowStaticQueryAttributes = allow
}

func (p *QueryParser) AllowOrderByValidation(allow bool) {
	p.allowOrderByValidation = allow
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
		allowOrderByValidation: p.allowOrderByValidation,
		numOfLabels: 0,
		version: p.version,
		isUuidQuery: true,
		typeOfQueryKnown: false,
		query: p.query,
		resultOrder: ResultOrder{Direction: ResultOrderDefaultDirection, Type: OrderByDefault},
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
	parseResult.OrderBy = listener.resultOrder

	var err error = nil
	if errorListener.err != nil {
		err = errorListener.err
	} else if listener.err != nil {
		err = listener.err
	}

	return parseResult, err
}