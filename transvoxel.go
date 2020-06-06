// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"log"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
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
	if (index ^ uint8((byte(c[7])>>7)&0xFF)) != 0 {
		var cornerNormals [8][3]float32
		log.Printf("computing corner normals for %d, %d, %d", x, y, z)
		for i := 0; i < 8; i++ {
			p := cornerVertices[i]
			ox, oy, oz := int(p[0]), int(p[1]), int(p[2])
			cx, cy, cz := x+ox, y+oy, z+oz
			log.Printf(" need density at %d, %d, %d", cx, cy, cz)
			nx := float32(voxels[cx+1][cy][cz]-voxels[cx-1][cy][cz]) * 0.5
			ny := float32(voxels[cx][cy+1][cz]-voxels[cx][cy-1][cz]) * 0.5
			nz := float32(voxels[cx][cy][cz+1]-voxels[cx][cy][cz-1]) * 0.5
			cornerNormals[i][0] = nx
			cornerNormals[i][1] = ny
			cornerNormals[i][2] = nz
		}
		cell := RegularCellData[RegularCellClass[index]]
		vertexLocations := RegularVertexData[index]
		vertexCount := cell.GetVertexCount()
		triangleCount := cell.GetTriangleCount()
		log.Printf("%d vertices and %d triangles", vertexCount, triangleCount)
		for i := 0; i < vertexCount; i++ {
			edge := vertexLocations[i] >> 8
			reuseIndex := edge & 0xF               // Vertex id which should be created or reused 1, 2 or 3
			rDir := edge >> 4                      // Direction to go to reach a previous cell for reusing
			v0 := (vertexLocations[i] >> 4) & 0x0F // First Corner Index
			v1 := (vertexLocations[i]) & 0x0F      // Second Corner Index
			d0 := c[v0]
			d1 := c[v1]
			t := float32(d1) / float32(d1-d0)
			log.Printf("vertex %d, reuse %d, direction %d", i, reuseIndex, rDir)
			if byte(t)&0x00FF != 0 {
				log.Println(" Vertex lies in the interior of the edge")
			} else if t == 0 {
				log.Println(" Vertex lies at the higher numbered endpoint")
				if v1 == 7 {
					log.Println("  This cell owns the vertex")
				} else {
					log.Println("  Try to re-use corner vertex from a preceding cell")
				}
			} else {
				log.Println(" Vertex lies at the lower-numbered endpoint. Try to reuse corner vertex from a preceding cell")
			}
			// Vertices at the two corners
			p0 := cornerVertices[v0]
			p1 := cornerVertices[v1]
			// Normals at the two corners
			n0 := cornerNormals[v0]
			n1 := cornerNormals[v1]
			// Linearly interpolate along the 2 vertices to get the new vertex
			qX, qY, qZ := p0[0]*t+(1-t)*p1[0], p0[1]*t+(1-t)*p1[1], p0[2]*t+(1-t)*p1[2]
			nX, nY, nZ := n0[0]*t+(1-t)*n1[0], n0[1]*t+(1-t)*n1[1], n0[2]*t+(1-t)*n1[2]
			positions = append(positions, qX)
			normals = append(normals, nX)
			indices = append(indices, uint32(len(positions)-1))
			positions = append(positions, qY)
			normals = append(normals, nY)
			indices = append(indices, uint32(len(positions)-1))
			positions = append(positions, qZ)
			normals = append(normals, nZ)
			indices = append(indices, uint32(len(positions)-1))
		}
	}
	return positions, normals, indices
}

func marchTransvoxels(voxels [][][]int8) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	n, m, l := len(voxels)-3, len(voxels[0])-3, len(voxels[0][0])-3
	for x := 0; x < n; x++ {
		for z := 0; z < l; z++ {
			for y := 0; y < m; y++ {
				p, n, i := generateTransvoxelMesh(x, y, z, voxels)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
	}
	return positions, normals, indices
}

func transvoxelCase(i int) *graphic.Mesh {
	voxels := [][][]int8{
		{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
		{{1, 1, 1}, {1, -1, 1}, {1, 1, 1}},
		{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
	}
	positions, normals, indices := generateTransvoxelMesh(1, 1, 1, voxels)
	return NewFastMesh(positions, normals, indices).(*graphic.Mesh)
}

func transvoxelize(voxels [][][]int8) *core.Node {
	return core.NewNode()
}
