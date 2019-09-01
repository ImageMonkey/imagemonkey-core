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
	allowImageWidth bool
	allowImageHeight bool
	allowAnnotationCoverage bool
	version int
	allowOrderByValidation bool
	allowImageCollection bool
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
        allowImageWidth: true,
		allowImageHeight: true,
		allowAnnotationCoverage: true,
        version: 2,
        allowOrderByValidation: false,
		allowImageCollection: false,
    } 
}

func (p *QueryParser) AllowImageHeight(allow bool) {
    p.allowImageHeight = allow
}

func (p *QueryParser) AllowImageWidth(allow bool) {
	p.allowImageWidth = allow
}

func (p *QueryParser) AllowAnnotationCoverage(allow bool) {
	p.allowAnnotationCoverage = allow
}
func (p *QueryParser) AllowImageCollection(allow bool) {
	p.allowImageCollection = allow
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
		allowImageWidth: p.allowImageWidth,
		allowImageHeight: p.allowImageHeight,
		allowAnnotationCoverage: p.allowAnnotationCoverage,
		allowOrderByValidation: p.allowOrderByValidation,
		allowImageCollection: p.allowImageCollection,
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
