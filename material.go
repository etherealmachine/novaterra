package main

import (
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

type Material struct {
	material.Standard
	uni  gls.Uniform
	Mode int32
}

func NewMaterial() *Material {
	m := new(Material)
	m.Standard.Init("terrain", math32.NewColor("white"))
	m.AddTexture(textures["dirt"])
	m.uni.Init("Mode")
	return m
}

// RenderSetup transfer this material uniforms and textures to the shader
func (m *Material) RenderSetup(gl *gls.GLS) {
	m.Standard.RenderSetup(gl)
	location := m.uni.Location(gl)
	gl.Uniform1i(location, m.Mode)
}
