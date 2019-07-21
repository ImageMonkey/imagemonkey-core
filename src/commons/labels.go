package commons

import (
	"github.com/google/go-jsonnet"
	"encoding/json"
	"path/filepath"
	"io/ioutil"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"os"
)

type LabelRepository struct {
    labelMap datastructures.LabelMap
    words []string
    pluralsMap map[string]string
	path string
}

func NewLabelRepository(path string) *LabelRepository {
    return &LabelRepository{
		path: path,
	} 
}

func (p *LabelRepository) Load() error {
    data, err := ioutil.ReadFile(p.path)
    if err != nil {
        return err
    }

    vm := jsonnet.MakeVM()

	dir, _ := filepath.Split(p.path)
	dir = dir + string(os.PathSeparator) + "includes" + string(os.PathSeparator) + "labels/"
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

func (p *LabelRepository) Contains(label string, sublabel string) bool {
	sublabels := []datastructures.Sublabel{}
	if sublabel != "" {
		sublabels = append(sublabels, datastructures.Sublabel{Name: sublabel}) 
	}

	if val, ok := p.labelMap.LabelMapEntries[label]; ok {
        if len(sublabels) > 0 {
            availableSublabels := val.LabelMapEntries
        	for _, value := range sublabels {
                _, ok := availableSublabels[value.Name]
                if !ok {
                    return false
                }
            }
            return true
        }
        return true
    }
	return false
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


type MetaLabels struct {
    metalabels datastructures.MetaLabelMap
    path string
}

func NewMetaLabels(path string) *MetaLabels {
    return &MetaLabels {
        path: path,
    } 
}

func (p *MetaLabels) Load() error {
	data, err := ioutil.ReadFile(p.path)
    if err != nil {
        return err
    }

	vm := jsonnet.MakeVM()

	dir, _ := filepath.Split(p.path)
	dir = dir + string(os.PathSeparator) + "includes" + string(os.PathSeparator) + "metalabels/"
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{dir},
	})

	out, err := vm.EvaluateSnippet("file", string(data))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(out), &p.metalabels)
    if err != nil {
        return err
    }

    return nil
}

func (p *MetaLabels) GetMapping() datastructures.MetaLabelMap {
    return p.metalabels
}

func (p *MetaLabels) Contains(val string) bool {
    if _, ok := p.metalabels.MetaLabelMapEntries[val]; ok {
        return true
    }

    return false
}


func IsLabelValid(labelsMap map[string]datastructures.LabelMapEntry, metalabels *MetaLabels, 
                    label string, sublabels []datastructures.Sublabel) bool {
    if val, ok := labelsMap[label]; ok {
        if len(sublabels) > 0 {
            availableSublabels := val.LabelMapEntries
        	for _, value := range sublabels {
                _, ok := availableSublabels[value.Name]
                if !ok {
                    return false
                }
            }
            return true
        }
        return true
    }

    if metalabels.Contains(label) {
        return true
    }

    return false
}
