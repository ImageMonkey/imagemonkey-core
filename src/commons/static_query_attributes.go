package commons

type Token int

const (
	LABEL = iota
	AND
	OR
	LPAR
	RPAR 
    NOT
	UNKNOWN
    ANNOTATION_COVERAGE
    IMAGE_WIDTH
    IMAGE_HEIGHT
)


var attributes = make(map[Token](QueryAttribute))

func init() {
	//attributes := make(map[string](QueryAttribute))
	attributes[ANNOTATION_COVERAGE] = QueryAttribute{InternalIdentifier: "annotation-coverage", 
																Name: "annotation.coverage",
																RegExp: "annotation\\.coverage[ ]*(>|<|>=|=|<=){1}[ ]*([0-9]*)%",
																QueryName: "q.annotated_percentage",
																BelongsToToken: ANNOTATION_COVERAGE,
													}


	attributes[IMAGE_WIDTH] = QueryAttribute{InternalIdentifier: "image-width", 
																Name: "image.width",
																RegExp: "image\\.width[ ]*(>|<|>=|=|<=){1}[ ]*([0-9]*)px",
																QueryName: "image_width",
																BelongsToToken: IMAGE_WIDTH,
											}

	attributes[IMAGE_HEIGHT] = QueryAttribute{InternalIdentifier: "image-height", 
																Name: "image.height",
																RegExp: "image\\.height[ ]*(>|<|>=|=|<=){1}[ ]*([0-9]*)px",
																QueryName: "q.image_height",
																BelongsToToken: IMAGE_HEIGHT,
											 }
}

type QueryAttribute struct {
	InternalIdentifier string `json:"internal_identifier"`
	Name string `json:"name"`
	RegExp string `json:"regexp"`
	QueryName string `json:"query_name"`
	BelongsToToken Token `json:"belongs_to"`
}

func GetStaticQueryAttributes() map[Token]QueryAttribute {
	return attributes
}

func IsGeneralLabelToken(t Token) bool {
	if t == LABEL {
		return true
	}

	if _, ok := attributes[t]; ok {
		return true
	}

	return false
}