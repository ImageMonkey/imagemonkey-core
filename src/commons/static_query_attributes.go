package commons

var attributes = make(map[string](QueryAttribute))

func init() {
	//attributes := make(map[string](QueryAttribute))
	attributes["annotation-coverage"] = QueryAttribute{InternalIdentifier: "annotation-coverage", 
																Name: "annotation.coverage",
																RegExp: "annotation\\.coverage[ ]*(>|<|>=|=|<=){1}[ ]*([0-9]*)%"}
}

type QueryAttribute struct {
	InternalIdentifier string `json:"internal_identifier"`
	Name string `json:"name"`
	RegExp string `json:"regexp"`
}

func GetStaticQueryAttributes() map[string]QueryAttribute {
	return attributes
}