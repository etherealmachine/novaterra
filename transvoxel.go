// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"github.com/g3n/engine/math32"
)

func corners(x, y, z int, voxels [19][19][19]int8) [8]int8 {
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

func generateTransvoxelMesh(x, y, z int, voxels [19][19][19]int8, maxIndex uint32, prevVertexCache [16][16][4]int32, currVertexCache [16][16][4]int32) ([]float32, []float32, []uint32) {
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
			p := NeighborOffsets[i]
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
		if v1 != 7 && rDir&directionMask == rDir {
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
			p0 := NeighborOffsets[v0]
			p1 := NeighborOffsets[v1]
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
			if rDir&8 != 0 {
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

func marchTransvoxels(voxels [19][19][19]int8) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	var prevVertexCache [16][16][4]int32
	var currVertexCache [16][16][4]int32
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for i := 0; i < 4; i++ {
				prevVertexCache[x][z][i] = -1
				currVertexCache[x][z][i] = -1
			}
		}
	}
	for y := 1; y < 16; y++ {
		for x := 1; x < 16; x++ {
			for z := 1; z < 16; z++ {
				p, n, i := generateTransvoxelMesh(x, y, z, voxels, uint32(len(positions)/3), prevVertexCache, currVertexCache)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
		prevVertexCache, currVertexCache = currVertexCache, prevVertexCache
	}
	return positions, normals, indices
}
