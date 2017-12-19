package main

import "testing"

func TestParserCorrect(t *testing.T) {  
	queryParser := NewQueryParser("a & b & c")
	_, err := queryParser.Parse()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestParserIncorrectSyntax(t *testing.T) {  
	queryParser := NewQueryParser("a & b & |")
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