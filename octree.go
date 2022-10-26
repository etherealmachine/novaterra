package main

import (
	"fmt"
	"strings"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/math32"
)

type Node struct {
	Position math32.Vector3
	Size     float32
	Parent   *Node
	Children [8]*Node
	Material int
	Density  float32
}

func NewTree(parent *Node, pos math32.Vector3, size float32) *Node {
	return &Node{Position: pos, Size: size, Parent: parent}
}

func (n *Node) At(x, y, z float32) *Node {
	if n.Leaf() && n.Contains(x, y, z) {
		return n
	}
	if n.Parent != nil && !n.Contains(x, y, z) {
		return n.Parent.At(x, y, z)
	}
	s := n.Size / 2
	o := n.Size / 4
	for i, offset := range [8][3]float32{
		{-o, -o, -o},
		{-o, -o, +o},
		{-o, +o, -o},
		{-o, +o, +o},
		{+o, -o, -o},
		{+o, -o, +o},
		{+o, +o, -o},
		{+o, +o, +o},
	} {
		cx, cy, cz := n.Position.X+offset[0], n.Position.Y+offset[1], n.Position.Z+offset[2]
		if contains(x, y, z, cx, cy, cz, s) {
			if n.Children[i] == nil {
				n.Children[i] = NewTree(n, *math32.NewVector3(cx, cy, cz), s)
			}
			return n.Children[i].At(x, y, z)
		}
	}
	return nil
}

func (n *Node) Leaf() bool {
	return n.Size <= 1
}

func (n *Node) Contains(x, y, z float32) bool {
	return contains(x, y, z, n.Position.X, n.Position.Y, n.Position.Z, n.Size)
}

func contains(x, y, z, x1, y1, z1, s float32) bool {
	o := s / 2
	return x1-o <= x && x <= x1+o &&
		y1-o <= y && y <= y1+o &&
		z1-o <= z && z <= z1+o
}

func (n *Node) DFS(fn func(*Node, int) bool) {
	n.dfs(fn, 0)
}

func (n *Node) dfs(fn func(*Node, int) bool, depth int) {
	if n == nil || !fn(n, depth) {
		return
	}
	for _, child := range n.Children {
		if child != nil {
			child.dfs(fn, depth+1)
		}
	}
}

// prune removes empty children from a node
func (n *Node) prune() {
	for i, child := range n.Children {
		if child.empty() {
			return
		}
		n.Children[i] = nil
	}
}

// A node is empty if it has no material or zero density
// nil nodes are trivially empty
func (n *Node) empty() bool {
	return n == nil || n.Material == 0 || n.Density == 0
}

// A node can be merged if all its children have the same material and non-zero density
// The node has the material of its children and the average of the densities
func (n *Node) merge() {
	if n == nil || n.Leaf() {
		return
	}
	n.prune()
	var density float32
	var mat *int
	sameMaterial := false
	for _, child := range n.Children {
		child.merge()
		if !child.empty() {
			density += child.Density
			if mat == nil {
				mat = &child.Material
				sameMaterial = true
			} else if *mat != child.Material {
				sameMaterial = false
			}
		} else {
			sameMaterial = false
		}
	}
	if mat != nil && sameMaterial {
		n.Material = *mat
		n.Density = density / 8
		for i := range n.Children {
			n.Children[i] = nil
		}
	}
}

func (n *Node) Clone() *Node {
	return n.clone(nil)
}

func (n *Node) clone(parent *Node) *Node {
	if n == nil {
		return nil
	}
	clone := &Node{
		Position: *n.Position.Clone(),
		Size:     n.Size,
		Parent:   parent,
		Material: n.Material,
		Density:  n.Density,
	}
	for i, child := range n.Children {
		clone.Children[i] = child.clone(clone)
	}
	return clone
}

func (n *Node) String() string {
	b := new(strings.Builder)
	n.DFS(func(n *Node, depth int) bool {
		for ; depth > 0; depth-- {
			b.WriteString(" ")
		}
		b.WriteString(n.string())
		b.WriteString("\n")
		return true
	})
	return b.String()
}

func (n *Node) string() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f) %.2f, %d, %.2f", n.Position.X, n.Position.Y, n.Position.Z, n.Size, n.Material, n.Density)
}

func (n *Node) DualContourMesh(mat *Material) core.INode {
	b := new(GeometryBuilder)
	n.merge()
	n.DFS(func(n *Node, _ int) bool {
		if !n.empty() {
			s := n.Size
			h := s / 2
			neighbor := n.At(n.Position.X, n.Position.Y+s, n.Position.Z)
			if neighbor.empty() {
				i := b.CurrentTriangleIndex()
				b.AddVertex(n.Position.X-h, n.Position.Y+s-h, n.Position.Z-h)
				b.AddVertex(n.Position.X+s-h, n.Position.Y+s-h, n.Position.Z-h)
				b.AddVertex(n.Position.X-h, n.Position.Y+s-h, n.Position.Z+s-h)
				b.AddVertex(n.Position.X+s-h, n.Position.Y+s-h, n.Position.Z+s-h)
				b.AddTriangle(i+0, i+2, i+1)
				b.AddTriangle(i+1, i+2, i+3)
			}
			return false
		}
		return true
	})
	root := core.NewNode()
	g := b.Build()
	m := graphic.NewMesh(g, mat)
	root.Add(m)
	return root
}

func (n *Node) NaiveVoxelMesh(mat *Material) core.INode {
	root := core.NewNode()
	n.DFS(func(n *Node, _ int) bool {
		if !n.empty() {
			g := geometry.NewCube(n.Size)
			m := graphic.NewMesh(g, mat)
			m.SetPositionVec(&n.Position)
			root.Add(m)
			return false
		}
		return true
	})
	return root
}

func (n *Node) MergedVoxelMesh(mat *Material) core.INode {
	n.merge()
	return n.NaiveVoxelMesh(mat)
}
