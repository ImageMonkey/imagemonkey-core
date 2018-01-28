package main

import(
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
	match, _ := regexp.MatchString("^[a-zA-Z]*\\.[a-zA-Z]*='[a-zA-Z\\s]*'$", s)
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

func IsTokenValid(currentToken Token, lastToken Token) bool {
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

func IsLastTokenValid(lastToken Token) bool {
    if lastToken == AND {
        return false
    } else if lastToken == OR {
        return false
    }
    return true
}

func Tokenize(input string) []string {
    var output []string
    label := ""
    in := StripWhitespaces(input)
    for _, ch := range in {
        if ch == '&' || ch == '|' || ch == '(' || ch == ')' {
            if label != "" {
                output = append(output, label)
                label = ""
            }
            output = append(output, string(ch))
            
        } else {
            label += string(ch)
        }
    }

    if label != "" {
        output = append(output, label)
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

func StripWhitespaces(s string) string {
    output := ""
    insideAssignment := false
    for _, r := range s {
        if r == '\'' {
            insideAssignment = !insideAssignment
        }

        if !insideAssignment && r != ' ' {
            output += string(r)
        } else if (insideAssignment) {
            output += string(r)
        }
    }

    return output 
}

type Parser interface {
	Parse() error
}

type QueryParser struct {
	query string
	lastToken Token
    lastStrToken string
	isFirst bool
	brackets int32
}

type ParseResult struct {
	input string
	query string
	//validationQuery string
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
	parseResult.query = ""

    tokens := Tokenize(p.query)

    var temp []string
    i := 1
    numOfLabels := 1
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
    		parseResult.query += ("a.accessor = $" + strconv.Itoa(i))
    		parseResult.queryValues = append(parseResult.queryValues, token)
    		i += 1
    		numOfLabels += 1
    	} else if t == AND {
    		parseResult.query += "AND"
    		temp = append(temp, "AND")
    	} else if t == OR {
    		parseResult.query += "OR"
    		temp = append(temp, "OR")
    	} else {
    		parseResult.query += token
    		temp = append(temp, "token")
    	}
    	parseResult.query += " "


    	if t == LPAR {
    		p.brackets += 1
    	}
    	if t == RPAR {
    		p.brackets -= 1
    	}

    	p.lastToken = t
        p.lastStrToken = token
    }

    if len(tokens) > 0 {
        if !IsLastTokenValid(p.lastToken) {
            e := "Error: invalid token\n" + p.lastStrToken + "\n^\nExpecting 'LABEL' (e.q dog)"
            return parseResult, errors.New(e)
        }
    }

    if numOfLabels > 10 {
    	return parseResult, errors.New("Please limit your query to 10 label expressions")
    }

    if p.brackets != 0 {
    	return parseResult, errors.New("brackets mismatch!")
    }

    return parseResult, nil
}
