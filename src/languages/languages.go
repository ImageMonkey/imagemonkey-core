package languages

type Language struct {
	Name string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

func GetAllSupported() map[string]Language {
	 m := make(map[string]Language)
	 m["en"] = Language{Name: "English", Abbreviation: "EN"}
	 m["ger"] = Language{Name: "German", Abbreviation: "GER"}
	 return m
}

func IsValid(language string) bool {
	if language == "en" || language == "ger" {
		return true
	}

	return false
}