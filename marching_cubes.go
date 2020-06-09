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

func generateTriangles(index int, n [8]*math32.Vector3, v [8]float32, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var l [12]*math32.Vector3
	for i := range MarchingCubesEdgeConnections {
		if MarchingCubesEdges[index]&(1<<i) != 0 {
			i0, i1 := MarchingCubesEdgeConnections[i][0], MarchingCubesEdgeConnections[i][1]
			v0, v1 := isosurface(n[i0]), isosurface(n[i1])
			l[i] = n[i0].Clone().Lerp(n[i1], (isolevel-v0)/(v1-v0))
		}
	}
	var vertices []*math32.Vector3
	triangles := MarchingCubesTriangles[index]
	for i := 0; triangles[i] != -1; i += 3 {
		a, b, c := l[triangles[i]], l[triangles[i+1]], l[triangles[i+2]]
		vertices = append(vertices, a)
		vertices = append(vertices, b)
		vertices = append(vertices, c)
	}
	return vertices
}

func marchCube(origin *math32.Vector3, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var n [8]*math32.Vector3
	var v [8]float32
	for i, offset := range MarchingCubesNeighborOffsets {
		n[i] = origin.Clone().Add(offset)
		v[i] = isosurface(n[i])
	}
	index := 0
	for i, value := range v {
		if value >= isolevel {
			index |= (1 << i)
		}
	}
	if MarchingCubesEdges[index] == 0 {
		return nil
	}
	return generateTriangles(index, n, v, isosurface, isolevel)
}

func marchCubes(size int, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var vertices []*math32.Vector3
	for i := -1; i <= size; i++ {
		for j := -1; j <= size; j++ {
			for k := -1; k <= size; k++ {
				v := &math32.Vector3{float32(i), float32(j), float32(k)}
				vertices = append(vertices, marchCube(v, isosurface, isolevel)...)
			}
		}
	}
	return vertices
}

func marchVoxels(voxels [][][]int8) []*math32.Vector3 {
	return marchCubes(len(voxels), func(v *math32.Vector3) float32 {
		x, y, z := int(v.X), int(v.Y), int(v.Z)
		if x >= 0 && x < len(voxels) && y >= 0 && y < len(voxels[x]) && z >= 0 && z < len(voxels[x][y]) {
			return float32(voxels[x][y][z])
		}
		return 0
	}, 0)
}

type MarchingCubesCase struct {
	*core.Node
	index  int
	label  *gui.Label
	labels *core.Node
	mesh   *core.Node
}

func NewMarchingCubesCase() *MarchingCubesCase {
	group := core.NewNode()

	cube := geometry.NewCube(1)
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetWireframe(true)
	group.Add(graphic.NewMesh(cube, mat))

	label := gui.NewLabel("Marching Cubes Case 0")
	label.SetFontSize(18)
	label.SetColor(math32.NewColor("White"))
	w, h := label.Size()
	width, _ := a.GetFramebufferSize()
	scaleW, _ := a.GetScale()
	label.SetPosition(float32(float64(width)/scaleW)/2-w/2, h)
	group.Add(label)

	c := &MarchingCubesCase{group, 0, label, nil, nil}
	c.Step(0)

	return c
}

func (c *MarchingCubesCase) Step(i int) int {
	c.index += i
	if c.index > 255 {
		c.index = 255
	}
	if c.index < 0 {
		c.index = 0
	}

	c.Node.Remove(c.labels)
	voxels := voxelsAtStep(uint8(c.index))

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
	c.Node.Add(c.labels)

	positions, normals, indices := fromVertices(marchCube(&math32.Vector3{0, 0, 0}, func(v *math32.Vector3) float32 {
		x, y, z := int(v.X), int(v.Y), int(v.Z)
		if x >= 0 && x < len(voxels) && y >= 0 && y < len(voxels[0]) && z >= 0 && z < len(voxels[0][0]) {
			return float32(voxels[x][y][z])
		}
		return 0
	}, 0))

	c.Node.Remove(c.mesh)
	c.mesh = NewDoubleSidedMesh(positions, normals, indices)
	c.mesh.SetPosition(-0.5, -0.5, -0.5)
	c.Node.Add(c.mesh)
	c.label.SetText(fmt.Sprintf("Marching Cubes Case %d", c.index))

	return int(c.index)
}

type MarchingCubesChunk struct {
	core.INode

	voxels [][][]int8
}

func (c *MarchingCubesChunk) HandleVoxelClick(x, y, z int, shift bool) {
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
	vertices := marchVoxels(c.voxels)
	geom := geometry.NewGeometry()
	indices := indices(len(vertices))
	positions := flatten(vertices)
	normals := computeNormals(vertices)
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mesh := c.Children()[0].(*graphic.Mesh)
	mat := mesh.GetMaterial(0)
	mesh.Init(geom, mat)
}

func NewMarchingCubesChunk(voxels [][][]int8) *MarchingCubesChunk {
	vertices := marchVoxels(voxels)
	mesh := NewMesh(vertices)
	node := core.NewNode()
	node.Add(mesh)
	node.SetPosition(-float32(len(voxels))/2+1, -float32(len(voxels[0]))/2+1, -float32(len(voxels[0][0]))/2+1)
	node.SetName("Marching Cubes")
	return &MarchingCubesChunk{node, voxels}
}
