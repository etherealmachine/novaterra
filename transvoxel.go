// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"fmt"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
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

func transvoxelCase(x, y, z int, c [8]int8) uint8 {
	return uint8(((c[0] >> 7) & 0x01) |
		((c[1] >> 6) & 0x02) |
		((c[2] >> 5) & 0x04) |
		((c[3] >> 4) & 0x08) |
		((c[4] >> 3) & 0x10) |
		((c[5] >> 2) & 0x20) |
		((c[6] >> 1) & 0x40) |
		int8((byte(c[7]) & 0x80)))
}

func generateTransvoxelMesh(x, y, z, N, M, L int, voxels [][][]int8, maxIndex uint32, prevVertexCache [][][]int32, currVertexCache [][][]int32) ([]float32, []float32, []uint32) {
	c := corners(x, y, z, voxels)
	index := transvoxelCase(x, y, z, c)
	if (index ^ uint8((byte(c[7])>>7)&0xFF)) == 0 {
		return nil, nil, nil
	}
	var directionMask byte
	if x > 0 {
		directionMask |= 1
	}
	if y > 0 {
		directionMask |= 2
	}
	if z > 0 {
		directionMask |= 4
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

		edge := vertexLocations[i] >> 8
		reuseIndex := int(edge & 0xF)          // vertex id which should be created or reused: 1, 2 or 3
		rDir := byte(edge >> 4)                // direction to go to reach the cell to reusing
		v0 := (vertexLocations[i] >> 4) & 0x0F // First Corner Index
		v1 := (vertexLocations[i]) & 0x0F      // Second Corner Index
		d0 := int16(c[v0])
		d1 := int16(c[v1])
		t := d1 << 8 / (d1 - d0)

		index := int32(-1)
		if currVertexCache != nil && prevVertexCache != nil && v1 != 7 && rDir&directionMask == rDir {
			rx, ry, rz := x, y, z
			if rDir&1 != 0 {
				rx--
			}
			if rDir&2 != 0 {
				rz--
			}
			if rDir&4 != 0 {
				ry--
			}
			if ry == y {
				index = currVertexCache[rx][rz][reuseIndex]
			} else {
				index = prevVertexCache[rx][rz][reuseIndex]
			}
		}

		if index == -1 {
			u := float32(t) / 256
			v := 1 - u
			// Vertices at the two corners
			p0 := cornerVertices[v0]
			p1 := cornerVertices[v1]
			// Normals at the two corners
			n0 := cornerGradient[v0]
			n1 := cornerGradient[v1]
			// Linearly interpolate along the 2 vertices to get the new vertex
			qX, qY, qZ := p0[0]*u+v*p1[0], p0[1]*u+v*p1[1], p0[2]*u+v*p1[2]
			index = int32(maxIndex + uint32(len(positions)/3))
			positions = append(positions, qX+float32(x))
			positions = append(positions, qY+float32(y))
			positions = append(positions, qZ+float32(z))
			nX, nY, nZ := n0[0]*u+v*n1[0], n0[1]*u+v*n1[1], n0[2]*u+v*n1[2]
			norm := math32.Sqrt(nX*nX + nY*nY + nZ*nZ)
			if norm == 0 {
				norm = 1
			}
			normals = append(normals, nX/norm)
			normals = append(normals, nY/norm)
			normals = append(normals, nZ/norm)
			if currVertexCache != nil && rDir&8 != 0 {
				currVertexCache[x][z][reuseIndex] = index
			}
		}
		indexMap[i] = uint32(index)
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
	voxels = inflate(voxels)
	N, M, L := len(voxels), len(voxels[0]), len(voxels[0][0])
	prevVertexCache := make([][][]int32, N-1)
	currVertexCache := make([][][]int32, N-1)
	for x := 0; x < N-1; x++ {
		prevVertexCache[x] = make([][]int32, L-1)
		currVertexCache[x] = make([][]int32, L-1)
		for z := 0; z < L-1; z++ {
			prevVertexCache[x][z] = make([]int32, 4)
			currVertexCache[x][z] = make([]int32, 4)
			for i := 0; i < 4; i++ {
				prevVertexCache[x][z][i] = -1
				currVertexCache[x][z][i] = -1
			}
		}
	}
	for y := 0; y < M-1; y++ {
		for x := 0; x < N-1; x++ {
			for z := 0; z < L-1; z++ {
				p, n, i := generateTransvoxelMesh(x, y, z, N, M, L, voxels, uint32(len(positions)/3), prevVertexCache, currVertexCache)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
		prevVertexCache, currVertexCache = currVertexCache, prevVertexCache
	}
	return positions, normals, indices
}

type TransvoxelCase struct {
	*core.Node
	index  int
	label  *gui.Label
	labels *core.Node
	mesh   *core.Node
}

func NewTransvoxelCase() *TransvoxelCase {
	group := core.NewNode()

	cube := geometry.NewCube(1)
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetWireframe(true)
	group.Add(graphic.NewMesh(cube, mat))

	label := gui.NewLabel("Transvoxel Case 0")
	label.SetFontSize(18)
	label.SetColor(math32.NewColor("White"))
	w, h := label.Size()
	width, _ := a.GetFramebufferSize()
	scaleW, _ := a.GetScale()
	label.SetPosition(float32(float64(width)/scaleW)/2-w/2, h)
	group.Add(label)

	c := &TransvoxelCase{group, 0, label, nil, nil}
	c.Step(0)
	return c
}

func (c *TransvoxelCase) setVoxels(voxels [][][]int8) {
	c.Node.Remove(c.labels)
	c.labels = core.NewNode()
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for z := 0; z < 2; z++ {
				label := NewSpriteLabel(fmt.Sprintf("%d", voxels[x][y][z]))
				label.SetPosition(float32(x)-0.5, float32(y)-0.5, float32(z)-0.5)
				c.labels.Add(label)
			}
		}
	}

	c.Node.Remove(c.mesh)
	N, M, L := len(voxels), len(voxels[0]), len(voxels[0][0])
	positions, normals, indices := generateTransvoxelMesh(1, 1, 1, N, M, L, inflate(voxels), 0, nil, nil)
	c.mesh = NewDoubleSidedMesh(positions, normals, indices)
	c.mesh.SetPosition(-1.5, -1.5, -1.5)
	c.Node.Add(c.mesh)
	c.label.SetText(fmt.Sprintf("Transvoxel Case %d", c.index))
}

func (c *TransvoxelCase) Step(i int) int {
	c.index += i
	if c.index > 255 {
		c.index = 255
	}
	if c.index < 0 {
		c.index = 0
	}

	c.setVoxels(voxelsAtStep(uint8(c.index)))

	return c.index
}

type TransvoxelChunk struct {
	*core.Node

	voxels [][][]int8
	pos    int
	cell   *TransvoxelCase
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

func (c *TransvoxelChunk) Step(i int) int {
	c.pos += i
	N, M, L := len(c.voxels), len(c.voxels[0]), len(c.voxels[0][0])
	N--
	M--
	L--
	if c.pos >= N*M*L {
		c.pos = 0
	}
	x, z, y := c.pos%N, c.pos/N%L, c.pos/(N*M)
	for {
		cnrs := corners(x, y, z, c.voxels)
		if transvoxelCase(x, y, z, cnrs)^uint8((byte(cnrs[7])>>7)&0xFF) != 0 {
			break
		}
		c.pos += i
		if c.pos >= N*M*L {
			c.pos = 0
		}
		x, z, y = c.pos%N, c.pos/N%L, c.pos/(N*M)
	}
	c.cell.setVoxels([][][]int8{
		{
			{c.voxels[x][y][z], c.voxels[x][y][z+1]},
			{c.voxels[x][y+1][z], c.voxels[x][y+1][z+1]},
		},
		{
			{c.voxels[x+1][y][z], c.voxels[x+1][y][z+1]},
			{c.voxels[x+1][y+1][z], c.voxels[x+1][y+1][z+1]},
		},
	})
	c.cell.SetPosition(float32(x)+1.5, float32(y)+1.5, float32(z)+1.5)
	c.SetPosition(-float32(x), -float32(y), -float32(z))
	return c.pos
}

func NewTransvoxelChunk(voxels [][][]int8) *TransvoxelChunk {

	voxels = inflate(voxels)
	N, M, L := len(voxels), len(voxels[0]), len(voxels[0][0])
	positions, normals, indices := marchTransvoxels(voxels)
	m := NewFastMesh(positions, normals, indices)
	group := core.NewNode()
	group.Add(m)
	m.GetMaterial(0).(*material.Standard).SetOpacity(0.3)
	group.SetPosition(-float32(N)/2-1, -float32(M)/2-1, -float32(L)/2-1)
	group.SetName("Transvoxel")

	cell := NewTransvoxelCase()
	group.Add(cell)

	c := &TransvoxelChunk{group, voxels, 0, cell}
	c.Step(1)
	return c
}
