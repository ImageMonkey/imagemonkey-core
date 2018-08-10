package main

import (
	"testing"
	"reflect"
)


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