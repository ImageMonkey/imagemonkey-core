// Code generated from ../grammar/ImagemonkeyQueryLang.g4 by ANTLR 4.7.1. DO NOT EDIT.

package imagemonkeyquerylang // ImagemonkeyQueryLang
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseImagemonkeyQueryLangListener is a complete listener for a parse tree produced by ImagemonkeyQueryLangParser.
type BaseImagemonkeyQueryLangListener struct{}

var _ ImagemonkeyQueryLangListener = &BaseImagemonkeyQueryLangListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseImagemonkeyQueryLangListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseImagemonkeyQueryLangListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitExpression(ctx *ExpressionContext) {}

// EnterImageHeightExpression is called when production imageHeightExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterImageHeightExpression(ctx *ImageHeightExpressionContext) {
}

// ExitImageHeightExpression is called when production imageHeightExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitImageHeightExpression(ctx *ImageHeightExpressionContext) {
}

// EnterOrExpression is called when production orExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterOrExpression(ctx *OrExpressionContext) {}

// ExitOrExpression is called when production orExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitOrExpression(ctx *OrExpressionContext) {}

// EnterParenthesesExpression is called when production parenthesesExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterParenthesesExpression(ctx *ParenthesesExpressionContext) {
}

// ExitParenthesesExpression is called when production parenthesesExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitParenthesesExpression(ctx *ParenthesesExpressionContext) {
}

// EnterImageNumLabelsExpression is called when production imageNumLabelsExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterImageNumLabelsExpression(ctx *ImageNumLabelsExpressionContext) {
}

// ExitImageNumLabelsExpression is called when production imageNumLabelsExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitImageNumLabelsExpression(ctx *ImageNumLabelsExpressionContext) {
}

// EnterAndExpression is called when production andExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterAndExpression(ctx *AndExpressionContext) {}

// ExitAndExpression is called when production andExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitAndExpression(ctx *AndExpressionContext) {}

// EnterImageNumAnnotationsExpression is called when production imageNumAnnotationsExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterImageNumAnnotationsExpression(ctx *ImageNumAnnotationsExpressionContext) {
}

// ExitImageNumAnnotationsExpression is called when production imageNumAnnotationsExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitImageNumAnnotationsExpression(ctx *ImageNumAnnotationsExpressionContext) {
}

// EnterImageWidthExpression is called when production imageWidthExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterImageWidthExpression(ctx *ImageWidthExpressionContext) {
}

// ExitImageWidthExpression is called when production imageWidthExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitImageWidthExpression(ctx *ImageWidthExpressionContext) {
}

// EnterAssignmentExpression is called when production assignmentExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterAssignmentExpression(ctx *AssignmentExpressionContext) {
}

// ExitAssignmentExpression is called when production assignmentExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitAssignmentExpression(ctx *AssignmentExpressionContext) {
}

// EnterNotExpression is called when production notExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterNotExpression(ctx *NotExpressionContext) {}

// ExitNotExpression is called when production notExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitNotExpression(ctx *NotExpressionContext) {}

// EnterUuidExpression is called when production uuidExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterUuidExpression(ctx *UuidExpressionContext) {}

// ExitUuidExpression is called when production uuidExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitUuidExpression(ctx *UuidExpressionContext) {}

// EnterLabelExpression is called when production labelExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterLabelExpression(ctx *LabelExpressionContext) {}

// ExitLabelExpression is called when production labelExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitLabelExpression(ctx *LabelExpressionContext) {}

// EnterAnnotationCoverageExpression is called when production annotationCoverageExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterAnnotationCoverageExpression(ctx *AnnotationCoverageExpressionContext) {
}

// ExitAnnotationCoverageExpression is called when production annotationCoverageExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitAnnotationCoverageExpression(ctx *AnnotationCoverageExpressionContext) {
}

// EnterOrderByValidationDescExpression is called when production orderByValidationDescExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterOrderByValidationDescExpression(ctx *OrderByValidationDescExpressionContext) {
}

// ExitOrderByValidationDescExpression is called when production orderByValidationDescExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitOrderByValidationDescExpression(ctx *OrderByValidationDescExpressionContext) {
}

// EnterOrderByValidationAscExpression is called when production orderByValidationAscExpression is entered.
func (s *BaseImagemonkeyQueryLangListener) EnterOrderByValidationAscExpression(ctx *OrderByValidationAscExpressionContext) {
}

// ExitOrderByValidationAscExpression is called when production orderByValidationAscExpression is exited.
func (s *BaseImagemonkeyQueryLangListener) ExitOrderByValidationAscExpression(ctx *OrderByValidationAscExpressionContext) {
}
