package commons

import (
	"github.com/google/go-jsonnet"
	"encoding/json"
	"path/filepath"
	"io/ioutil"
	"../datastructures"
)

type LabelRepository struct {
    labelMap datastructures.LabelMap
    words []string
    pluralsMap map[string]string
}

func NewLabelRepository() *LabelRepository {
    return &LabelRepository{} 
}

func (p *LabelRepository) Load(path string) error {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }

    vm := jsonnet.MakeVM()

    dir, _ := filepath.Split(path)
    vm.Importer(&jsonnet.FileImporter{
        JPaths: []string{dir},
    })

    out, err := vm.EvaluateSnippet("file", string(data))
    if err != nil {
        return err
    }

    err = json.Unmarshal([]byte(out), &p.labelMap)
    if err != nil {
        return err
    }

    p.words = make([]string, len(p.labelMap.LabelMapEntries))
    p.pluralsMap = make(map[string]string, len(p.labelMap.LabelMapEntries))
    i := 0
    for key, val := range p.labelMap.LabelMapEntries {
        p.words[i] = key
        p.pluralsMap[key] = val.Plural
        i++
    }

    return nil
}

func (p *LabelRepository) GetMapping() map[string]datastructures.LabelMapEntry {
	return p.labelMap.LabelMapEntries
}

func (p *LabelRepository) GetPluralsMapping() map[string]string {
	return p.pluralsMap
}

func (p *LabelRepository) GetWords() []string {
	return p.words
}
