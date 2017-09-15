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
func getWordLists(path string) ([]string, error) {
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