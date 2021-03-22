package imagemonkeyquerylang

import (
	"testing"
	"reflect"
	"runtime"
	"fmt"
	"path/filepath"
)

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// notOk fails the test if an err is nil.
func notOk(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error, expected not nil, but got nil: \033[39m\n\n", filepath.Base(file), line)
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}


func TestParserCorrect(t *testing.T) {  
	queryParser := NewQueryParser("a & b & c")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND q.accessors @> ARRAY[$2]::text[] AND q.accessors @> ARRAY[$3]::text[]")
}

func TestParserCorrectWithParentheses(t *testing.T) {  
	queryParser := NewQueryParser("(a & b & c)")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.Query, "(q.accessors @> ARRAY[$1]::text[] AND q.accessors @> ARRAY[$2]::text[] AND q.accessors @> ARRAY[$3]::text[])")
}

func TestParserCorrectWithMultipleParentheses(t *testing.T) {  
	queryParser := NewQueryParser("(a & (b | c))")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.Query, "(q.accessors @> ARRAY[$1]::text[] AND (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[]))")
}

func TestParserCorrectWithMultipleParentheses2(t *testing.T) {  
	queryParser := NewQueryParser("(a | (b | c))")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.Query, "(q.accessors @> ARRAY[$1]::text[] OR (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[]))")
}

func TestParserCorrectWithMultipleParentheses3(t *testing.T) {  
	queryParser := NewQueryParser("((a | (b | c)))")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.Query, "((q.accessors @> ARRAY[$1]::text[] OR (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[])))")
}

func TestParserIncorrectSyntax(t *testing.T) {  
	queryParser := NewQueryParser("a & b & |")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectSyntax2(t *testing.T) {  
	queryParser := NewQueryParser("a |")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectBrackets(t *testing.T) {  
	queryParser := NewQueryParser("a & b & )")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectLength(t *testing.T) {  
	queryParser := NewQueryParser("a & b & c & d & e & f & f & g & h & z & u)")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserKeepWhitespacesInAssigment(t *testing.T) {  
	queryParser := NewQueryParser("a | b.c = 'hello world'")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b.c='hello world'"})
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] OR q.accessors @> ARRAY[$2]::text[]")
}

func TestParserKeepWhitespacesInAssigmentWithMultipleSpacesBefore(t *testing.T) {  
	queryParser := NewQueryParser("a | b.c   =   'hello world'")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b.c='hello world'"})
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] OR q.accessors @> ARRAY[$2]::text[]")
}

func TestParserAssignment(t *testing.T) {  
	queryParser := NewQueryParser("hello='world'")
	_, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestParserAssignment1(t *testing.T) {  
	queryParser := NewQueryParser("a & hello='big world'")
	_, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestComplexQuery(t *testing.T) {
	queryParser := NewQueryParser("a & (b | c) | d")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 4)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c", "d"})
	equals(t, parseResult.Subquery, "a.accessor = $1 OR (a.accessor = $2 OR a.accessor = $3) OR a.accessor = $4")
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[]) OR q.accessors @> ARRAY[$4]::text[]")
}

func TestWrongComplexQuery(t *testing.T) {
	queryParser := NewQueryParser("a & (b | c) | d )")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected nil, but got not nil")
	}
}

func TestAnotherWrongComplexQuery(t *testing.T) {
	queryParser := NewQueryParser("a & (b | c) | d ()")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestAnotherComplexQuery(t *testing.T) {
	queryParser := NewQueryParser("(a & (b | c) | d)")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 4)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c", "d"})
	equals(t, parseResult.Subquery, "(a.accessor = $1 OR (a.accessor = $2 OR a.accessor = $3) OR a.accessor = $4)")
	equals(t, parseResult.Query, "(q.accessors @> ARRAY[$1]::text[] AND (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[]) OR q.accessors @> ARRAY[$4]::text[])")
}

func TestWrongComplexQuery1(t *testing.T) {
	queryParser := NewQueryParser(")a & (b | c) | d)")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestWrongComplexQuery2(t *testing.T) {
	queryParser := NewQueryParser("(a & (b | c) | d")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestComplexQuery1(t *testing.T) {
	queryParser := NewQueryParser("(a & (string with spaces | c) | d)")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 4)
	equals(t, parseResult.QueryValues, []interface{}{"a", "string with spaces", "c", "d"})
	equals(t, parseResult.Subquery, "(a.accessor = $1 OR (a.accessor = $2 OR a.accessor = $3) OR a.accessor = $4)")
	equals(t, parseResult.Query, "(q.accessors @> ARRAY[$1]::text[] AND (q.accessors @> ARRAY[$2]::text[] OR q.accessors @> ARRAY[$3]::text[]) OR q.accessors @> ARRAY[$4]::text[])")
}

func TestResultSetComplexQuery1(t *testing.T) {
	queryParser := NewQueryParser("(a & (string with spaces | c) | d)")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}

	ref := []string{"a", "string with spaces", "c", "d"}
	reflect.DeepEqual(parseResult.QueryValues, ref)
}

func TestNotOperator(t *testing.T) {
	queryParser := NewQueryParser("a & ~b")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Subquery, "a.accessor = $1 OR NOT a.accessor = $2")
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND NOT q.accessors @> ARRAY[$2]::text[]")
}

func TestNotOperator1(t *testing.T) {
	queryParser := NewQueryParser("~a & b")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Subquery, "NOT a.accessor = $1 OR a.accessor = $2")
	equals(t, parseResult.Query, "NOT q.accessors @> ARRAY[$1]::text[] AND q.accessors @> ARRAY[$2]::text[]")
}

func TestNotOperator2(t *testing.T) {
	queryParser := NewQueryParser("~(a | b)")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Subquery, "NOT (a.accessor = $1 OR a.accessor = $2)")
	equals(t, parseResult.Query, "NOT (q.accessors @> ARRAY[$1]::text[] OR q.accessors @> ARRAY[$2]::text[])")
}

func TestNotOperatorShouldFailDueToWrongSyntax(t *testing.T) {
	queryParser := NewQueryParser("~~a | b)")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestNotOperatorShouldFailDueToWrongSyntax1(t *testing.T) {
	queryParser := NewQueryParser("~&a | b)")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestNotOperator3(t *testing.T) {
	queryParser := NewQueryParser("~a")
	parseResult, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
	equals(t, len(parseResult.QueryValues), 1)
	equals(t, parseResult.QueryValues, []interface{}{"a"})
	equals(t, parseResult.Subquery, "NOT a.accessor = $1")
	equals(t, parseResult.Query, "NOT q.accessors @> ARRAY[$1]::text[]")
}

func TestNotOperator6(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b")
	_, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestQueryAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage=10%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Subquery, "(NOT a.accessor = $1) OR a.accessor = $2")
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage=10")
}

/*func TestQueryAnnotationCoverage1(t *testing.T) {
	queryParser := NewQueryParser("annotation.coverage < 50%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "q.annotated_percentage<50")
}*/

func TestQueryAnnotationCoverageMultipleWhitespaces(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage = 10%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage=10")
}

func TestQueryWrongAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.cov=1")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestQueryAnnotationCoverageOperator(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage>=1%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage>=1")
}

func TestQueryAnnotationCoverageOperator1(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage<=50%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage<=50")
}

func TestQueryAnnotationCoverageOperator2(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage=70%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage=70")
}

func TestQueryAnnotationCoverageOperator3(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage<50%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage<50")
}

func TestQueryAnnotationCoverageOperator4(t *testing.T) {
	queryParser := NewQueryParser("(~a) & b & annotation.coverage>50%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 2)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b"})
	equals(t, parseResult.Query, "(NOT q.accessors @> ARRAY[$1]::text[]) AND q.accessors @> ARRAY[$2]::text[] AND q.annotated_percentage>50")
}

func TestQueryMultipleAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParser("annotation.coverage > 10% & annotation.coverage < 10%")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "q.annotated_percentage>10 AND q.annotated_percentage<10")
}

func TestQueryImageWidth(t *testing.T) {
	queryParser := NewQueryParser("image.width > 50px")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "image_width>50")
}

func TestQueryImageHeight(t *testing.T) {
	queryParser := NewQueryParser("image.height > 50px")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "q.image_height>50")
}

func TestQueryImageHeightShouldFailDueToWrongFormat(t *testing.T) {
	queryParser := NewQueryParser("image.height > xxpx")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	notOk(t, err)
	equals(t, len(parseResult.QueryValues), 0)
}

func TestQueryImageWidthShouldFailDueToWrongFormat(t *testing.T) {
	queryParser := NewQueryParser("image.width >= abcpx")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	notOk(t, err)
	equals(t, len(parseResult.QueryValues), 0)
}

func TestComplexQuery4(t *testing.T) {
	queryParser := NewQueryParser("apple & image.width > 15px & image.height > 15px")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 1)
	equals(t, parseResult.QueryValues, []interface{}{"apple"})
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND image_width>15 AND q.image_height>15")
}

func TestOrderByValidationFunctionality(t *testing.T) {
	queryParser := NewQueryParser("a & b & c !order by validation desc")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowOrderByValidation(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.OrderBy.Direction, ResultOrderDescDirection)
	equals(t, parseResult.OrderBy.Type, OrderByNumOfExistingValidations)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND q.accessors @> ARRAY[$2]::text[] AND q.accessors @> ARRAY[$3]::text[]")
}

func TestOrderByValidationFunctionalityShouldFailDueToInvalidQuery(t *testing.T) {
	queryParser := NewQueryParser("a & b & c !order by")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowOrderByValidation(true)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestOrderByValidationFunctionalityMissingDirectionShouldDefaultToDesc(t *testing.T) {
	queryParser := NewQueryParser("a & b & c !order by validation")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowOrderByValidation(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 3)
	equals(t, parseResult.QueryValues, []interface{}{"a", "b", "c"})
	equals(t, parseResult.OrderBy.Direction, ResultOrderDescDirection)
	equals(t, parseResult.OrderBy.Type, OrderByNumOfExistingValidations)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] AND q.accessors @> ARRAY[$2]::text[] AND q.accessors @> ARRAY[$3]::text[]")
}

func TestOrderByValidationFunctionalityShouldFailDueToInvalidQuery1(t *testing.T) {
	queryParser := NewQueryParser("a & b & c !order by validation !order by validation asc")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowOrderByValidation(true)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestOrderByValidationFunctionalityShouldFailBecauseDeactivated(t *testing.T) {
	queryParser := NewQueryParser("a & b & c !order by validation")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowOrderByValidation(false)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestQueryImageCollection(t *testing.T) {
	queryParser := NewQueryParser("image.collection='abc'")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowImageCollection(true)
	queryParser.AllowOrderByValidation(false)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "image_collection = $1")
	equals(t, parseResult.QueryValues, []interface{}{"abc"})
}

func TestQueryImageCollection1(t *testing.T) {
	queryParser := NewQueryParser("image.collection='abc with spaces'")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowImageCollection(true)
	queryParser.AllowOrderByValidation(false)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "image_collection = $1")
	equals(t, parseResult.QueryValues, []interface{}{"abc with spaces"})
}

func TestQueryImageCollectionShouldFailBecauseDeactivated(t *testing.T) {
	queryParser := NewQueryParser("image.collection='abc'")
	queryParser.AllowImageCollection(false)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestQueryImageUnlabeledShouldFailBecauseDeactivated(t *testing.T) {
	queryParser := NewQueryParser("image.unlabeled='true'")
	queryParser.AllowImageHasLabels(false)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestQueryImageUnlabeledShouldFailBecauseInvalidValue(t *testing.T) {
	queryParser := NewQueryParser("image.unlabeled='notexisting'")
	queryParser.AllowImageHasLabels(true)
	_, err := queryParser.Parse()
	notOk(t, err)
}

func TestQueryImageUnlabeledShouldSucceedTrue(t *testing.T) {
	queryParser := NewQueryParser("image.unlabeled='true'")
	queryParser.AllowImageHasLabels(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "is_unlabeled = $1")
	equals(t, parseResult.QueryValues, []interface{}{"true"})
}

func TestQueryImageUnlabeledShouldSucceedFalse(t *testing.T) {
	queryParser := NewQueryParser("image.unlabeled='false'")
	queryParser.AllowImageHasLabels(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "is_unlabeled = $1")
	equals(t, parseResult.QueryValues, []interface{}{"false"})
}

func TestQueryImageUnlabeledAndLabelShouldSucceed(t *testing.T) {
	queryParser := NewQueryParser("apple | image.unlabeled='true'")
	queryParser.AllowImageHasLabels(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[] OR is_unlabeled = $2")
	equals(t, parseResult.QueryValues, []interface{}{"apple", "true"})
}

func TestQueryImageWithUnderscoreInLabelName(t *testing.T) {
	queryParser := NewQueryParser("blue_fish")
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[]")
	equals(t, parseResult.QueryValues, []interface{}{"blue_fish"})
}

func TestQueryImageWithSlashInLabelName(t *testing.T) {
	queryParser := NewQueryParser("blue/fish")
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[]")
	equals(t, parseResult.QueryValues, []interface{}{"blue/fish"})
}

func TestQueryImageWithSlashAndUnderscoreInLabelName(t *testing.T) {
	queryParser := NewQueryParser("blue/red_fish")
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, parseResult.Query, "q.accessors @> ARRAY[$1]::text[]")
	equals(t, parseResult.QueryValues, []interface{}{"blue/red_fish"})
}

func TestQueryImageNumLabels1(t *testing.T) {
	queryParser := NewQueryParser("image.num_labels > 50")
	queryParser.AllowImageNumLabels(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "q.image_num_labels>50")
}

func TestQueryImageNumLabels2(t *testing.T) {
	queryParser := NewQueryParser("image.num_labels = 1")
	queryParser.AllowImageWidth(true)
	queryParser.AllowImageHeight(true)
	queryParser.AllowAnnotationCoverage(true)
	queryParser.AllowImageCollection(true)
	queryParser.AllowImageHasLabels(true)
	queryParser.AllowImageNumLabels(true)
	parseResult, err := queryParser.Parse()
	ok(t, err)
	equals(t, len(parseResult.QueryValues), 0)
	equals(t, parseResult.Query, "q.image_num_labels=1")
}
