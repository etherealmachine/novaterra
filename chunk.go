package main

import (
	"math/rand"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

const (
	Empty = iota
	Rock
	Dirt
	Grass
	Water
)

const ChunkSize = 32

type Chunk struct {
	data [ChunkSize][ChunkSize][ChunkSize]int
}

func NewChunk() *Chunk {
	c := &Chunk{}
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				if y >= 2+rand.Intn(2) || (y > 0 && c.data[x][y-1][z] == 0) {
					continue
				}
				if y == 0 {
					c.data[x][y][z] = Rock
				} else if y < 3 {
					c.data[x][y][z] = Dirt
				} else {
					c.data[x][y][z] = Grass
				}
				if y == 3 && rand.Float32() < 0.2 {
					c.data[x][y][z] = Water
				}
			}
		}
	}
	return c
}

func (c *Chunk) SimpleGeom() geometry.IGeometry {
	var positions []float32
	var uvs []float32
	var indices []uint32
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				if c.data[x][y][z] != 0 {
					xf := float32(x)
					yf := float32(y)
					zf := float32(z)

					i := uint32(len(positions) / 3)
					// 0, 0, 0
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf)
					// 0, 0, 1
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					// 1, 0, 0
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf)
					// 1, 0, 1
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					// 0, 1, 0
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					// 0, 1, 1
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)
					// 1, 1, 0
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					// 1, 1, 1
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)

					// bottom
					indices = append(indices, i+0)
					indices = append(indices, i+2)
					indices = append(indices, i+1)
					indices = append(indices, i+1)
					indices = append(indices, i+2)
					indices = append(indices, i+3)

					// top
					indices = append(indices, i+4)
					indices = append(indices, i+5)
					indices = append(indices, i+6)
					indices = append(indices, i+6)
					indices = append(indices, i+5)
					indices = append(indices, i+7)

					// front
					indices = append(indices, i+5)
					indices = append(indices, i+1)
					indices = append(indices, i+7)
					indices = append(indices, i+7)
					indices = append(indices, i+1)
					indices = append(indices, i+3)

					// back
					indices = append(indices, i+4)
					indices = append(indices, i+6)
					indices = append(indices, i+0)
					indices = append(indices, i+0)
					indices = append(indices, i+6)
					indices = append(indices, i+2)

					// left
					indices = append(indices, i+0)
					indices = append(indices, i+1)
					indices = append(indices, i+4)
					indices = append(indices, i+4)
					indices = append(indices, i+1)
					indices = append(indices, i+5)

					// right
					indices = append(indices, i+3)
					indices = append(indices, i+2)
					indices = append(indices, i+7)
					indices = append(indices, i+7)
					indices = append(indices, i+2)
					indices = append(indices, i+6)
				}
			}
		}
	}
	normals := make([]float32, len(positions))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	normals = geometry.CalculateNormals(indices, positions, normals)
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	geom.SetIndices(indices)
	return geom
}

func (c *Chunk) interior(x, y, z int) bool {
	return z < 31 && c.data[x][y][z+1] != 0 &&
		z > 0 && c.data[x][y][z-1] != 0 &&
		y < 31 && c.data[x][y+1][z] != 0 &&
		y > 0 && c.data[x][y-1][z] != 0 &&
		x < 31 && c.data[x+1][y][z] != 0 &&
		x > 0 && c.data[x-1][y][z] != 0
}

func (c *Chunk) CulledGeom() geometry.IGeometry {
	var positions []float32
	var uvs []float32
	var indices []uint32
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				if c.data[x][y][z] != 0 && !c.interior(x, y, z) {
					xf := float32(x)
					yf := float32(y)
					zf := float32(z)

					i := uint32(len(positions) / 3)
					// 0, 0, 0
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf)
					// 0, 0, 1
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					// 1, 0, 0
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf)
					// 1, 0, 1
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					// 0, 1, 0
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					// 0, 1, 1
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)
					// 1, 1, 0
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					// 1, 1, 1
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)

					// bottom
					indices = append(indices, i+0)
					indices = append(indices, i+2)
					indices = append(indices, i+1)
					indices = append(indices, i+1)
					indices = append(indices, i+2)
					indices = append(indices, i+3)

					// top
					indices = append(indices, i+4)
					indices = append(indices, i+5)
					indices = append(indices, i+6)
					indices = append(indices, i+6)
					indices = append(indices, i+5)
					indices = append(indices, i+7)

					// front
					indices = append(indices, i+5)
					indices = append(indices, i+1)
					indices = append(indices, i+7)
					indices = append(indices, i+7)
					indices = append(indices, i+1)
					indices = append(indices, i+3)

					// back
					indices = append(indices, i+4)
					indices = append(indices, i+6)
					indices = append(indices, i+0)
					indices = append(indices, i+0)
					indices = append(indices, i+6)
					indices = append(indices, i+2)

					// left
					indices = append(indices, i+0)
					indices = append(indices, i+1)
					indices = append(indices, i+4)
					indices = append(indices, i+4)
					indices = append(indices, i+1)
					indices = append(indices, i+5)

					// right
					indices = append(indices, i+3)
					indices = append(indices, i+2)
					indices = append(indices, i+7)
					indices = append(indices, i+7)
					indices = append(indices, i+2)
					indices = append(indices, i+6)
				}
			}
		}
	}
	normals := make([]float32, len(positions))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	normals = geometry.CalculateNormals(indices, positions, normals)
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	geom.SetIndices(indices)
	return geom
}

func (c *Chunk) GreedyGeom() geometry.IGeometry {
	// Sweep over each axis (X, Y and Z)
	for d := 0; d < 3; d++ {
		u := (d + 1) % 3
		v := (d + 2) % 3
		var x [3]int
		var q [3]int

		var mask [ChunkSize * ChunkSize]bool
		q[d] = 1

		// Check each slice of the chunk one at a time
		for x[d] = -1; x[d] < ChunkSize; {
			// Compute the mask
			n := 0
			for x[v] = 0; x[v] < ChunkSize; x[v]++ {
				for x[u] = 0; x[u] < ChunkSize; x[u]++ {
					// q determines the direction (X, Y or Z) that we are searching
					blockCurrent := c.data[x[0]][x[1]][x[2]]
					blockCompare := c.data[x[0]+q[0]][x[1]+q[1]][x[2]+q[2]]
					// The mask is set to true if there is a visible face between two blocks,
					//   i.e. both aren't empty and both aren't blocks
					mask[n] = blockCurrent != blockCompare
					n++
				}
			}

			x[d]++

			n = 0

			// Generate a mesh from the mask using lexicographic ordering,
			//   by looping over each block in this slice of the chunk
			for j := 0; j < ChunkSize; j++ {
				for i := 0; i < ChunkSize; {
					if mask[n] {
						w, h := 1, 1
						// Compute the width of this quad and store it in w
						//   This is done by searching along the current axis until mask[n + w] is false
						for ; i+w < ChunkSize && mask[n+w]; w++ {
						}

						// Compute the height of this quad and store it in h
						//   This is done by checking if every block next to this row (range 0 to w) is also part of the mask.
						//   For example, if w is 5 we currently have a quad of dimensions 1 x 5. To reduce triangle count,
						//   greedy meshing will attempt to expand this quad out to CHUNK_SIZE x 5, but will stop if it reaches a hole in the mask

						done := false
						for ; j+h < ChunkSize; h++ {
							// Check each block next to this quad
							for k := 0; k < w; k++ {
								// If there's a hole in the mask, exit
								if !mask[n+k+h*ChunkSize] {
									done = true
									break
								}
							}

							if done {
								break
							}
						}

						x[u] = i
						x[v] = j

						// du and dv determine the size and orientation of this face
						var du [3]int
						du[u] = w

						var dv [3]int
						dv[v] = h

						/*
						   // Create a quad for this face. Colour, normal or textures are not stored in this block vertex format.
						   BlockVertex.AppendQuad(new Int3(x[0],                 x[1],                 x[2]),                 // Top-left vertice position
						                          new Int3(x[0] + du[0],         x[1] + du[1],         x[2] + du[2]),         // Top right vertice position
						                          new Int3(x[0] + dv[0],         x[1] + dv[1],         x[2] + dv[2]),         // Bottom left vertice position
						                          new Int3(x[0] + du[0] + dv[0], x[1] + du[1] + dv[1], x[2] + du[2] + dv[2])  // Bottom right vertice position
						                          );
						*/

						// Clear this part of the mask, so we don't add duplicate faces
						for l := 0; l < h; l++ {
							for k := 0; k < w; k++ {
								mask[n+k+l*ChunkSize] = false
							}
						}

						// Increment counters and continue
						i += w
						n += w
					} else {
						i++
						n++
					}
				}
			}
		}
	}
}

func (c *Chunk) Mesh() graphic.IGraphic {
	mat := material.NewStandard(math32.NewColor("white"))
	mat.AddTexture(textures["dirt"])
	mat.SetShader("terrain")
	return graphic.NewMesh(c.CulledGeom(), mat)
}
