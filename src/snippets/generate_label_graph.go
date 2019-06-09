package main

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"github.com/gofrs/uuid"
	"io/ioutil"
	"strings"
	"strconv"
)

func generateRandomGraph(num int) string {
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
	    panic(err)
	}

	i := 0
	prev := ""
	for i < num {
		u := uuid.NewV4().String()
		u = strings.Replace(u, "-", "", -1)
		u = strconv.Quote(u)

		fmt.Printf("%d", i)

		graph.AddNode("G", u, nil)

		if prev != "" {
			graph.AddEdge(prev, u, true, nil)
		}

		prev = u

		i += 1
	}
	
	output := graph.String()

	return output
}

func main() {
	fmt.Printf("a")


	output := generateRandomGraph(30000)
	fmt.Printf("writing")

	path := "../wordlists/en/graph_big.dot"
	err := ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	//fmt.Printf(output)
}
