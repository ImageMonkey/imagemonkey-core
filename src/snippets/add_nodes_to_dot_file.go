package main

import (
	"github.com/awalterschulze/gographviz"
	"os"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"io/ioutil"
)

func standardizeUnderlines(s string) string {
	f := func(c rune) bool {
		return c == '_'
	}

    return strings.Join(strings.FieldsFunc(s, f), "_")
}

func main() {
	path := "../../wordlists/en/graphdefinitions/imagenet.dot"

	f, err := os.Open(path)
	if err != nil{
		panic(err)
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)

	graphAst, err := gographviz.ParseString(buf.String())
	if err != nil {
		panic(err)
	}
	
	graph := gographviz.NewGraph()
	graph.SetDir(true)
	if err := gographviz.Analyse(graphAst, graph); err != nil {
    	panic(err)
	}

	labelNames := make(map[string]string)

	edgeStrs := ""
	edges :=graph.Edges.Edges
	for _, edge := range edges {
		unquotedEdgeSrc,err := strconv.Unquote(edge.Src)
		if err != nil {
			fmt.Printf("%s\n", edge.Src)
			panic(err)
		}

		strippedEdgeSrc := strings.Replace(unquotedEdgeSrc, " ", "_", -1)
		strippedEdgeSrc = strings.Replace(strippedEdgeSrc, "(", "_", -1)
		strippedEdgeSrc = strings.Replace(strippedEdgeSrc, ")", "_", -1)
		strippedEdgeSrc = strings.Trim(strippedEdgeSrc, "_")
		strippedEdgeSrc = standardizeUnderlines(strippedEdgeSrc)
		labelNames[strippedEdgeSrc] = unquotedEdgeSrc

		unquotedEdgeDst, err := strconv.Unquote(edge.Dst)
		if err != nil {
			fmt.Printf("%s\n", edge.Dst)
			panic(err)
		}

		strippedEdgeDst := strings.Replace(unquotedEdgeDst, " ", "_", -1)
		strippedEdgeDst = strings.Replace(strippedEdgeDst, "(", "_", -1)
		strippedEdgeDst = strings.Replace(strippedEdgeDst, ")", "_", -1)
		strippedEdgeDst = strings.Trim(strippedEdgeDst, "_")
		strippedEdgeDst = standardizeUnderlines(strippedEdgeDst)
		labelNames[strippedEdgeDst] = unquotedEdgeDst

		edgeStrs += (strippedEdgeSrc + "->" + strippedEdgeDst + "\n")
	}

	nodeStrs := ""
	for k, v := range labelNames {
		nodeStrs += (k + " [label=\"" + v +"\"]" + "\n")
	}

	complete := nodeStrs + "\n\n" + edgeStrs

	err = ioutil.WriteFile("temp.dot", []byte(complete), 0644)
	if err != nil {
		panic(err)
	}
}