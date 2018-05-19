package main

import (
	"bytes"
	"os"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"time"
	//"io/ioutil"
	//"encoding/json"
	"strconv"
)

type LabelGraphNode struct {
    Id int `json:"id"`
    Idenfifier string `json:"identifier"`
    Name string `json:"name"`
    Size string `json:"size"`
    FontSize string `json:"fontsize"`
    Color string `json:"color"`
    Uuid string `json:"uuid"`
}

type LabelGraphEdge struct {
    Source int `json:"source"`
    Target int `json:"target"`
}


type LabelGraphJson struct {
    Nodes []LabelGraphNode `json:"nodes"`
    Links []LabelGraphEdge `json:"links"`
}

type LabelGraph struct {
    path string
    graphDefinition string
    graph *gographviz.Graph
}

func NewLabelGraph(path string) *LabelGraph {
    return &LabelGraph{
        path: path,
        graphDefinition: "",
    } 
}

func (p *LabelGraph) Load() error {
    f, err := os.Open(p.path)
	if err != nil{
		return err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	p.graphDefinition = buf.String()

	graphAst, _ := gographviz.ParseString(p.graphDefinition)
	p.graph = gographviz.NewGraph()
	p.graph.SetDir(true)
	if err := gographviz.Analyse(graphAst, p.graph); err != nil {
    	return err
	}

	return nil
}


func (p *LabelGraph) GetChildren(identifier string) []*gographviz.Node {
	var result []*gographviz.Node

	var identifiers []string
	identifiers = append(identifiers, identifier)

	var innerFct func(graph *gographviz.Graph, identifiers []string)

	innerFct = func(graph *gographviz.Graph, identifiers []string) {

		if(len(identifiers) == 0) {
			return
		}

		//remove first element from list
		identifier := identifiers[0]
		identifiers = append(identifiers[:0], identifiers[1:]...)

		children := p.graph.Edges.SrcToDsts[identifier]
		for _, child := range children {
		    for _,c := range child {
		    	identifiers = append(identifiers, c.Dst)
		    	result = append(result, p.graph.Nodes.Lookup[c.Dst])
		    }
		}

		innerFct(p.graph, identifiers)
	}

	
	innerFct(p.graph, identifiers)

	return result
}

/*func (p *LabelGraph) GetChildrenIds(identifier string) []string {
	var uuids []string

	children := p.GetChildren(identifier)
	for _, child := range children {
		uuid := child.Attrs["id"]
		if uuid == "" {
			continue
		}
		uuids = append(uuids, uuid)
	}

	return uuids
}*/


func (p *LabelGraph) GetJson() (LabelGraphJson, error) {
	var result LabelGraphJson
	var err error

	m := make(map[string]int)
	nodes := p.graph.Nodes.Nodes
	for i, node := range nodes {

		var labelGraphNode LabelGraphNode
		labelGraphNode.Id = i
		labelGraphNode.Idenfifier = node.Name
		labelGraphNode.Size = node.Attrs["size"]
		if labelGraphNode.Size == "" {
			labelGraphNode.Size = "100"
		}
		labelGraphNode.FontSize = node.Attrs["fontsize"]
		if labelGraphNode.FontSize == "" {
			labelGraphNode.FontSize = "14"
		}
		labelGraphNode.Color = node.Attrs["color"]
		labelGraphNode.Uuid = node.Attrs["id"]
		labelGraphNode.Name, _ = strconv.Unquote(node.Attrs["label"])
		if err != nil {
			return result, err
		}

		m[node.Name] = i

		result.Nodes = append(result.Nodes, labelGraphNode)
	}

	edges :=p.graph.Edges.Edges
	for _, edge := range edges {
		var labelGraphEdge LabelGraphEdge
		labelGraphEdge.Source = m[edge.Src]
		labelGraphEdge.Target = m[edge.Dst] //edge.Dst

		result.Links = append(result.Links, labelGraphEdge)
	}

	return result, nil
}



/*func getGraphDefinition(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil{
		return "", err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	contents := buf.String()

	return contents, nil
}*/

/*func findAllChildren(graph *gographviz.Graph, labels []string) {

	if(len(labels) == 0) {
		return
	}

	//remove first element from list
	label := labels[0]
	labels = append(labels[:0], labels[1:]...)

	childrens := graph.Edges.SrcToDsts[label]
	for _, child := range childrens {
	    for _,c := range child {
	    	labels = append(labels, c.Dst)
	    	bla = append(bla, c.Dst)
	    }
	}

	findAllChildren(graph, labels)
}*/




/*func findAllChildren(graph *gographviz.Graph, label string) []string {
	defer timeTrack(time.Now(), "parsing")

	var result []string

	var labels []string
	labels = append(labels, label)

	var innerFct func(graph *gographviz.Graph, labels []string)

	innerFct = func(graph *gographviz.Graph, labels []string) {

		if(len(labels) == 0) {
			return
		}

		//remove first element from list
		label := labels[0]
		labels = append(labels[:0], labels[1:]...)

		children := graph.Edges.SrcToDsts[label]
		for _, child := range children {
		    for _,c := range child {
		    	labels = append(labels, c.Dst)
		    	result = append(result, c.Dst)
		    }
		}

		innerFct(graph, labels)
	}

	
	innerFct(graph, labels)

	return result
}*/

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s", name, elapsed)
}



/*func main() {
	fmt.Printf("Starting..\n")

	path := "../wordlists/en/graph.dot"
	labelGraph := NewLabelGraph(path)
	err := labelGraph.Load()
	if err != nil {
		panic(err)
	}

	children := labelGraph.GetChildren("vehicle")
	for _,child := range children {
		if _, ok := child.Attrs["id"]; ok {
			fmt.Printf("contains id: %s\n", child.Name)
		}
	}

	labelGraphJson := labelGraph.GetJson()
	out, _ := json.Marshal(labelGraphJson)
    err = ioutil.WriteFile("../html/templates/graph.json", out, 0644)
    if err != nil {
    	panic(err)
    }
*/
	/*//path := "../wordlists/en/graph.dot"
	path := "../wordlists/en/graph_big.dot"

	graphDefinition, err := getGraphDefinition(path)
	if err != nil {
		panic(err)
	}

	graphAst, _ := gographviz.ParseString(graphDefinition)
	graph := gographviz.NewGraph()
	graph.SetDir(true)
	if err := gographviz.Analyse(graphAst, graph); err != nil {
    	panic(err)
	}


	findAllChildren(graph, strconv.Quote("f3ac43bfd30a47139cffd1647e166286"))*/
//}