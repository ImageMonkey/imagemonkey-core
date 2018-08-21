package commons

type QueryAttribute struct {
	InternalIdentifier string
	Name string
	RegExp string
}

func GetStaticQueryAttributes() map[string]QueryAttribute {
	attributes := make(map[string](QueryAttribute))
	attributes["annotation-coverage"] = QueryAttribute{InternalIdentifier: "annotation-coverage", 
																Name: "annotation.coverage",
																RegExp: "annotation\\.coverage[ ]*(>|<|>=|=|<=){1}[ ]*([0-9]*)%"}

	return attributes
}