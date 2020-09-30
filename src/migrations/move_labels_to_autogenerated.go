package main

import (
	"github.com/google/go-jsonnet"
	"encoding/json"
	"io/ioutil"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	log "github.com/sirupsen/logrus"
	"flag"
	"github.com/gofrs/uuid"
	"bufio"
	"os"
)


func readLabels(path string) (datastructures.LabelMap, error) {
    var labelMap datastructures.LabelMap

	data, err := ioutil.ReadFile(path)
    if err != nil {
        return labelMap, err
    }

    vm := jsonnet.MakeVM()

    out, err := vm.EvaluateSnippet("file", string(data))
    if err != nil {
        return labelMap, err
    }

    err = json.Unmarshal([]byte(out), &labelMap)
    if err != nil {
        return labelMap, err
    }

    return labelMap, nil
}

func main() {
	inputPath := flag.String("input-path", "", "Input Path")
	outputDir := flag.String("output-dir", "", "Output Directory")

	flag.Parse()

	if *inputPath == "" {
		log.Fatal("Please provide a input path")
	}
	
	if *outputDir == "" {
		log.Fatal("Please provide a output directory")
	}
	
	labels, err := readLabels(*inputPath)
	if err != nil {
		log.Fatal("Couldn't read labels")
	}

	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		log.Fatal(*outputDir, " is not a directory!")
	}

	for key, entry := range labels.LabelMapEntries {
		u, err := uuid.NewV4()
		if err != nil {
			log.Fatal("Couldn't create UUID")
		}
		outputPath := *outputDir + "/" + u.String() + ".json"
		

		var e map[string]datastructures.LabelMapEntry
		e = make(map[string]datastructures.LabelMapEntry)
		e[key] = entry

		j, err := json.MarshalIndent(&e, "", "  ")
		if err != nil {
			log.Fatal("Couldn't marshal json")
		}

		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatal("Couldn't create file")
		}

		defer f.Close()

		log.Info(string(j))

		w := bufio.NewWriter(f)
    	_, err = w.WriteString(string(j))
		if err != nil {
			log.Fatal("Couldn't write file")
		}

		w.Flush()
	}
}
