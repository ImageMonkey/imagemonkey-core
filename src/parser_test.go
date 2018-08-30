package main

import (
	"testing"
	"reflect"
	"runtime"
	"fmt"
	"path/filepath"
	"strings"
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
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestParserIncorrectSyntax(t *testing.T) {  
	queryParser := NewQueryParser("a & b & |")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectSyntax2(t *testing.T) {  
	queryParser := NewQueryParser("a |")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectBrackets(t *testing.T) {  
	queryParser := NewQueryParser("a & b & )")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserIncorrectLength(t *testing.T) {  
	queryParser := NewQueryParser("a & b & c & d & e & f & f & g & h & z & u)")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}

func TestParserKeepWhitespacesInAssigment(t *testing.T) {  
	queryParser := NewQueryParser("a | b.c = 'hello world'")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}

	//t.Errorf("a = %s\n", parseResult.queryValues)
}

func TestParserAssignment(t *testing.T) {  
	queryParser := NewQueryParser("hello='world'")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestParserAssignment1(t *testing.T) {  
	queryParser := NewQueryParser("a & hello='big world'")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestComplexQuery(t *testing.T) {
	queryParser := NewQueryParserV2("a & (b | c) | d")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestWrongComplexQuery(t *testing.T) {
	queryParser := NewQueryParserV2("a & (b | c) | d )")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestAnotherWrongComplexQuery(t *testing.T) {
	queryParser := NewQueryParserV2("a & (b | c) | d ()")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestAnotherComplexQuery(t *testing.T) {
	queryParser := NewQueryParserV2("(a & (b | c) | d)")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestWrongComplexQuery1(t *testing.T) {
	queryParser := NewQueryParserV2(")a & (b | c) | d)")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestWrongComplexQuery2(t *testing.T) {
	queryParser := NewQueryParserV2("(a & (b | c) | d")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}


func TestComplexQuery1(t *testing.T) {
	queryParser := NewQueryParserV2("(a & (string with spaces | c) | d)")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestResultSetComplexQuery1(t *testing.T) {
	queryParser := NewQueryParserV2("(a & (string with spaces | c) | d)")
	parseResult, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}

	ref := []string{"a", "string with spaces", "c", "d"}
	reflect.DeepEqual(parseResult.queryValues, ref)
}

func TestNotOperator(t *testing.T) {
	queryParser := NewQueryParserV2("a & ~b")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestNotOperator1(t *testing.T) {
	queryParser := NewQueryParserV2("~a & b")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestNotOperator2(t *testing.T) {
	queryParser := NewQueryParserV2("~(a | b)")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestNotOperator3(t *testing.T) {
	queryParser := NewQueryParserV2("~~a | b)")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestNotOperator4(t *testing.T) {
	queryParser := NewQueryParserV2("~&a | b)")
	_, err := queryParser.Parse(1)
	if err == nil {
		t.Errorf("Expected not nil, but got nil: %s", err.Error())
	}
}

func TestNotOperator5(t *testing.T) {
	queryParser := NewQueryParserV2("~a")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestNotOperator6(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b")
	_, err := queryParser.Parse(1)
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}

func TestQueryAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage=10%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage=10"), true)
}

func TestQueryAnnotationCoverage1(t *testing.T) {
	queryParser := NewQueryParserV2("annotation.coverage < 50%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, len(parseResult.queryValues), 1)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage<50"), true)
}

func TestQueryAnnotationCoverageMultipleWhitespaces(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage = 10%")
	queryParser.AllowStaticQueryAttributes(true)
	_, err := queryParser.Parse(1)
	ok(t, err)
}

func TestQueryWrongAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.cov=1")
	queryParser.AllowStaticQueryAttributes(true)
	_, err := queryParser.Parse(1)
	notOk(t, err)
}

func TestQueryAnnotationCoverageOperator(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage>=1%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage>=1"), true)
}

func TestQueryAnnotationCoverageOperator1(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage<=50%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage<=50"), true)
}

func TestQueryAnnotationCoverageOperator2(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage=70%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage=70"), true)
}

func TestQueryAnnotationCoverageOperator3(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage<50%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage<50"), true)
}

func TestQueryAnnotationCoverageOperator4(t *testing.T) {
	queryParser := NewQueryParserV2("(~a) & b & annotation.coverage>50%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage>50"), true)
}

func TestQueryMultipleAnnotationCoverage(t *testing.T) {
	queryParser := NewQueryParserV2("annotation.coverage > 10% & annotation.coverage < 10%")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.annotated_percentage>10 AND q.annotated_percentage<10"), true)
}

func TestQueryImageWidth(t *testing.T) {
	queryParser := NewQueryParserV2("image.width > 50px")
	queryParser.AllowStaticQueryAttributes(true)
	parseResult, err := queryParser.Parse(1)
	ok(t, err)
	equals(t, strings.Contains(parseResult.query, "q.image_width>50"), true)
}


