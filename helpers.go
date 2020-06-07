package main

import (
	"bytes"
	"encoding/binary"

	"github.com/ojrac/opensimplex-go"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/texture"
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

func unflatten(a math32.ArrayF32) []*math32.Vector3 {
	l := make([]*math32.Vector3, len(a)/3)
	for i := 0; i < len(a); i += 3 {
		l[i/3] = &math32.Vector3{
			a[i],
			a[i+1],
			a[i+2],
		}
	}
	return l
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

func fromVertices(vertices []*math32.Vector3) ([]float32, []float32, []uint32) {
	return flatten(vertices), computeNormals(vertices), indices(len(vertices))
}

func reverseWinding(indices []uint32) []uint32 {
	reverse := make([]uint32, len(indices))
	for i := 0; i < len(indices)/3; i++ {
		reverse[i*3] = indices[i*3+2]
		reverse[i*3+1] = indices[i*3+1]
		reverse[i*3+2] = indices[i*3]
	}
	return reverse
}

func inflate(a [][][]int8) [][][]int8 {
	n, m, l := len(a), len(a[0]), len(a[0][0])
	e := make([][][]int8, n+3)
	for x := 0; x < len(e); x++ {
		e[x] = make([][]int8, m+3)
		for y := 0; y < len(e[0]); y++ {
			e[x][y] = make([]int8, l+3)
		}
	}
	for x := 0; x < n; x++ {
		for y := 0; y < m; y++ {
			for z := 0; z < l; z++ {
				mx, my, mz := x, y, z
				e[x+1][y+1][z+1] = a[mx][my][mz]
			}
		}
	}
	return e
}

func NewMesh(vertices []*math32.Vector3) *graphic.Mesh {
	geom := geometry.NewGeometry()
	positions, normals, indices := fromVertices(vertices)
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewStandard(math32.NewColor("Green"))
	return graphic.NewMesh(geom, mat)
}

func NewFastMesh(positions []float32, normals []float32, indices []uint32) *graphic.Mesh {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewStandard(math32.NewColor("Green"))
	return graphic.NewMesh(geom, mat)
}

func NewDoubleSidedMesh(positions []float32, normals []float32, indices []uint32) *core.Node {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewStandard(math32.NewColor("Green"))
	mesh := graphic.NewMesh(geom, mat)
	rGeom := geometry.NewGeometry()
	rGeom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	rGeom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	rGeom.SetIndices(reverseWinding(indices))
	rMat := material.NewStandard(math32.NewColor("Red"))
	rMesh := graphic.NewMesh(rGeom, rMat)
	n := core.NewNode()
	n.Add(mesh)
	n.Add(rMesh)
	return n
}

func NewWireframeMesh(vertices []*math32.Vector3) *graphic.Mesh {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	mat := material.NewStandard(math32.NewColor("White"))
	mat.SetWireframe(true)
	return graphic.NewMesh(geom, mat)
}

func NewPointsMesh(vertices []*math32.Vector3) *graphic.Points {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	mat := material.NewPoint(math32.NewColor("White"))
	mat.SetSize(50)
	return graphic.NewPoints(geom, mat)
}

func NewLinesMesh(lines []float32) *graphic.Lines {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(lines).AddAttrib(gls.VertexPosition))
	mat := material.NewStandard(math32.NewColor("Red"))
	return graphic.NewLines(geom, mat)
}

func NewSpriteLabel(txt string) *graphic.Sprite {
	textImg := font.DrawText(txt)
	tex := texture.NewTexture2DFromRGBA(textImg)
	textMat := material.NewStandard(math32.NewColor("White"))
	textMat.AddTexture(tex)
	textMat.SetTransparent(true)
	label := graphic.NewSprite(float32(textImg.Bounds().Dx())/100, float32(textImg.Bounds().Dy())/100, textMat)
	return label
}

func octaveNoise(noise opensimplex.Noise32, iters int, x, y, z float32, persistence, scale float32) float32 {
	var maxamp float32 = 0
	var amp float32 = 1
	freq := scale
	var value float32 = 0

	for i := 0; i < iters; i++ {
		value += noise.Eval3(x*freq, y*freq, z*freq) * amp
		maxamp += amp
		amp *= persistence
		freq *= 2
	}

	return value / maxamp
}

func readColorAt(gl *gls.GLS, x, y int) *math32.Color4 {
	pixels := gl.ReadPixels(x, y, 1, 1, gls.RGBA, gls.FLOAT)
	var r, g, b, a float32
	binary.Read(bytes.NewBuffer(pixels[0:4]), binary.LittleEndian, &r)
	binary.Read(bytes.NewBuffer(pixels[4:8]), binary.LittleEndian, &g)
	binary.Read(bytes.NewBuffer(pixels[8:12]), binary.LittleEndian, &b)
	binary.Read(bytes.NewBuffer(pixels[12:16]), binary.LittleEndian, &a)
	return &math32.Color4{R: r, G: g, B: b, A: a}
}

func simplexTerrain(n, m, l int) [][][]int8 {
	noise := opensimplex.NewNormalized32(0)
	voxels := make([][][]int8, n)
	for x := 0; x < n; x++ {
		voxels[x] = make([][]int8, m)
		for y := 0; y < m; y++ {
			voxels[x][y] = make([]int8, l)
		}
	}
	for x := 0; x < n; x++ {
		for y := 0; y < m; y++ {
			for z := 0; z < l; z++ {
				voxels[x][y][z] = int8(256*octaveNoise(noise, 16, float32(x), float32(y), float32(z), .5, 0.07) - 127)
			}
		}
	}
	return voxels
}

func voxelsAtStep(i uint8) [][][]int8 {
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
		voxels[0][1][1] = -127
	}
	if (i>>7)&1 != 0 {
		voxels[1][1][1] = -127
	}
	return voxels
}
