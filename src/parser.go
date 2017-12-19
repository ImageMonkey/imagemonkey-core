package main

import(
	"strings"
	"unicode"
	"errors"
	"strconv"
	"regexp"
)

type Token int

const (
	LABEL = iota
	AND
	OR
	LPAR
	RPAR 
	UNKNOWN
)

func IsLetter(s string) bool {
    for _, r := range s {
        if !unicode.IsLetter(r) {
            return false
        }
    }
    return true
}

func IsAssignment(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z]*\\.[a-zA-Z]*='[a-zA-Z]*'$", s)
	return match
}

func StrToToken(str string) Token {
	switch str {
		case "&":
			return AND
		case "|":
			return OR
		case "(":
			return LPAR
		case ")":
			return RPAR
		default:
			if IsAssignment(str) || IsLetter(str) {
				return LABEL
			}
	}

	return UNKNOWN
}

func IsTokenValid(currentToken Token, lastToken Token) bool{
	switch currentToken {
		case AND:
			if lastToken == LABEL {
				return true
			}
			return false
		case OR:
			if lastToken == LABEL {
				return true
			}
			return false
		case LABEL:
			if (lastToken == OR) || (lastToken == AND) {
				return true
			}
			return false
		case RPAR:
			if lastToken == LABEL {
				return true
			}
			return false
		case LPAR:
			if (lastToken == LABEL) || (lastToken == AND) || (lastToken == OR) {
				return true
			}
			return false
	}

	return false
} 

func Tokenize(input string) string {
	output := ""
	in := strings.Replace(input, " ", "", -1)
	for _, ch := range in {
		if ch == '&' || ch == '|' || ch == '(' || ch == ')' {
			output += (" " + string(ch) + " ")
		} else {
			output += string(ch)
		}
	}
	return output
}

func GetBaseLabel(s string) string {
	for i, r := range s {
        if r == '.' {
            return s[:i]
        }
    }
    return s
}

type Parser interface {
	Parse() error
}

type QueryParser struct {
	query string
	lastToken Token
	isFirst bool
	brackets int32
}

type ParseResult struct {
	input string
	annotationQuery string
	validationQuery string
	queryValues []interface{}
}

func NewQueryParser(query string) *QueryParser {
    return &QueryParser {
        query: query,
        isFirst: true,
    } 
}

func (p *QueryParser) Parse() (ParseResult, error) {
	parseResult := ParseResult{}
	parseResult.annotationQuery = ""
	parseResult.validationQuery = ""

	in := Tokenize(p.query)
    tokens := strings.Split(in, " ")

    var temp []string
    var validationQueryValues []interface{}
    i := 1
    for _, token := range tokens {
    	if token == "" {
    		continue
    	}

    	t := StrToToken(token)
    	if p.isFirst {
    		if t != LABEL {
    			e := "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), 'ASSIGNMENT' (e.q dog.breed='Labrador'), '|' "
    			return parseResult, errors.New(e)
    		}

    		p.isFirst = false
    	} else {
    		if !IsTokenValid(t, p.lastToken) {
    			e := "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), 'ASSIGNMENT' (e.q dog.breed='Labrador'), '|' "
    			return parseResult, errors.New(e)
    		}
    	}

    	if t == LABEL {
    		parseResult.annotationQuery += ("q.accessor = $" + strconv.Itoa(i))
    		temp = append(temp, "$")
    		parseResult.queryValues = append(parseResult.queryValues, token)
    		validationQueryValues = append(validationQueryValues, GetBaseLabel(token)) //for the validation query we need only the base label (i.e label before the '.')
    		i += 1
    	} else if t == AND {
    		parseResult.annotationQuery += "AND"
    		temp = append(temp, "AND")
    	} else if t == OR {
    		parseResult.annotationQuery += "OR"
    		temp = append(temp, "OR")
    	} else {
    		parseResult.annotationQuery += token
    		temp = append(temp, "token")
    	}
    	parseResult.annotationQuery += " "


    	if t == LPAR {
    		p.brackets += 1
    	}
    	if t == RPAR {
    		p.brackets -= 1
    	}

    	p.lastToken = t
    }

    //adapt positional arguments so that they start at startPos
    startPos := i
    for _, val := range temp {
    	if val == "$" {
    		parseResult.validationQuery += ("a.accessor = $" + strconv.Itoa(startPos) + " ")
    		startPos += 1
    	} else {
    		parseResult.validationQuery += (val + " ")
    	}
    }
    
    parseResult.queryValues = append(parseResult.queryValues, validationQueryValues...)


    if p.brackets != 0 {
    	return parseResult, errors.New("brackets mismatch!")
    }

    return parseResult, nil
}
