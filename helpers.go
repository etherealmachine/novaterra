package main

import (
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

func flatten(l []*math32.Vector3) math32.ArrayF32 {
	a := math32.NewArrayF32(3*len(l), 3*len(l))
	for i, vec := range l {
		a[i*3] = vec.X
		a[i*3+1] = vec.Y
		a[i*3+2] = vec.Z
	}
	return a
}

func indices(count int) math32.ArrayU32 {
	l := make([]uint32, count)
	for i := 0; i < count; i++ {
		l[i] = uint32(i)
	}
	return math32.ArrayU32(l)
}

func computeNormals(vertices []*math32.Vector3) math32.ArrayF32 {
	normals := math32.NewArrayF32(3*len(vertices), 3*len(vertices))
	for i := 0; i < len(vertices); i += 3 {
		t := math32.NewTriangle(vertices[i], vertices[i+1], vertices[i+2])
		n := t.Normal(nil).Negate()
		for j := 0; j < 3; j++ {
			normals[i*3+j*3] = n.X
			normals[i*3+j*3+1] = n.Y
			normals[i*3+j*3+2] = n.Z
		}
	}
	return normals
}

func NewMesh(vertices []*math32.Vector3) graphic.IGraphic {
	geom := geometry.NewGeometry()
	indices := indices(len(vertices))
	positions := flatten(vertices)
	normals := computeNormals(vertices)
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewStandard(math32.NewColor("Green"))
	mat.SetSide(material.SideDouble)
	return graphic.NewMesh(geom, mat)
}

func NewWireframeMesh(vertices []*math32.Vector3) graphic.IGraphic {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetWireframe(true)
	return graphic.NewMesh(geom, mat)
}

func NewPointsMesh(vertices []*math32.Vector3) graphic.IGraphic {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	mat := material.NewPoint(math32.NewColor("White"))
	mat.SetSize(50)
	return graphic.NewPoints(geom, mat)
}

func NewLinesMesh(points []*math32.Vector3) graphic.IGraphic {
	lines := make([]float32, len(points)*3)
	for i := 0; i < len(points)/2; i++ {
		from := points[i*2]
		to := points[i*2+1]
		lines[i*6] = from.X
		lines[i*6+1] = from.Y
		lines[i*6+2] = from.Z
		lines[i*6+3] = to.X
		lines[i*6+4] = to.Y
		lines[i*6+5] = to.Z
	}
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(math32.ArrayF32(lines)).AddAttrib(gls.VertexPosition))
	mat := material.NewStandard(math32.NewColor("Red"))
	return graphic.NewLines(geom, mat)
}

func NewNormalsMesh(vertices []*math32.Vector3) graphic.IGraphic {
	lines := make([]float32, len(vertices)*2)
	for i := 0; i < len(vertices)/3; i++ {
		t := math32.NewTriangle(vertices[i*3], vertices[i*3+1], vertices[i*3+2])
		c := vertices[i*3].Clone().Add(vertices[i*3+1].Clone()).Add(vertices[i*3+2].Clone()).MultiplyScalar(1.0 / 3.0)
		n := c.Clone().Add(t.Normal(nil))
		lines[i*6] = c.X
		lines[i*6+1] = c.Y
		lines[i*6+2] = c.Z
		lines[i*6+3] = n.X
		lines[i*6+4] = n.Y
		lines[i*6+5] = n.Z
	}
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(math32.ArrayF32(lines)).AddAttrib(gls.VertexPosition))
	mat := material.NewStandard(math32.NewColor("Red"))
	return graphic.NewLines(geom, mat)
}
