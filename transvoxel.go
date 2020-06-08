// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"fmt"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

var cornerVertices = [8][3]float32{
	{0, 0, 0},
	{1, 0, 0},
	{0, 0, 1},
	{1, 0, 1},
	{0, 1, 0},
	{1, 1, 0},
	{0, 1, 1},
	{1, 1, 1},
}

func corners(x, y, z int, voxels [][][]int8) [8]int8 {
	return [8]int8{
		voxels[x][y][z],
		voxels[x+1][y][z],
		voxels[x][y][z+1],
		voxels[x+1][y][z+1],
		voxels[x][y+1][z],
		voxels[x+1][y+1][z],
		voxels[x][y+1][z+1],
		voxels[x+1][y+1][z+1],
	}
}

func generateTransvoxelMesh(x, y, z int, voxels [][][]int8, maxIndex uint32) ([]float32, []float32, []uint32) {
	c := corners(x, y, z, voxels)
	index := uint8(((c[0] >> 7) & 0x01) |
		((c[1] >> 6) & 0x02) |
		((c[2] >> 5) & 0x04) |
		((c[3] >> 4) & 0x08) |
		((c[4] >> 3) & 0x10) |
		((c[5] >> 2) & 0x20) |
		((c[6] >> 1) & 0x40) |
		int8((byte(c[7]) & 0x80)))
	if (index ^ uint8((byte(c[7])>>7)&0xFF)) == 0 {
		return nil, nil, nil
	}
	var positions []float32
	var normals []float32
	var indices []uint32
	var indexMap [15]uint32
	cell := RegularCellData[RegularCellClass[index]]
	vertexLocations := RegularVertexData[index]
	vertexCount := cell.GetVertexCount()
	triangleCount := cell.GetTriangleCount()
	for i := 0; i < vertexCount; i++ {
		var cornerGradient [8][3]float32
		for i := 0; i < 8; i++ {
			p := cornerVertices[i]
			ox, oy, oz := int(p[0]), int(p[1]), int(p[2])
			cx, cy, cz := x+ox, y+oy, z+oz
			nx := (float32(voxels[cx+1][cy][cz]) - float32(voxels[cx-1][cy][cz])) * 0.5
			ny := (float32(voxels[cx][cy+1][cz]) - float32(voxels[cx][cy-1][cz])) * 0.5
			nz := (float32(voxels[cx][cy][cz+1]) - float32(voxels[cx][cy][cz-1])) * 0.5
			cornerGradient[i][0] = nx
			cornerGradient[i][1] = ny
			cornerGradient[i][2] = nz
		}

		v0 := (vertexLocations[i] >> 4) & 0x0F // First Corner Index
		v1 := (vertexLocations[i]) & 0x0F      // Second Corner Index
		d0 := float32(c[v0])
		d1 := float32(c[v1])
		t := d1 / (d1 - d0)
		// Vertices at the two corners
		p0 := cornerVertices[v0]
		p1 := cornerVertices[v1]
		// Normals at the two corners
		n0 := cornerGradient[v0]
		n1 := cornerGradient[v1]
		// Linearly interpolate along the 2 vertices to get the new vertex
		qX, qY, qZ := p0[0]*t+(1-t)*p1[0], p0[1]*t+(1-t)*p1[1], p0[2]*t+(1-t)*p1[2]
		indexMap[i] = maxIndex + uint32(len(positions)/3)
		positions = append(positions, qX+float32(x))
		positions = append(positions, qY+float32(y))
		positions = append(positions, qZ+float32(z))
		nX, nY, nZ := n0[0]*t+(1-t)*n1[0], n0[1]*t+(1-t)*n1[1], n0[2]*t+(1-t)*n1[2]
		norm := math32.Sqrt(nX*nX + nY*nY + nZ*nZ)
		if norm == 0 {
			norm = 1
		}
		normals = append(normals, nX/norm)
		normals = append(normals, nY/norm)
		normals = append(normals, nZ/norm)
	}
	for t := 0; t < triangleCount; t++ {
		for i := 0; i < 3; i++ {
			indices = append(indices, indexMap[cell.VertexIndex[t*3+(2-i)]])
		}
	}
	return positions, normals, indices
}

func marchTransvoxels(voxels [][][]int8) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	n, m, l := len(voxels), len(voxels[0]), len(voxels[0][0])
	voxels = inflate(voxels)
	for x := 1; x < n; x++ {
		for z := 1; z < m; z++ {
			for y := 1; y < l; y++ {
				p, n, i := generateTransvoxelMesh(x, y, z, voxels, uint32(len(positions)/3))
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
	}
	return positions, normals, indices
}

type TransvoxelCase struct {
	core.INode
}

func NewTransvoxelCase() *TransvoxelCase {
	group := core.NewNode()

	cube := geometry.NewCube(1)
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetWireframe(true)
	group.Add(graphic.NewMesh(cube, mat))

	c := &TransvoxelCase{group}
	c.Step(0)
	c.SetName("Transvoxel Case 0")
	return c
}

func (c *TransvoxelCase) Step(i uint8) {
	oldMesh := c.Children()[0].GetNode().FindPath("/Mesh")
	if oldMesh != nil {
		c.Children()[0].GetNode().Remove(oldMesh)
	}
	oldLabels := c.Children()[0].GetNode().FindPath("/Labels")
	if oldLabels != nil {
		c.Children()[0].GetNode().Remove(oldLabels)
	}
	voxels := voxelsAtStep(i)

	labels := core.NewNode()
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for z := 0; z < 2; z++ {
				label := NewSpriteLabel(fmt.Sprintf("%d", voxels[x][y][z]))
				label.SetPosition(float32(x)-0.5, float32(y)-0.5, float32(z)-0.5)
				labels.Add(label)
			}
		}
	}
	labels.SetName("Labels")
	c.Children()[0].GetNode().Add(labels)

	positions, normals, indices := generateTransvoxelMesh(1, 1, 1, inflate(voxels), 0)
	mesh := NewDoubleSidedMesh(positions, normals, indices)
	mesh.SetName("Mesh")
	mesh.SetPosition(-0.5, -0.5, -0.5)
	c.Children()[0].GetNode().Add(mesh)
	c.SetName(fmt.Sprintf("Transvoxel Case %d", i))
}

type TransvoxelChunk struct {
	core.INode

	voxels [][][]int8
}

func (c *TransvoxelChunk) HandleVoxelClick(x, y, z int, shift bool) {
	n, m, l := len(c.voxels), len(c.voxels[0]), len(c.voxels[0][0])
	for ox := -1; ox <= 1; ox++ {
		for oy := -1; oy <= 1; oy++ {
			for oz := -1; oz <= 1; oz++ {
				cx, cy, cz := x+ox, y+oy, z+oz
				if cx >= 0 && cx < n && cy >= 0 && cy < m && cz >= 0 && cz < l {
					if shift && c.voxels[cx][cy][cz] < 0 {
						c.voxels[cx][cy][cz]++
					} else if !shift && c.voxels[cx][cy][cz] > -127 {
						c.voxels[cx][cy][cz]--
					}
				}
			}
		}
	}
	positions, normals, indices := marchTransvoxels(inflate(c.voxels))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mesh := c.Children()[0].(*graphic.Mesh)
	mat := mesh.GetMaterial(0)
	mesh.Init(geom, mat)
}

func NewTransvoxelChunk(voxels [][][]int8) *TransvoxelChunk {
	positions, normals, indices := marchTransvoxels(inflate(voxels))
	m := NewFastMesh(positions, normals, indices)
	group := core.NewNode()
	group.Add(m)
	group.SetPosition(-float32(len(voxels))/2-1, -float32(len(voxels[0]))/2-1, -float32(len(voxels[0][0]))/2-1)
	group.SetName("Transvoxel")
	return &TransvoxelChunk{group, voxels}
}
