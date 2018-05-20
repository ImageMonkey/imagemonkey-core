package main

import (
	"bytes"
	"os"
	"github.com/awalterschulze/gographviz"
	"path/filepath"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"errors"
)

type LabelGraphNode struct {
    Id int `json:"id"`
    Idenfifier string `json:"identifier"`
    Name string `json:"name"`
    Size string `json:"size"`
    FontSize string `json:"fontsize"`
    Color string `json:"color"`
    Uuid string `json:"uuid"`
    OnHover string `json:"onhover"`
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

func (p *LabelGraph) GetJson() (LabelGraphJson, error) {
	var result LabelGraphJson
	var err error

	m := make(map[string]int)
	nodes := p.graph.Nodes.Nodes
	for i, node := range nodes {

		var labelGraphNode LabelGraphNode
		labelGraphNode.Id = i
		labelGraphNode.Idenfifier = node.Name
		labelGraphNode.OnHover = node.Attrs["URL"]
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


type LabelGraphRepository struct {
    dir string
    labelGraphs map[string]*LabelGraph
}

func NewLabelGraphRepository(dir string) *LabelGraphRepository {
    return &LabelGraphRepository{
        dir: dir,
    } 
}

func (p *LabelGraphRepository) Get(id string) (*LabelGraph, error) {
	val, ok := p.labelGraphs[id]
	if !ok {
		return nil, errors.New((id + " is not a valid label graph"))
	}

	return val, nil
}

func (p *LabelGraphRepository) Load() error {
	type LabelGraphMappingEntry struct {
	    Name string `json:"name"`
	}

	type LabelGraphMapping map[string]LabelGraphMappingEntry
	var labelGraphMapping LabelGraphMapping

	p.labelGraphs = make(map[string]*LabelGraph)

	//read the mapping.json file
	raw, err := ioutil.ReadFile((p.dir + "/mapping.json"))
    if err != nil {
        return err
    }
    err = json.Unmarshal(raw, &labelGraphMapping)
    if err != nil {
    	return err
    }

	err = filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := filepath.Base(path)
		if filename == "mapping.json" {
			return nil
		}


		val, ok := labelGraphMapping[filename]
		if !ok {
			return errors.New((filename + " not in label graph mapping file"))
		}


		labelGraph := NewLabelGraph(path)
		err = labelGraph.Load()
		if err != nil {
			return err
		}
		p.labelGraphs[val.Name] = labelGraph

		return nil
	})
	return err
}