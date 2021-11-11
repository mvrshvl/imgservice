package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

var graphNodes = []opts.GraphNode{
	{Name: "Node1"},
	{Name: "Node2"},
	{Name: "Node3"},
	{Name: "Node4"},
	{Name: "Node5"},
	{Name: "Node6"},
	{Name: "Node7"},
	{Name: "Node8"},
	{Name: "Node9"},
	{Name: "Node10"},
	{Name: "Node11"},
	{Name: "Node12"},
	{Name: "Node13"},
	{Name: "Node14"},
	{Name: "Node15"},
	{Name: "Node16"},
	{Name: "Node17"},
	{Name: "Node18"},
	{Name: "Node19"},
	{Name: "Node20"},
	{Name: "Node21"},
	{Name: "Node22"},
	{Name: "Node23"},
	{Name: "Node24"},
	{Name: "Node25"},
	{Name: "Node26"},
	{Name: "Node27"},
	{Name: "Node28"},
	{Name: "Node29"},
}

func genLinks() []opts.GraphLink {
	links := make([]opts.GraphLink, 0)
	for i := 0; i < len(graphNodes); i++ {
		for j := 0; j < len(graphNodes); j++ {
			links = append(links, opts.GraphLink{Source: graphNodes[i].Name, Target: graphNodes[j].Name})
		}
	}
	return links
}

func graphBase() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic graph example"}),
	)
	graph.AddSeries("graph", graphNodes, genLinks(),
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 28000}},
		),
	)
	return graph
}

func graphCircle() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Circular layout"}),
	)

	graph.AddSeries("graph", graphNodes, genLinks()).
		SetSeriesOptions(
			charts.WithGraphChartOpts(
				opts.GraphChart{
					Force:  &opts.GraphForce{Repulsion: 1000},
					Layout: "circular",
				}),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "right"}),
		)
	return graph
}

func graphNpmDep() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "npm dependencies demo",
		}))

	f, err := ioutil.ReadFile("./graph/deps/npmdepgraph.json")
	if err != nil {
		panic(err)
	}

	type Data struct {
		Nodes []opts.GraphNode
		Links []opts.GraphLink
	}

	var data Data
	if err := json.Unmarshal(f, &data); err != nil {
		fmt.Println(err)
	}

	graph.AddSeries("graph", data.Nodes, data.Links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Layout:             "none",
				Roam:               true,
				FocusNodeAdjacency: true,
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     true,
					Color:    "black",
					Position: "left",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.3,
			}),
		)
	return graph
}

type GraphExamples struct{}

func (GraphExamples) Examples() {
	page := components.NewPage()

	g := graphBase()

	page.AddCharts(
		g,
	)

	f, err := os.Create("blockchain_data/graph.html")
	if err != nil {
		panic(err)

	}

	page.Width = "1000"
	page.Render(io.MultiWriter(f))
}
