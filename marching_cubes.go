package main

import (
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
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

func computeNormals(vertices []*math32.Vector3) math32.ArrayF32 {
	normals := math32.NewArrayF32(len(vertices), len(vertices))
	for i := 0; i < len(vertices)/3; i++ {
		n := math32.Normal(vertices[i*3], vertices[i*3+1], vertices[i*3+2], nil)
		normals[i*3] = n.X
		normals[i*3+1] = n.Y
		normals[i*3+2] = n.Z
	}
	return normals
}

func indices(vertices []*math32.Vector3) math32.ArrayU32 {
	indices := math32.NewArrayU32(len(vertices), len(vertices))
	for i := range vertices {
		indices[i] = uint32(i)
	}
	return indices
}

func main() {

	a := app.App()
	scene := core.NewNode()

	// Create a mesh and add it to the scene
	geom := geometry.NewGeometry()
	vertices := []*math32.Vector3{
		{-1, -1, 0},
		{1, -1, 0},
		{0, 1, 0},
	}
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(computeNormals(vertices)).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices(vertices))
	mat := material.NewStandard(math32.NewColor("Green"))
	mat.SetSide(material.SideDouble)
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)

	// Lights
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 10)
	camera.NewOrbitControl(cam)
	scene.Add(cam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	cam.SetAspect(float32(width) / float32(height))

	// Action!
	a.Run(func(renderer *renderer.Renderer, _ time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
