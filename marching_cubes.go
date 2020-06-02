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

func marchCubes(inObject func(v *math32.Vector3) bool) []*math32.Vector3 {
	var points []*math32.Vector3
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			for k := 0; k < 100; k++ {
				x := float32(i)/10 - 5
				y := float32(j)/10 - 5
				z := float32(k)/10 - 5
				i := 0
				n0 := &math32.Vector3{x - 1, y - 0.5, z - 1}
				n1 := &math32.Vector3{x + 1, y - 0.5, z - 1}
				n2 := &math32.Vector3{x + 1, y - 0.5, z + 1}
				n3 := &math32.Vector3{x - 1, y - 0.5, z + 1}
				n4 := &math32.Vector3{x - 1, y + 0.5, z - 1}
				n5 := &math32.Vector3{x + 1, y + 0.5, z - 1}
				n6 := &math32.Vector3{x + 1, y + 0.5, z + 1}
				n7 := &math32.Vector3{x - 1, y + 0.5, z + 1}
				if inObject(n0) {
					i |= 1
					points = append(points, n0)
				}
				if inObject(n1) {
					i |= 2
					points = append(points, n1)
				}
				if inObject(n2) {
					i |= 4
					points = append(points, n2)
				}
				if inObject(n3) {
					i |= 8
					points = append(points, n3)
				}
				if inObject(n4) {
					i |= 16
					points = append(points, n4)
				}
				if inObject(n5) {
					i |= 32
					points = append(points, n5)
				}
				if inObject(n6) {
					i |= 64
					points = append(points, n6)
				}
				if inObject(n7) {
					i |= 128
					points = append(points, n7)
				}
			}
		}
	}
	return points
}

func main() {

	a := app.App()
	scene := core.NewNode()

	// Create a mesh and add it to the scene
	geom := geometry.NewGeometry()
	vertices := marchCubes(func(v *math32.Vector3) bool {
		return v.DistanceTo(&math32.Vector3{0, 0, 0}) < 5
	})
	geom.AddVBO(gls.NewVBO(flatten(vertices)).AddAttrib(gls.VertexPosition))
	/*
		geom.AddVBO(gls.NewVBO(computeNormals(vertices)).AddAttrib(gls.VertexNormal))
		geom.SetIndices(indices(vertices))
		mat := material.NewStandard(math32.NewColor("Green"))
		mat.SetSide(material.SideDouble)
	*/
	mat := material.NewPoint(math32.NewColor("white"))
	mat.SetSize(50)
	//mesh := graphic.NewMesh(geom, mat)
	mesh := graphic.NewPoints(geom, mat)
	scene.Add(mesh)

	// Lights
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 40)
	camControl := camera.NewOrbitControl(cam)
	camControl.Rotate(math32.DegToRad(45), math32.DegToRad(-25))
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
