// Code generated from ..\grammar\ImagemonkeyQueryLang.g4 by ANTLR 4.7.1. DO NOT EDIT.

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
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 18, 45, 4,
	2, 9, 2, 4, 3, 9, 3, 3, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 5, 3, 32, 10, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 7, 3, 40, 10, 3, 12, 3, 14, 3, 43, 11, 3, 3, 3, 2, 3, 4, 4, 2,
	4, 2, 2, 2, 51, 2, 6, 3, 2, 2, 2, 4, 31, 3, 2, 2, 2, 6, 7, 5, 4, 3, 2,
	7, 8, 7, 2, 2, 3, 8, 3, 3, 2, 2, 2, 9, 10, 8, 3, 1, 2, 10, 11, 7, 16, 2,
	2, 11, 12, 5, 4, 3, 2, 12, 13, 7, 17, 2, 2, 13, 32, 3, 2, 2, 2, 14, 15,
	7, 15, 2, 2, 15, 32, 5, 4, 3, 11, 16, 17, 7, 3, 2, 2, 17, 18, 7, 8, 2,
	2, 18, 19, 7, 12, 2, 2, 19, 32, 7, 6, 2, 2, 20, 21, 7, 5, 2, 2, 21, 22,
	7, 8, 2, 2, 22, 23, 7, 12, 2, 2, 23, 32, 7, 7, 2, 2, 24, 25, 7, 4, 2, 2,
	25, 26, 7, 8, 2, 2, 26, 27, 7, 12, 2, 2, 27, 32, 7, 7, 2, 2, 28, 32, 7,
	9, 2, 2, 29, 32, 7, 10, 2, 2, 30, 32, 7, 11, 2, 2, 31, 9, 3, 2, 2, 2, 31,
	14, 3, 2, 2, 2, 31, 16, 3, 2, 2, 2, 31, 20, 3, 2, 2, 2, 31, 24, 3, 2, 2,
	2, 31, 28, 3, 2, 2, 2, 31, 29, 3, 2, 2, 2, 31, 30, 3, 2, 2, 2, 32, 41,
	3, 2, 2, 2, 33, 34, 12, 10, 2, 2, 34, 35, 7, 13, 2, 2, 35, 40, 5, 4, 3,
	11, 36, 37, 12, 9, 2, 2, 37, 38, 7, 14, 2, 2, 38, 40, 5, 4, 3, 10, 39,
	33, 3, 2, 2, 2, 39, 36, 3, 2, 2, 2, 40, 43, 3, 2, 2, 2, 41, 39, 3, 2, 2,
	2, 41, 42, 3, 2, 2, 2, 42, 5, 3, 2, 2, 2, 43, 41, 3, 2, 2, 2, 5, 31, 39,
	41,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'annotation.coverage'", "'image.width'", "'image.height'", "'%'",
	"'px'", "", "", "", "", "", "'&'", "'|'", "'~'", "'('", "')'",
}
var symbolicNames = []string{
	"", "ANNOTATION_COVERAGE_PREFIX", "IMAGE_WIDTH_PREFIX", "IMAGE_HEIGHT_PREFIX",
	"PERCENT", "PIXEL", "OPERATOR", "ASSIGNMENT", "LABEL", "UUID", "VAL", "AND",
	"OR", "NOT", "LPAR", "RPAR", "SKIPPED_TOKENS",
}

var ruleNames = []string{
	"expression", "exp",
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
	ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX = 1
	ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX         = 2
	ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX        = 3
	ImagemonkeyQueryLangParserPERCENT                    = 4
	ImagemonkeyQueryLangParserPIXEL                      = 5
	ImagemonkeyQueryLangParserOPERATOR                   = 6
	ImagemonkeyQueryLangParserASSIGNMENT                 = 7
	ImagemonkeyQueryLangParserLABEL                      = 8
	ImagemonkeyQueryLangParserUUID                       = 9
	ImagemonkeyQueryLangParserVAL                        = 10
	ImagemonkeyQueryLangParserAND                        = 11
	ImagemonkeyQueryLangParserOR                         = 12
	ImagemonkeyQueryLangParserNOT                        = 13
	ImagemonkeyQueryLangParserLPAR                       = 14
	ImagemonkeyQueryLangParserRPAR                       = 15
	ImagemonkeyQueryLangParserSKIPPED_TOKENS             = 16
)

// ImagemonkeyQueryLangParser rules.
const (
	ImagemonkeyQueryLangParserRULE_expression = 0
	ImagemonkeyQueryLangParserRULE_exp        = 1
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
		p.SetState(4)
		p.exp(0)
	}
	{
		p.SetState(5)
		p.Match(ImagemonkeyQueryLangParserEOF)
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
	p.SetState(29)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case ImagemonkeyQueryLangParserLPAR:
		localctx = NewParenthesesExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(8)
			p.Match(ImagemonkeyQueryLangParserLPAR)
		}
		{
			p.SetState(9)
			p.exp(0)
		}
		{
			p.SetState(10)
			p.Match(ImagemonkeyQueryLangParserRPAR)
		}

	case ImagemonkeyQueryLangParserNOT:
		localctx = NewNotExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(12)
			p.Match(ImagemonkeyQueryLangParserNOT)
		}
		{
			p.SetState(13)
			p.exp(9)
		}

	case ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX:
		localctx = NewAnnotationCoverageExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(14)
			p.Match(ImagemonkeyQueryLangParserANNOTATION_COVERAGE_PREFIX)
		}
		{
			p.SetState(15)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(16)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(17)
			p.Match(ImagemonkeyQueryLangParserPERCENT)
		}

	case ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX:
		localctx = NewImageHeightExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(18)
			p.Match(ImagemonkeyQueryLangParserIMAGE_HEIGHT_PREFIX)
		}
		{
			p.SetState(19)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(20)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(21)
			p.Match(ImagemonkeyQueryLangParserPIXEL)
		}

	case ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX:
		localctx = NewImageWidthExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(22)
			p.Match(ImagemonkeyQueryLangParserIMAGE_WIDTH_PREFIX)
		}
		{
			p.SetState(23)
			p.Match(ImagemonkeyQueryLangParserOPERATOR)
		}
		{
			p.SetState(24)
			p.Match(ImagemonkeyQueryLangParserVAL)
		}
		{
			p.SetState(25)
			p.Match(ImagemonkeyQueryLangParserPIXEL)
		}

	case ImagemonkeyQueryLangParserASSIGNMENT:
		localctx = NewAssignmentExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(26)
			p.Match(ImagemonkeyQueryLangParserASSIGNMENT)
		}

	case ImagemonkeyQueryLangParserLABEL:
		localctx = NewLabelExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(27)
			p.Match(ImagemonkeyQueryLangParserLABEL)
		}

	case ImagemonkeyQueryLangParserUUID:
		localctx = NewUuidExpressionContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(28)
			p.Match(ImagemonkeyQueryLangParserUUID)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(39)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(37)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewAndExpressionContext(p, NewExpContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ImagemonkeyQueryLangParserRULE_exp)
				p.SetState(31)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
				}
				{
					p.SetState(32)
					p.Match(ImagemonkeyQueryLangParserAND)
				}
				{
					p.SetState(33)
					p.exp(9)
				}

			case 2:
				localctx = NewOrExpressionContext(p, NewExpContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ImagemonkeyQueryLangParserRULE_exp)
				p.SetState(34)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
				}
				{
					p.SetState(35)
					p.Match(ImagemonkeyQueryLangParserOR)
				}
				{
					p.SetState(36)
					p.exp(8)
				}

			}

		}
		p.SetState(41)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
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
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 7)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
