// Code generated from ../grammar/ImagemonkeyQueryLang.g4 by ANTLR 4.7.1. DO NOT EDIT.

package imagemonkeyquerylang // ImagemonkeyQueryLang
import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 24, 60, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 5,
	2, 15, 10, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 5, 3, 42, 10, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 7, 3, 50, 10, 3, 12, 3, 14, 3, 53, 11, 3, 3, 4, 3, 4, 3, 4, 5, 4, 58,
	10, 4, 3, 4, 2, 3, 4, 5, 2, 4, 6, 2, 2, 2, 69, 2, 8, 3, 2, 2, 2, 4, 41,
	3, 2, 2, 2, 6, 57, 3, 2, 2, 2, 8, 14, 5, 4, 3, 2, 9, 10, 7, 3, 2, 2, 10,
	11, 5, 6, 4, 2, 11, 12, 7, 2, 2, 3, 12, 15, 3, 2, 2, 2, 13, 15, 7, 2, 2,
	3, 14, 9, 3, 2, 2, 2, 14, 13, 3, 2, 2, 2, 15, 3, 3, 2, 2, 2, 16, 17, 8,
	3, 1, 2, 17, 18, 7, 22, 2, 2, 18, 19, 5, 4, 3, 2, 19, 20, 7, 23, 2, 2,
	20, 42, 3, 2, 2, 2, 21, 22, 7, 21, 2, 2, 22, 42, 5, 4, 3, 12, 23, 24, 7,
	4, 2, 2, 24, 25, 7, 10, 2, 2, 25, 26, 7, 18, 2, 2, 26, 42, 7, 8, 2, 2,
	27, 28, 7, 6, 2, 2, 28, 29, 7, 10, 2, 2, 29, 30, 7, 18, 2, 2, 30, 42, 7,
	9, 2, 2, 31, 32, 7, 5, 2, 2, 32, 33, 7, 10, 2, 2, 33, 34, 7, 18, 2, 2,
	34, 42, 7, 9, 2, 2, 35, 36, 7, 7, 2, 2, 36, 37, 7, 10, 2, 2, 37, 42, 7,
	18, 2, 2, 38, 42, 7, 11, 2, 2, 39, 42, 7, 16, 2, 2, 40, 42, 7, 17, 2, 2,
	41, 16, 3, 2, 2, 2, 41, 21, 3, 2, 2, 2, 41, 23, 3, 2, 2, 2, 41, 27, 3,
	2, 2, 2, 41, 31, 3, 2, 2, 2, 41, 35, 3, 2, 2, 2, 41, 38, 3, 2, 2, 2, 41,
	39, 3, 2, 2, 2, 41, 40, 3, 2, 2, 2, 42, 51, 3, 2, 2, 2, 43, 44, 12, 11,
	2, 2, 44, 45, 7, 19, 2, 2, 45, 50, 5, 4, 3, 12, 46, 47, 12, 10, 2, 2, 47,
	48, 7, 20, 2, 2, 48, 50, 5, 4, 3, 11, 49, 43, 3, 2, 2, 2, 49, 46, 3, 2,
	2, 2, 50, 53, 3, 2, 2, 2, 51, 49, 3, 2, 2, 2, 51, 52, 3, 2, 2, 2, 52, 5,
	3, 2, 2, 2, 53, 51, 3, 2, 2, 2, 54, 58, 7, 13, 2, 2, 55, 58, 7, 14, 2,
	2, 56, 58, 7, 15, 2, 2, 57, 54, 3, 2, 2, 2, 57, 55, 3, 2, 2, 2, 57, 56,
	3, 2, 2, 2, 58, 7, 3, 2, 2, 2, 7, 14, 41, 49, 51, 57,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'!'", "'annotation.coverage'", "'image.width'", "'image.height'",
	"'image.num_labels'", "'%'", "'px'", "", "", "", "", "", "", "", "", "",
	"'&'", "'|'", "'~'", "'('", "')'",
}
var symbolicNames = []string{
	"", "SEP", "ANNOTATION_COVERAGE_PREFIX", "IMAGE_WIDTH_PREFIX", "IMAGE_HEIGHT_PREFIX",
	"IMAGE_NUM_LABELS_PREFIX", "PERCENT", "PIXEL", "OPERATOR", "ASSIGNMENT",
	"ORDER_BY", "ORDER_BY_VALIDATION_DESC", "ORDER_BY_VALIDATION_ASC", "ORDER_BY_VALIDATION",
	"LABEL", "UUID", "VAL", "AND", "OR", "NOT", "LPAR", "RPAR", "SKIPPED_TOKENS",
}

var ruleNames = []string{
	"expression", "exp", "order_by",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type ImagemonkeyQueryLangParser struct {
	*antlr.BaseParser
}

func NewImagemonkeyQueryLangParser(input antlr.TokenStream) *ImagemonkeyQueryLangParser {
	this := new(ImagemonkeyQueryLangParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "ImagemonkeyQueryLang.g4"

	return this
}

// ImagemonkeyQueryLangParser tokens.
const (
	ImagemonkeyQueryLangParserEOF                        = antlr.TokenEOF
	ImagemonkeyQueryLangParserSEP                        = 1
	ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX = 2
	ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX         = 3
	ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX        = 4
	ImagemonkeyQueryLangParserIMAGE_NUM_LABELS_PREFIX    = 5
	ImagemonkeyQueryLangParserPERCENT                    = 6
	ImagemonkeyQueryLangParserPIXEL                      = 7
	ImagemonkeyQueryLangParserOPERATOR                   = 8
	ImagemonkeyQueryLangParserASSIGNMENT                 = 9
	ImagemonkeyQueryLangParserORDER_BY                   = 10
	ImagemonkeyQueryLangParserORDER_BY_VALIDATION_DESC   = 11
	ImagemonkeyQueryLangParserORDER_BY_VALIDATION_ASC    = 12
	ImagemonkeyQueryLangParserORDER_BY_VALIDATION        = 13
	ImagemonkeyQueryLangParserLABEL                      = 14
	ImagemonkeyQueryLangParserUUID                       = 15
	ImagemonkeyQueryLangParserVAL                        = 16
	ImagemonkeyQueryLangParserAND                        = 17
	ImagemonkeyQueryLangParserOR                         = 18
	ImagemonkeyQueryLangParserNOT                        = 19
	ImagemonkeyQueryLangParserLPAR                       = 20
	ImagemonkeyQueryLangParserRPAR                       = 21
	ImagemonkeyQueryLangParserSKIPPED_TOKENS             = 22
)

// ImagemonkeyQueryLangParser rules.
const (
	ImagemonkeyQueryLangParserRULE_expression = 0
	ImagemonkeyQueryLangParserRULE_exp        = 1
	ImagemonkeyQueryLangParserRULE_order_by   = 2
)

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_expression
	return p
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) Exp() IExpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpContext)
}

func (s *ExpressionContext) SEP() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserSEP, 0)
}

func (s *ExpressionContext) Order_by() IOrder_byContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOrder_byContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOrder_byContext)
}

func (s *ExpressionContext) EOF() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserEOF, 0)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (p *ImagemonkeyQueryLangParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ImagemonkeyQueryLangParserRULE_expression)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(6)
		p.exp(0)
	}
	p.SetState(12)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case ImagemonkeyQueryLangParserSEP:
		{
			p.SetState(7)
			p.Match(ImagemonkeyQueryLangParserSEP)
		}
		{
			p.SetState(8)
			p.Order_by()
		}
		{
			p.SetState(9)
			p.Match(ImagemonkeyQueryLangParserEOF)
		}

	case ImagemonkeyQueryLangParserEOF:
		{
			p.SetState(11)
			p.Match(ImagemonkeyQueryLangParserEOF)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IExpContext is an interface to support dynamic dispatch.
type IExpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExpContext differentiates from other interfaces.
	IsExpContext()
}

type ExpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpContext() *ExpContext {
	var p = new(ExpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_exp
	return p
}

func (*ExpContext) IsExpContext() {}

func NewExpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpContext {
	var p = new(ExpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_exp

	return p
}

func (s *ExpContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpContext) CopyFrom(ctx *ExpContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ExpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ImageHeightExpressionContext struct {
	*ExpContext
}

func NewImageHeightExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ImageHeightExpressionContext {
	var p = new(ImageHeightExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *ImageHeightExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ImageHeightExpressionContext) IMAGE_HEIGHT_PREFIX() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX, 0)
}

func (s *ImageHeightExpressionContext) OPERATOR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserOPERATOR, 0)
}

func (s *ImageHeightExpressionContext) VAL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserVAL, 0)
}

func (s *ImageHeightExpressionContext) PIXEL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserPIXEL, 0)
}

func (s *ImageHeightExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterImageHeightExpression(s)
	}
}

func (s *ImageHeightExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitImageHeightExpression(s)
	}
}

type OrExpressionContext struct {
	*ExpContext
}

func NewOrExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OrExpressionContext {
	var p = new(OrExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *OrExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrExpressionContext) AllExp() []IExpContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpContext)(nil)).Elem())
	var tst = make([]IExpContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpContext)
		}
	}

	return tst
}

func (s *OrExpressionContext) Exp(i int) IExpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpContext)
}

func (s *OrExpressionContext) OR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserOR, 0)
}

func (s *OrExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterOrExpression(s)
	}
}

func (s *OrExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitOrExpression(s)
	}
}

type ParenthesesExpressionContext struct {
	*ExpContext
}

func NewParenthesesExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenthesesExpressionContext {
	var p = new(ParenthesesExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *ParenthesesExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenthesesExpressionContext) LPAR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserLPAR, 0)
}

func (s *ParenthesesExpressionContext) Exp() IExpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpContext)
}

func (s *ParenthesesExpressionContext) RPAR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserRPAR, 0)
}

func (s *ParenthesesExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterParenthesesExpression(s)
	}
}

func (s *ParenthesesExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitParenthesesExpression(s)
	}
}

type ImageNumLabelsExpressionContext struct {
	*ExpContext
}

func NewImageNumLabelsExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ImageNumLabelsExpressionContext {
	var p = new(ImageNumLabelsExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *ImageNumLabelsExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ImageNumLabelsExpressionContext) IMAGE_NUM_LABELS_PREFIX() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserIMAGE_NUM_LABELS_PREFIX, 0)
}

func (s *ImageNumLabelsExpressionContext) OPERATOR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserOPERATOR, 0)
}

func (s *ImageNumLabelsExpressionContext) VAL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserVAL, 0)
}

func (s *ImageNumLabelsExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterImageNumLabelsExpression(s)
	}
}

func (s *ImageNumLabelsExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitImageNumLabelsExpression(s)
	}
}

type AndExpressionContext struct {
	*ExpContext
}

func NewAndExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AndExpressionContext {
	var p = new(AndExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *AndExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndExpressionContext) AllExp() []IExpContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpContext)(nil)).Elem())
	var tst = make([]IExpContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpContext)
		}
	}

	return tst
}

func (s *AndExpressionContext) Exp(i int) IExpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpContext)
}

func (s *AndExpressionContext) AND() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserAND, 0)
}

func (s *AndExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterAndExpression(s)
	}
}

func (s *AndExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitAndExpression(s)
	}
}

type ImageWidthExpressionContext struct {
	*ExpContext
}

func NewImageWidthExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ImageWidthExpressionContext {
	var p = new(ImageWidthExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *ImageWidthExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ImageWidthExpressionContext) IMAGE_WIDTH_PREFIX() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX, 0)
}

func (s *ImageWidthExpressionContext) OPERATOR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserOPERATOR, 0)
}

func (s *ImageWidthExpressionContext) VAL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserVAL, 0)
}

func (s *ImageWidthExpressionContext) PIXEL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserPIXEL, 0)
}

func (s *ImageWidthExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterImageWidthExpression(s)
	}
}

func (s *ImageWidthExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitImageWidthExpression(s)
	}
}

type AssignmentExpressionContext struct {
	*ExpContext
}

func NewAssignmentExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AssignmentExpressionContext {
	var p = new(AssignmentExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *AssignmentExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AssignmentExpressionContext) ASSIGNMENT() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserASSIGNMENT, 0)
}

func (s *AssignmentExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterAssignmentExpression(s)
	}
}

func (s *AssignmentExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitAssignmentExpression(s)
	}
}

type NotExpressionContext struct {
	*ExpContext
}

func NewNotExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NotExpressionContext {
	var p = new(NotExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *NotExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NotExpressionContext) NOT() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserNOT, 0)
}

func (s *NotExpressionContext) Exp() IExpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpContext)
}

func (s *NotExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterNotExpression(s)
	}
}

func (s *NotExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitNotExpression(s)
	}
}

type UuidExpressionContext struct {
	*ExpContext
}

func NewUuidExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *UuidExpressionContext {
	var p = new(UuidExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *UuidExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UuidExpressionContext) UUID() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserUUID, 0)
}

func (s *UuidExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterUuidExpression(s)
	}
}

func (s *UuidExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitUuidExpression(s)
	}
}

type LabelExpressionContext struct {
	*ExpContext
}

func NewLabelExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LabelExpressionContext {
	var p = new(LabelExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *LabelExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LabelExpressionContext) LABEL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserLABEL, 0)
}

func (s *LabelExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterLabelExpression(s)
	}
}

func (s *LabelExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitLabelExpression(s)
	}
}

type AnnotationCoverageExpressionContext struct {
	*ExpContext
}

func NewAnnotationCoverageExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AnnotationCoverageExpressionContext {
	var p = new(AnnotationCoverageExpressionContext)

	p.ExpContext = NewEmptyExpContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpContext))

	return p
}

func (s *AnnotationCoverageExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AnnotationCoverageExpressionContext) ANNOTATION_COVERAGE_PREFIX() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX, 0)
}

func (s *AnnotationCoverageExpressionContext) OPERATOR() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserOPERATOR, 0)
}

func (s *AnnotationCoverageExpressionContext) VAL() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserVAL, 0)
}

func (s *AnnotationCoverageExpressionContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserPERCENT, 0)
}

func (s *AnnotationCoverageExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterAnnotationCoverageExpression(s)
	}
}

func (s *AnnotationCoverageExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitAnnotationCoverageExpression(s)
	}
}

func (p *ImagemonkeyQueryLangParser) Exp() (localctx IExpContext) {
	return p.exp(0)
}

func (p *ImagemonkeyQueryLangParser) exp(_p int) (localctx IExpContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExpContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 2
	p.EnterRecursionRule(localctx, 2, ImagemonkeyQueryLangParserRULE_exp, _p)

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(39)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case ImagemonkeyQueryLangParserLPAR:
		localctx = NewParenthesesExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(15)
			p.Match(ImagemonkeyQueryLangParserLPAR)
		}
		{
			p.SetState(16)
			p.exp(0)
		}
		{
			p.SetState(17)
			p.Match(ImagemonkeyQueryLangParserRPAR)
		}

	case ImagemonkeyQueryLangParserNOT:
		localctx = NewNotExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(19)
			p.Match(ImagemonkeyQueryLangParserNOT)
		}
		{
			p.SetState(20)
			p.exp(10)
		}

	case ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX:
		localctx = NewAnnotationCoverageExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(21)
			p.Match(ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX)
		}
		{
			p.SetState(22)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(23)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(24)
			p.Match(ImagemonkeyQueryLangParserPERCENT)
		}

	case ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX:
		localctx = NewImageHeightExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(25)
			p.Match(ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX)
		}
		{
			p.SetState(26)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(27)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(28)
			p.Match(ImagemonkeyQueryLangParserPIXEL)
		}

	case ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX:
		localctx = NewImageWidthExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(29)
			p.Match(ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX)
		}
		{
			p.SetState(30)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(31)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(32)
			p.Match(ImagemonkeyQueryLangParserPIXEL)
		}

	case ImagemonkeyQueryLangParserIMAGE_NUM_LABELS_PREFIX:
		localctx = NewImageNumLabelsExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(33)
			p.Match(ImagemonkeyQueryLangParserIMAGE_NUM_LABELS_PREFIX)
		}
		{
			p.SetState(34)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(35)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}

	case ImagemonkeyQueryLangParserASSIGNMENT:
		localctx = NewAssignmentExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(36)
			p.Match(ImagemonkeyQueryLangParserASSIGNMENT)
		}

	case ImagemonkeyQueryLangParserLABEL:
		localctx = NewLabelExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(37)
			p.Match(ImagemonkeyQueryLangParserLABEL)
		}

	case ImagemonkeyQueryLangParserUUID:
		localctx = NewUuidExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(38)
			p.Match(ImagemonkeyQueryLangParserUUID)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(47)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext()) {
			case 1:
				localctx = NewAndExpressionContext(p, NewExpContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ImagemonkeyQueryLangParserRULE_exp)
				p.SetState(41)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
				}
				{
					p.SetState(42)
					p.Match(ImagemonkeyQueryLangParserAND)
				}
				{
					p.SetState(43)
					p.exp(10)
				}

			case 2:
				localctx = NewOrExpressionContext(p, NewExpContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ImagemonkeyQueryLangParserRULE_exp)
				p.SetState(44)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
				}
				{
					p.SetState(45)
					p.Match(ImagemonkeyQueryLangParserOR)
				}
				{
					p.SetState(46)
					p.exp(9)
				}

			}

		}
		p.SetState(51)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())
	}

	return localctx
}

// IOrder_byContext is an interface to support dynamic dispatch.
type IOrder_byContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsOrder_byContext differentiates from other interfaces.
	IsOrder_byContext()
}

type Order_byContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrder_byContext() *Order_byContext {
	var p = new(Order_byContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_order_by
	return p
}

func (*Order_byContext) IsOrder_byContext() {}

func NewOrder_byContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Order_byContext {
	var p = new(Order_byContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = ImagemonkeyQueryLangParserRULE_order_by

	return p
}

func (s *Order_byContext) GetParser() antlr.Parser { return s.parser }

func (s *Order_byContext) CopyFrom(ctx *Order_byContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *Order_byContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Order_byContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type OrderByValidationAscExpressionContext struct {
	*Order_byContext
}

func NewOrderByValidationAscExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OrderByValidationAscExpressionContext {
	var p = new(OrderByValidationAscExpressionContext)

	p.Order_byContext = NewEmptyOrder_byContext()
	p.parser = parser
	p.CopyFrom(ctx.(*Order_byContext))

	return p
}

func (s *OrderByValidationAscExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderByValidationAscExpressionContext) ORDER_BY_VALIDATION_ASC() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserORDER_BY_VALIDATION_ASC, 0)
}

func (s *OrderByValidationAscExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterOrderByValidationAscExpression(s)
	}
}

func (s *OrderByValidationAscExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitOrderByValidationAscExpression(s)
	}
}

type OrderByValidationDescExpressionContext struct {
	*Order_byContext
}

func NewOrderByValidationDescExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OrderByValidationDescExpressionContext {
	var p = new(OrderByValidationDescExpressionContext)

	p.Order_byContext = NewEmptyOrder_byContext()
	p.parser = parser
	p.CopyFrom(ctx.(*Order_byContext))

	return p
}

func (s *OrderByValidationDescExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderByValidationDescExpressionContext) ORDER_BY_VALIDATION_DESC() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserORDER_BY_VALIDATION_DESC, 0)
}

func (s *OrderByValidationDescExpressionContext) ORDER_BY_VALIDATION() antlr.TerminalNode {
	return s.GetToken(ImagemonkeyQueryLangParserORDER_BY_VALIDATION, 0)
}

func (s *OrderByValidationDescExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.EnterOrderByValidationDescExpression(s)
	}
}

func (s *OrderByValidationDescExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ImagemonkeyQueryLangListener); ok {
		listenerT.ExitOrderByValidationDescExpression(s)
	}
}

func (p *ImagemonkeyQueryLangParser) Order_by() (localctx IOrder_byContext) {
	localctx = NewOrder_byContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ImagemonkeyQueryLangParserRULE_order_by)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(55)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case ImagemonkeyQueryLangParserORDER_BY_VALIDATION_DESC:
		localctx = NewOrderByValidationDescExpressionContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(52)
			p.Match(ImagemonkeyQueryLangParserORDER_BY_VALIDATION_DESC)
		}

	case ImagemonkeyQueryLangParserORDER_BY_VALIDATION_ASC:
		localctx = NewOrderByValidationAscExpressionContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(53)
			p.Match(ImagemonkeyQueryLangParserORDER_BY_VALIDATION_ASC)
		}

	case ImagemonkeyQueryLangParserORDER_BY_VALIDATION:
		localctx = NewOrderByValidationDescExpressionContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(54)
			p.Match(ImagemonkeyQueryLangParserORDER_BY_VALIDATION)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

func (p *ImagemonkeyQueryLangParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 1:
		var t *ExpContext = nil
		if localctx != nil {
			t = localctx.(*ExpContext)
		}
		return p.Exp_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *ImagemonkeyQueryLangParser) Exp_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 8)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
