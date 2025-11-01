package graph

import (
	"sync"
)

// Graph represents a link graph with source -> target edges
type Graph struct {
	edges map[string][]string
	mu    sync.RWMutex
}

// NewGraph creates a new Graph instance
func NewGraph() *Graph {
	return &Graph{
		edges: make(map[string][]string),
	}
}

// AddEdge adds a directed edge from source to target
func (g *Graph) AddEdge(source, target string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if edge already exists
	existing := g.edges[source]
	for _, t := range existing {
		if t == target {
			return // Edge already exists
		}
	}

	g.edges[source] = append(g.edges[source], target)
}

// AddEdges adds multiple edges from a source to multiple targets
func (g *Graph) AddEdges(source string, targets []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	existing := make(map[string]bool)
	for _, t := range g.edges[source] {
		existing[t] = true
	}

	for _, target := range targets {
		if !existing[target] {
			g.edges[source] = append(g.edges[source], target)
			existing[target] = true
		}
	}
}

// GetEdges returns all edges from a source node
func (g *Graph) GetEdges(source string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edges[source]
}

// GetAllEdges returns a map of all edges
func (g *Graph) GetAllEdges() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make(map[string][]string)
	for source, targets := range g.edges {
		result[source] = make([]string, len(targets))
		copy(result[source], targets)
	}
	return result
}

// GetEdgeList returns a flat list of edges as [source, target] pairs
func (g *Graph) GetEdgeList() [][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edgeList := make([][]string, 0)
	for source, targets := range g.edges {
		for _, target := range targets {
			edgeList = append(edgeList, []string{source, target})
		}
	}
	return edgeList
}

// NodeCount returns the number of nodes in the graph
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// EdgeCount returns the total number of edges in the graph
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, targets := range g.edges {
		count += len(targets)
	}
	return count
}

