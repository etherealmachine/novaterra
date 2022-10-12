package main

import (
	"log"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

type Node struct {
	Position math32.Vector3
	Size     float32
	Children [8]*Node
	Material int
}

func NewTree(pos math32.Vector3, size float32) *Node {
	root := &Node{Position: pos, Size: size}
	return root
}

func (n *Node) At(x, y, z float32) *Node {
	if n.Size <= 1 {
		return n
	}
	if n.Children[0] == nil {
		s := n.Size / 2
		o := n.Size / 4
		n.Children[0] = NewTree(*n.Position.Clone().Add(math32.NewVector3(-o, -o, -o)), s)
		n.Children[1] = NewTree(*n.Position.Clone().Add(math32.NewVector3(-o, -o, +o)), s)
		n.Children[2] = NewTree(*n.Position.Clone().Add(math32.NewVector3(-o, +o, -o)), s)
		n.Children[3] = NewTree(*n.Position.Clone().Add(math32.NewVector3(-o, +o, +o)), s)
		n.Children[4] = NewTree(*n.Position.Clone().Add(math32.NewVector3(+o, -o, -o)), s)
		n.Children[5] = NewTree(*n.Position.Clone().Add(math32.NewVector3(+o, -o, +o)), s)
		n.Children[6] = NewTree(*n.Position.Clone().Add(math32.NewVector3(+o, +o, -o)), s)
		n.Children[7] = NewTree(*n.Position.Clone().Add(math32.NewVector3(+o, +o, +o)), s)
	}
	for _, child := range n.Children {
		if child.Contains(x, y, z) {
			return child.At(x, y, z)
		}
	}
	return nil
}

func (n *Node) Contains(x, y, z float32) bool {
	s := n.Size / 2
	v := (n.Position.X-s <= x && x <= n.Position.X+s &&
		n.Position.Y-s <= y && y <= n.Position.Y+s &&
		n.Position.Z-s <= z && z <= n.Position.Z+s)
	return v
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
		}
	}
	if n.Material == 0 && mat != 0 {
		for i := range n.Children {
			n.Children[i] = nil
		}
		n.Material = mat
	}
}

func (n *Node) Node() core.INode {
	root := core.NewNode()
	n.merge()
	n.DFS(func(n *Node, _ int) bool {
		if n.Material != 0 {
			g := geometry.NewCube(n.Size)
			mat := material.NewStandard(math32.NewColor("white"))
			mat.AddTexture(textures["dirt"])
			mat.SetShader("terrain")
			m := graphic.NewMesh(g, mat)
			m.SetPositionVec(&n.Position)
			root.Add(m)
		}
		return true
	})
	log.Println(len(root.Children()))
	return root
}
