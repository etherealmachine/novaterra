package main

import (
	"github.com/ojrac/opensimplex-go"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/texture"
)

var WireframeMaterial *material.Standard

func init() {
	WireframeMaterial = material.NewStandard(math32.NewColor(("White")))
	WireframeMaterial.SetWireframe(true)
}

func NewFastMesh(positions []float32, normals []float32, indices []uint32) *graphic.Mesh {
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewStandard(math32.NewColor("Green"))
	return graphic.NewMesh(geom, mat)
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
