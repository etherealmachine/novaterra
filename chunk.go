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

type Chunk struct {
	data [32][32][32]int
}

func NewChunk() *Chunk {
	c := &Chunk{}
	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			for z := 0; z < 32; z++ {
				if y >= 4 {
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

func (c *Chunk) Geom() geometry.IGeometry {
	var positions []float32
	var indices []uint32
	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			for z := 0; z < 32; z++ {
				if c.data[x][y][z] != 0 {
					xf := float32(x)
					yf := float32(x)
					zf := float32(x)
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf)
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf)
					positions = append(positions, xf)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					positions = append(positions, xf+1)
					positions = append(positions, yf)
					positions = append(positions, zf+1)
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf)
					positions = append(positions, xf)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)
					positions = append(positions, xf+1)
					positions = append(positions, yf+1)
					positions = append(positions, zf+1)
					i := uint32(len(indices))
					indices = append(indices, i+0)
					indices = append(indices, i+1)
					indices = append(indices, i+2)
				}
			}
		}
	}
	normals := make([]float32, len(positions))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	normals = geometry.CalculateNormals(indices, positions, normals)
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	return geom
}

func (c *Chunk) Mesh() graphic.IGraphic {
	return graphic.NewMesh(c.Geom(), material.NewStandard(math32.NewColor("white")))
}
