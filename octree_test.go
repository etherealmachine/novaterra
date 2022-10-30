package main

import (
	"testing"

	"github.com/g3n/engine/math32"
)

func TestMerge(t *testing.T) {
	tree := NewTree(nil, math32.Vector3{X: 4, Y: 4, Z: 4}, 8)
	for _, p := range [][3]float32{
		{0.5, 0.5, 0.5},
		{1.5, 0.5, 0.5},
		{0.5, 0.5, 1.5},
		{1.5, 0.5, 1.5},
		{0.5, 1.5, 0.5},
		{1.5, 1.5, 0.5},
		{0.5, 1.5, 1.5},
		{1.5, 1.5, 1.5},
		{7.5, 7.5, 7.5},
	} {
		node := tree.At(p[0], p[1], p[2])
		if node == nil {
			t.Fatalf("got nil node at (%.2f, %.2f, %.2f)", p[0], p[1], p[2])
		}
		node.Material = 1
		node.Density = 1
	}
	count := 0
	tree.DFS(func(n *Node, i int) bool {
		if n.Material > 0 && n.Density > 0 {
			count++
		}
		return true
	})
	if count != 9 {
		t.Fatalf("tree has wrong number of nodes, got %d, want 9", count)
	}
	tree.merge()
	count = 0
	tree.DFS(func(n *Node, i int) bool {
		if n.Material > 0 && n.Density > 0 {
			count++
		}
		return true
	})
	if count != 2 {
		t.Fatalf("tree has wrong number of nodes, got %d, want 2", count)
	}
}
