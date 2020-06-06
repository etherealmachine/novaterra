package main

import (
	"fmt"
	"log"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

func cubes(voxels [][][]int8) *core.Node {
	mat := material.NewStandard(math32.NewColor("Gray"))
	cubes := core.NewNode()
	for x := 0; x < len(voxels); x++ {
		for y := 0; y < len(voxels[x]); y++ {
			for z := 0; z < len(voxels[x][y]); z++ {
				if voxels[x][y][z] < 0 {
					cube := geometry.NewCube(1)
					mesh := graphic.NewMesh(cube, mat)
					mesh.SetPosition(float32(x), float32(y), float32(z))
					cubes.Add(mesh)
				}
			}
		}
	}
	cubes.SetPosition(-8.5, 0, -8.5)
	return cubes
}

func countVertices(n core.INode) int {
	sum := 0
	if m, ok := n.(*graphic.Mesh); ok {
		sum += len(*m.GetGeometry().VBO(gls.VertexPosition).Buffer())
	}
	for _, child := range n.Children() {
		sum += countVertices(child)
	}
	return sum
}

func voxelDemo() {

	a := app.App()
	scene := core.NewNode()

	voxels := simplexTerrain(16, 16, 16)

	group := core.NewNode()
	group.Add(cubes(voxels))
	group.Add(marchingCubes(voxels))
	group.Add(transvoxelize(voxels))
	group.ChildAt(1).SetVisible(false)
	group.ChildAt(2).SetVisible(false)
	scene.Add(group)

	label := gui.NewLabel(fmt.Sprintf("Vertices: %d", countVertices(group.Children()[0])))
	label.SetPosition(10, 10)
	label.SetFontSize(14)
	label.SetColor(math32.NewColor("White"))
	scene.Add(label)

	log.Println("Lights")
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	sphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), material.NewStandard(math32.NewColor("White")))
	sphere.SetPosition(40, 40, 40)
	scene.Add(sphere)
	pointLight := light.NewPoint(math32.NewColor("White"), 1000.0)
	pointLight.SetPosition(40, 40, 40)
	scene.Add(pointLight)

	log.Println("Camera")
	cam := camera.New(1)
	cam.SetPosition(0, 0, 40)
	camControl := camera.NewOrbitControl(cam)
	camControl.SetEnabled(camera.OrbitZoom | camera.OrbitKeys)
	camControl.Rotate(math32.DegToRad(45), math32.DegToRad(-25))
	scene.Add(cam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	cam.SetAspect(float32(width) / float32(height))

	a.SubscribeID(window.OnKeyDown, a, func(evname string, ev interface{}) {
		e := ev.(*window.KeyEvent)
		if e.Key == window.KeyTab {
			for i, c := range group.Children() {
				if c.Visible() {
					c.SetVisible(false)
					if i+1 >= len(group.Children()) {
						i = -1
					}
					group.Children()[i+1].SetVisible(true)
					label.SetText(fmt.Sprintf("Vertices: %d", countVertices(group.Children()[i+1])))
					break
				}
			}
		}
		var offset *math32.Vector3
		if e.Key == window.KeyW {
			offset = &math32.Vector3{-5, 0, 0}
		}
		if e.Key == window.KeyA {
			offset = &math32.Vector3{0, 0, 5}
		}
		if e.Key == window.KeyS {
			offset = &math32.Vector3{5, 0, 0}
		}
		if e.Key == window.KeyD {
			offset = &math32.Vector3{0, 0, -5}
		}
		if offset != nil {
			p := group.Position()
			newPos := (&p).Add(offset)
			group.SetPosition(newPos.X, newPos.Y, newPos.Z)
		}

		switch e.Key {
		case window.KeyUp:
			camControl.Rotate(0, -camControl.KeyRotSpeed)
		case window.KeyDown:
			camControl.Rotate(0, camControl.KeyRotSpeed)
		case window.KeyLeft:
			camControl.Rotate(-camControl.KeyRotSpeed, 0)
		case window.KeyRight:
			camControl.Rotate(camControl.KeyRotSpeed, 0)
		}
	})

	mouseX, mouseY := float32(0), float32(0)
	a.SubscribeID(window.OnCursor, a, func(evname string, ev interface{}) {
		e := ev.(*window.CursorEvent)
		mouseX = e.Xpos
		mouseY = e.Ypos
	})

	mouseDown := false
	a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
		mouseDown = true
	})
	a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
		mouseDown = false
	})

	log.Println("Action!")
	a.Run(func(renderer *renderer.Renderer, _ time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
