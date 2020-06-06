// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"fmt"

	"github.com/g3n/engine/core"
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
		voxels[x][y+1][z+1],
		voxels[x+1][y+1][z+1],
	}
}

func generateTransvoxelMesh(x, y, z int, voxels [][][]int8) ([]float32, []float32, []uint32) {
	c := corners(x, y, z, voxels)
	index := uint8(((c[0] >> 7) & 0x01) |
		((c[1] >> 6) & 0x02) |
		((c[2] >> 5) & 0x04) |
		((c[3] >> 4) & 0x08) |
		((c[4] >> 3) & 0x10) |
		((c[5] >> 2) & 0x20) |
		((c[6] >> 1) & 0x40) |
		int8((byte(c[7]) & 0x80)))
	var positions []float32
	var normals []float32
	var indices []uint32
	var indexMap [15]uint32
	if (index ^ uint8((byte(c[7])>>7)&0xFF)) != 0 {
		cell := RegularCellData[RegularCellClass[index]]
		vertexLocations := RegularVertexData[index]
		vertexCount := cell.GetVertexCount()
		triangleCount := cell.GetTriangleCount()
		fmt.Println(vertexCount, triangleCount)
		for i := 0; i < vertexCount; i++ {
			//edge := vertexLocations[i] >> 8
			//reuseIndex := edge & 0xF               // Vertex id which should be created or reused 1, 2 or 3
			//rDir := edge >> 4                      // Direction to go to reach a previous cell for reusing
			v0 := (vertexLocations[i] >> 4) & 0x0F // First Corner Index
			v1 := (vertexLocations[i]) & 0x0F      // Second Corner Index
			d0 := c[v0]
			d1 := c[v1]
			t := int((d1 << 8) / (d1 - d0))
			u := 0x0100 - t
			t0 := float32(t) / 256
			t1 := float32(u) / 256
			if byte(t)&0x00FF != 0 {
				// Vertex lies in the interior of the edge
			} else if t == 0 {
				// Vertex lies at the higher numbered endpoint
				if v1 == 7 {
					// This cell owns the vertex
				} else {
					// Try to re-use corner vertex from a preceding cell
				}
			} else {
				// Vertex lies at the lower-numbered endpoint. Try to reuse corner vertex from a preceding cell
			}
			// Vertices at the two corners
			p0 := cornerVertices[v0]
			p1 := cornerVertices[v1]
			// Linearly interpolate along the 2 vertices to get the new vertex
			qX, qY, qZ := p0[0]*t0+t1*p1[0], p0[1]*t0+t1*p1[1], p0[2]*t0+t1*p1[2]
			positions = append(positions, qX+float32(x))
			positions = append(positions, qY+float32(y))
			positions = append(positions, qZ+float32(z))
			indexMap[i] = uint32(i)
		}
		for t := 0; t < triangleCount; t++ {
			for i := 0; i < 3; i++ {
				indices = append(indices, indexMap[cell.VertexIndex[t*3+i]])
			}
		}
		if len(indices)/3 != triangleCount {
			panic("not enough indices")
		}
		fmt.Println(positions, indices)
	}
	return positions, normals, indices
}

func marchTransvoxels(voxels [][][]int8) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	for x := 0; x < len(voxels)-1; x++ {
		for z := 0; z < len(voxels[x])-1; z++ {
			for y := 0; y < len(voxels[x][y])-1; y++ {
				p, n, i := generateTransvoxelMesh(x, y, z, voxels)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
	}
	return positions, normals, indices
}

func transvoxelCase(i uint8) core.INode {
	voxels := [][][]int8{
		{{127, 127}, {127, 127}},
		{{127, 127}, {127, 127}},
	}
	if i&1 != 0 {
		voxels[0][0][0] = -127
	}
	if (i>>1)&1 != 0 {
		voxels[1][0][0] = -127
	}
	if (i>>2)&1 != 0 {
		voxels[0][0][1] = -127
	}
	if (i>>3)&1 != 0 {
		voxels[1][0][1] = -127
	}
	if (i>>4)&1 != 0 {
		voxels[0][1][0] = -127
	}
	if (i>>5)&1 != 0 {
		voxels[1][1][0] = -127
	}
	if (i>>6)&1 != 0 {
		voxels[0][1][1] = -117
	}
	if (i>>7)&1 != 0 {
		voxels[1][1][1] = -127
	}

	group := core.NewNode()
	positions, normals, indices := generateTransvoxelMesh(0, 0, 0, voxels)
	mesh := NewFastMesh(positions, normals, indices)
	group.Add(mesh)
	group.Add(visualizeVoxels(voxels))
	group.SetName(fmt.Sprintf("Transvoxel Case %d", i))
	return group
}

func transvoxelMesh(voxels [][][]int8) core.INode {
	positions, normals, indices := marchTransvoxels(voxels)
	m := NewFastMesh(positions, normals, indices)
	group := core.NewNode()
	group.Add(m)
	group.SetPosition(-float32(len(voxels))/2+0.5, -1, -float32(len(voxels[0][0]))/2+0.5)
	group.SetName("Transvoxel Mesh")
	return group
}
