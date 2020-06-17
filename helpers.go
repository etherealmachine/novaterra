package main

import (
	"github.com/ojrac/opensimplex-go"

	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

var WireframeMaterial *material.Standard

func init() {
	WireframeMaterial = material.NewStandard(math32.NewColor(("White")))
	WireframeMaterial.SetWireframe(true)
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
