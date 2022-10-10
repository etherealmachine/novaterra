package main

import (
	"fmt"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
)

type GeometryBuilder struct {
	positions math32.ArrayF32
	indices   math32.ArrayU32
}

func (g *GeometryBuilder) CurrentTriangleIndex() uint32 {
	return uint32(len(g.positions) / 3)
}

func (g *GeometryBuilder) AddVertex(x, y, z float32) {
	g.positions = append(g.positions, x)
	g.positions = append(g.positions, y)
	g.positions = append(g.positions, z)
}

func (g *GeometryBuilder) AddTriangle(i, j, k uint32) {
	g.indices = append(g.indices, i)
	g.indices = append(g.indices, j)
	g.indices = append(g.indices, k)
}

func (g *GeometryBuilder) Build() geometry.IGeometry {
	fmt.Println(len(g.positions), len(g.indices))
	normals := make([]float32, len(g.positions))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(g.positions).AddAttrib(gls.VertexPosition))
	normals = geometry.CalculateNormals(g.indices, g.positions, normals)
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(g.indices)
	return geom
}
