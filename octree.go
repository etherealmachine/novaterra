package main

import (
	"fmt"
	"log"
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
}

func NewTree(parent *Node, pos math32.Vector3, size float32) *Node {
	return &Node{Position: pos, Size: size, Parent: parent}
}

func (n *Node) At(x, y, z float32) *Node {
	if n.Leaf() {
		return n
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

func (n *Node) merge() {
	mat := 0
	for _, child := range n.Children {
		if child != nil {
			child.merge()
			if mat == 0 {
				mat = child.Material
			}
			if child.Material != mat {
				return
			}
		} else {
			mat = 0
		}
	}
	if n.Material == 0 && mat != 0 {
		for i := range n.Children {
			n.Children[i] = nil
		}
		n.Material = mat
	}
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
	return fmt.Sprintf("(%.2f, %.2f, %.2f) %.2f, %d", n.Position.X, n.Position.Y, n.Position.Z, n.Size, n.Material)
}

func (n *Node) Node(mat *Material) core.INode {
	log.Println("\n" + n.String())
	n.merge()
	log.Println("\n" + n.String())
	root := core.NewNode()
	n.DFS(func(n *Node, _ int) bool {
		if n.Material > 0 {
			g := geometry.NewCube(n.Size)
			m := graphic.NewMesh(g, mat)
			m.SetPosition(n.Position.X, n.Position.Y, n.Position.Z)
			root.Add(m)
		}
		return true
	})
	return root
}
