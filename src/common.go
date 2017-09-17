package main

import (
	"math/rand"
	"time"
	"strings"
	"io/ioutil"
)

type Report struct {
    Reason string `json:"reason"`
}

type Label struct {
    Name string `json:"name"`
}

func use(vals ...interface{}) {
    for _, val := range vals {
        _ = val
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

/*
 * Loads all data in memory.
 * If file gets too big, refactor it!
 */
func getWordLists(path string) ([]Label, error) {
    var lines []string
    var labels []Label
	data, err := ioutil.ReadFile(path)
    if(err != nil){
        return labels, err
    }
    lines = strings.Split(string(data), "\r\n")
    for _, v := range lines {
        var label Label
        label.Name = v
        labels = append(labels, label)
    }

    return labels, nil
}

func getStrWordLists(path string) ([]string, error) {
    var lines []string
    data, err := ioutil.ReadFile(path)
    if(err != nil){
        return lines, err
    }
    lines = strings.Split(string(data), "\r\n")
    return lines, nil
}

/*func prettyPrintJSON(b []byte) ([]byte, error) {
    var out bytes.Buffer
    err := JSON.Indent(&out, b, "", "    ")
    return out.Bytes(), err
}*/