package main

import "testing"

/*func TestParserCorrect(t *testing.T) {  
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

func TestParserIncorrectLength(t *testing.T) {  
	queryParser := NewQueryParser("a & b & c & d & e & f & f & g & h & z & u)")
	_, err := queryParser.Parse()
	if err == nil {
		t.Errorf("Expected not nil, but got nil")
	}
}*/

func TestParserKeepWhitespacesInAssigment(t *testing.T) {  
	queryParser := NewQueryParser("a | b.c = 'hello world'")
	_, err := queryParser.Parse()
	if err != nil {
		t.Errorf("Expected nil, but got not nil: %s", err.Error())
	}
}