package parser

import(
	"unicode"
	"errors"
	"strconv"
    "github.com/satori/go.uuid"
	"regexp"
    "strings"
    "../commons"
)

type StaticQueryAttribute struct {
    Operator string 
    Value int
    QueryName string
}

func IsLetter(s string) bool {
    for _, r := range s {
        if !unicode.IsLetter(r) {
            return false
        }
    }
    return true
}

func IsLetterOrSpace(s string) bool {
    for _, r := range s {
        if unicode.IsLetter(r) || unicode.IsSpace(r) {
            continue
        }

        return false
    }
    return true
}

func IsUuid(s string) bool {
    _, err := uuid.FromString(s)
    if err == nil {
        return true
    }
    return false
}

func IsAssignment(s string) bool {
	match, _ := regexp.MatchString("^([a-zA-Z]*\\.)?[a-zA-Z]*='[a-zA-Z\\s]*'$", s)
	return match
}

func IsStaticQueryAttribute(s string) (bool, commons.Token) {
    attributes := commons.GetStaticQueryAttributes()
    for _, v := range attributes { 
        match, _ := regexp.MatchString(v.RegExp, s)
        if match {
            return true, v.BelongsToToken
        }
    }

    return false, commons.UNKNOWN
}

func GetStaticQueryAttributeFromToken(s string, token commons.Token) (StaticQueryAttribute, error) {
    var err error
    staticQueryAttr := commons.GetStaticQueryAttributes()[token]

    staticQueryAttribute := StaticQueryAttribute{Value: 0, Operator: ">=", QueryName: staticQueryAttr.QueryName}
    r := regexp.MustCompile(staticQueryAttr.RegExp)
    matches := r.FindStringSubmatch(s)
    if len(matches) > 2 {
        staticQueryAttribute.Operator = matches[1]
        staticQueryAttribute.Value, err = strconv.Atoi(matches[2])
        if err != nil {
            return staticQueryAttribute, errors.New("Oops: " + matches[2])
        }
    }

    return staticQueryAttribute, nil
}

func StrToToken(str string) commons.Token {
	switch str {
		case "&":
			return commons.AND
		case "|":
			return commons.OR
		case "(":
			return commons.LPAR
		case ")":
			return commons.RPAR
        case "~":
            return commons.NOT
		default:
            isStaticQueryAttribute, tok := IsStaticQueryAttribute(str)
            if isStaticQueryAttribute {
                return tok
            }

			if IsAssignment(str) || IsLetterOrSpace(str) || IsUuid(str) {
				return commons.LABEL
			}
	}

	return commons.UNKNOWN
}

func IsTokenValid(currentToken commons.Token, lastToken commons.Token) bool {
	switch currentToken {
		case commons.AND:
			if commons.IsGeneralLabelToken(lastToken) || (lastToken == commons.RPAR)  {
				return true
			}
			return false
		case commons.OR:
			if commons.IsGeneralLabelToken(lastToken) || (lastToken == commons.RPAR) {
				return true
			}
			return false
		case commons.RPAR:
			if commons.IsGeneralLabelToken(lastToken) {
				return true
			}
			return false
		case commons.LPAR:
			if commons.IsGeneralLabelToken(lastToken) || (lastToken == commons.AND) || (lastToken == commons.OR) || (lastToken == commons.NOT) {
				return true
			}
			return false
        case commons.NOT:
            if commons.IsGeneralLabelToken(lastToken) || (lastToken == commons.AND) || (lastToken == commons.OR) || (lastToken == commons.LPAR) {
                return true
            }
            return false
        default:
            if commons.IsGeneralLabelToken(currentToken) {
                if (lastToken == commons.OR) || (lastToken == commons.AND) || (lastToken == commons.LPAR) || (lastToken == commons.NOT) {
                    return true
                }
            }

            return false
	}

	return false
} 

func IsLastTokenValid(lastToken commons.Token) bool {
    if lastToken == commons.AND {
        return false
    } else if lastToken == commons.OR {
        return false
    }
    return true
}

func Tokenize(input string) []string {
    var output []string
    label := ""
    in := StripWhitespaces(input)
    for _, ch := range in {
        if ch == '&' || ch == '|' || ch == '(' || ch == ')' || ch == '~' {
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

func StripWhitespacesExceptWithinQuotes(s string) string {
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

func StripWhitespaces(s string) string {
    output := ""
    temp := ""
    potentialMultiWordString := false
    for _, r := range s {

        if unicode.IsLetter(r) || unicode.IsSpace(r) || r == '\'' || r == '.' || r == '=' {
            potentialMultiWordString = true
        } else {
            potentialMultiWordString = false

            if temp != "" {
                if strings.Contains(temp, ".") && strings.Contains(temp, "=") { //is it a label assignment?(e.q dog.has='mouth')
                    output += StripWhitespacesExceptWithinQuotes(temp) //remove whitespaces except within the quotes
                } else {
                    output += temp
                }
            }

            output += string(r)

            temp = ""
        }

        if potentialMultiWordString {
            temp += string(r)
        }
    }

    if temp != "" {
        if strings.Contains(temp, ".") && strings.Contains(temp, "=") { //is it a label assignment?(e.q dog.has='mouth')
            output += StripWhitespacesExceptWithinQuotes(temp) //remove whitespaces except within the quotes
        } else {
            output += temp
        }
    }

    return output 
}

type Parser interface {
	Parse() error
}

type QueryParser struct {
	query string
	lastToken commons.Token
    lastStrToken string
	isFirst bool
	brackets int32
    isUuidQuery bool
    version int32
    allowStaticQueryAttributes bool
}

type ParseResult struct {
	Input string
	Query string
    Subquery string
    IsUuidQuery bool
	//validationQuery string
	QueryValues []interface{}
}

func NewQueryParser(query string) *QueryParser {
    return &QueryParser {
        query: query,
        isFirst: true,
        version: 1,
        allowStaticQueryAttributes: false,
    } 
}

func NewQueryParserV2(query string) *QueryParser {
    return &QueryParser {
        query: query,
        isFirst: true,
        version: 2,
    } 
}

func (p *QueryParser) _getErrorMessage(token string, isFirst bool) string {
    e := ""
    if p.allowStaticQueryAttributes {
        if isFirst {
            e = "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), 'ATTRIBUTE' (e.q " + 
              commons.GetStaticQueryAttributes()[commons.ANNOTATION_COVERAGE].Name + "= 10%), '~' or '('"
        } else {
            e = "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), 'ATTRIBUTE' (e.q " +
                 commons.GetStaticQueryAttributes()[commons.ANNOTATION_COVERAGE].Name + "= 10%), 'ASSIGNMENT' (e.q dog.breed='Labrador'), '|', '&' or '~' "
        }
    } else {
        if isFirst {
            e = "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), '~' or '('"
        } else {
            e = "Error: invalid token\n" + token + "\n^\nExpecting 'LABEL' (e.q dog), 'ASSIGNMENT' (e.q dog.breed='Labrador'), '|', '&' or '~' "
        }
    }

    return e
}

func (p *QueryParser) AllowStaticQueryAttributes(allow bool) {
    p.allowStaticQueryAttributes = allow
}

func (p *QueryParser) Parse(offset int) (ParseResult, error) {
	parseResult := ParseResult{}
	parseResult.Query = ""
    lastSubqueryOperator := ""
    //parseResult.isUuidQuery = p.isUuidQuery
    parseResult.IsUuidQuery = false

    tokens := Tokenize(p.query)

    i := offset
    numOfLabels := 1
    for _, token := range tokens {
        //strip tailing and leading white spaces
        token = strings.TrimSpace(token)

        if token == "" {
            continue
        }

    	t := StrToToken(token)
    	if p.isFirst {
    		if !((t == commons.LABEL) || (t == commons.LPAR) || (t == commons.NOT) || (p.allowStaticQueryAttributes && commons.IsGeneralLabelToken(t))) {
    			return parseResult, errors.New(p._getErrorMessage(token, true))
    		}

            //use the first entry to determine whether its a UUID or not. We can't have both labels and UUIDs in the same query, so
            //we use the first entry to determine the type of the query.
            parseResult.IsUuidQuery = IsUuid(token)


    		p.isFirst = false
    	} else {
    		if !IsTokenValid(t, p.lastToken) {
    			return parseResult, errors.New(p._getErrorMessage(token, false))
    		}
    	}

    	if t == commons.LABEL {
            if p.version == 1 {
                if parseResult.IsUuidQuery {
                    parseResult.Query += ("l.uuid = $" + strconv.Itoa(i))
                } else {
                    parseResult.Query += ("a.accessor = $" + strconv.Itoa(i))
                }
            } else {
                if !parseResult.IsUuidQuery {
                    parseResult.Query += ("q.accessors @> ARRAY[$" + strconv.Itoa(i) + "]::text[]")
                    
                    if lastSubqueryOperator != "" {
                        parseResult.Subquery += (lastSubqueryOperator + " ")
                        lastSubqueryOperator = ""
                    }
                    parseResult.Subquery += ("a.accessor = $" + strconv.Itoa(i) + " ")
                }
            }
    		parseResult.QueryValues = append(parseResult.QueryValues, token)
    		i += 1
    		numOfLabels += 1
    	} else if t == commons.AND {
    		parseResult.Query += "AND"
            lastSubqueryOperator = "OR"
    	} else if t == commons.OR {
    		parseResult.Query += "OR"
            lastSubqueryOperator = "OR"
        } else if t == commons.NOT {
            parseResult.Query += "NOT"
            lastSubqueryOperator = "NOT"
    	} else if t == commons.LPAR {
            p.brackets += 1
        } else if t == commons.RPAR {
            p.brackets -= 1
        } else if t != commons.UNKNOWN {
            if p.allowStaticQueryAttributes {
                if !parseResult.IsUuidQuery {
                    staticQueryAttribute, err := GetStaticQueryAttributeFromToken(token, t)
                    if err != nil {
                        return parseResult, err
                    }
                    parseResult.Query += (staticQueryAttribute.QueryName + staticQueryAttribute.Operator + strconv.Itoa(staticQueryAttribute.Value))
                }
            } else {
                return parseResult, errors.New(p._getErrorMessage(token, false))
            }
        }
    	parseResult.Query += " "

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
