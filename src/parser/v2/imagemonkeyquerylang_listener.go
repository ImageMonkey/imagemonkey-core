// Code generated from ../grammar/ImagemonkeyQueryLang.g4 by ANTLR 4.7.1. DO NOT EDIT.

package imagemonkeyquerylang // ImagemonkeyQueryLang
import "github.com/antlr/antlr4/runtime/Go/antlr"

// ImagemonkeyQueryLangListener is a complete listener for a parse tree produced by ImagemonkeyQueryLangParser.
type ImagemonkeyQueryLangListener interface {
	antlr.ParseTreeListener

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterImageHeightExpression is called when entering the imageHeightExpression production.
	EnterImageHeightExpression(c *ImageHeightExpressionContext)

	// EnterOrExpression is called when entering the orExpression production.
	EnterOrExpression(c *OrExpressionContext)

	// EnterParenthesesExpression is called when entering the parenthesesExpression production.
	EnterParenthesesExpression(c *ParenthesesExpressionContext)

	// EnterAndExpression is called when entering the andExpression production.
	EnterAndExpression(c *AndExpressionContext)

	// EnterImageWidthExpression is called when entering the imageWidthExpression production.
	EnterImageWidthExpression(c *ImageWidthExpressionContext)

	// EnterAssignmentExpression is called when entering the assignmentExpression production.
	EnterAssignmentExpression(c *AssignmentExpressionContext)

	// EnterNotExpression is called when entering the notExpression production.
	EnterNotExpression(c *NotExpressionContext)

	// EnterUuidExpression is called when entering the uuidExpression production.
	EnterUuidExpression(c *UuidExpressionContext)

	// EnterLabelExpression is called when entering the labelExpression production.
	EnterLabelExpression(c *LabelExpressionContext)

	// EnterAnnotationCoverageExpression is called when entering the annotationCoverageExpression production.
	EnterAnnotationCoverageExpression(c *AnnotationCoverageExpressionContext)

	// EnterOrderByValidationDescExpression is called when entering the orderByValidationDescExpression production.
	EnterOrderByValidationDescExpression(c *OrderByValidationDescExpressionContext)

	// EnterOrderByValidationAscExpression is called when entering the orderByValidationAscExpression production.
	EnterOrderByValidationAscExpression(c *OrderByValidationAscExpressionContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitImageHeightExpression is called when exiting the imageHeightExpression production.
	ExitImageHeightExpression(c *ImageHeightExpressionContext)

	// ExitOrExpression is called when exiting the orExpression production.
	ExitOrExpression(c *OrExpressionContext)

	// ExitParenthesesExpression is called when exiting the parenthesesExpression production.
	ExitParenthesesExpression(c *ParenthesesExpressionContext)

	// ExitAndExpression is called when exiting the andExpression production.
	ExitAndExpression(c *AndExpressionContext)

	// ExitImageWidthExpression is called when exiting the imageWidthExpression production.
	ExitImageWidthExpression(c *ImageWidthExpressionContext)

	// ExitAssignmentExpression is called when exiting the assignmentExpression production.
	ExitAssignmentExpression(c *AssignmentExpressionContext)

	// ExitNotExpression is called when exiting the notExpression production.
	ExitNotExpression(c *NotExpressionContext)

	// ExitUuidExpression is called when exiting the uuidExpression production.
	ExitUuidExpression(c *UuidExpressionContext)

	// ExitLabelExpression is called when exiting the labelExpression production.
	ExitLabelExpression(c *LabelExpressionContext)

	// ExitAnnotationCoverageExpression is called when exiting the annotationCoverageExpression production.
	ExitAnnotationCoverageExpression(c *AnnotationCoverageExpressionContext)

	// ExitOrderByValidationDescExpression is called when exiting the orderByValidationDescExpression production.
	ExitOrderByValidationDescExpression(c *OrderByValidationDescExpressionContext)

	// ExitOrderByValidationAscExpression is called when exiting the orderByValidationAscExpression production.
	ExitOrderByValidationAscExpression(c *OrderByValidationAscExpressionContext)
}
