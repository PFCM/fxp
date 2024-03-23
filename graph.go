package fxp

import (
	"fmt"

	"github.com/pfcm/fxp/fix"
)

// Graph is a directed acyclic graph of Tickers. It is itself a Ticker.
type Graph struct {
	// TODO: this representation is probably bad, we should ideally
	// keep them all in a flat list (?)
	heads           []GraphNode
	inputs, outputs int
}

type GraphNode struct {
	Ticker
	Children []GraphNode
}

// TODO: only actually works for trees.
func NewGraph(n GraphNode) *Graph {
	g := &Graph{
		heads:  []GraphNode{n},
		inputs: 1,
	}
	// We have to traverse the graph to figure out the number of outputs.
	// This seems a bit silly.
	outs := 0
	todo := []GraphNode{n}
	for len(todo) != 0 {
		current := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		if len(current.Children) == 0 {
			outs += current.Outputs()
		}
		todo = append(todo, current.Children...)
	}
	g.outputs = outs

	return g
}

// Node is a convenience for defining simple subgraphs. The outputs
// of the first ticker will be connected to the provided children
// in order. Nested calls will build trees.
func Node(t Ticker, children ...GraphNode) (GraphNode, error) {
	if o := t.Outputs(); o != len(children) {
		return GraphNode{}, fmt.Errorf("%v wants %d outputs: got %v", t, o, children)
	}
	return GraphNode{
		Ticker:   t,
		Children: children,
	}, nil
}

var _ Ticker = &Graph{}

func (g *Graph) Inputs() int    { return g.inputs }
func (g *Graph) Outputs() int   { return g.outputs }
func (g *Graph) String() string { return fmt.Sprint(g.heads) }

func (g *Graph) Tick(inputs, outputs [][]fix.S17) {}
