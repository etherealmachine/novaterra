package main

import (
	"log"
	"time"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

type TransvoxelChunk struct {
	*core.Node

	voxels  [][][]int8
	x, y, z int
	lod     int
}

func (c *TransvoxelChunk) HandleVoxelClick(x, y, z int, shift bool) {
	n, m, l := len(c.voxels), len(c.voxels[0]), len(c.voxels[0][0])
	for ox := -1; ox <= 1; ox++ {
		for oy := -1; oy <= 1; oy++ {
			for oz := -1; oz <= 1; oz++ {
				cx, cy, cz := x+ox, y+oy, z+oz
				if cx >= 0 && cx < n && cy >= 0 && cy < m && cz >= 0 && cz < l {
					if shift && c.voxels[cx][cy][cz] < 0 {
						c.voxels[cx][cy][cz]++
					} else if !shift && c.voxels[cx][cy][cz] > -127 {
						c.voxels[cx][cy][cz]--
					}
				}
			}
		}
	}
	positions, normals, indices := marchTransvoxels(inflate(c.voxels))
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mesh := c.Children()[0].(*graphic.Mesh)
	mat := mesh.GetMaterial(0)
	mesh.Init(geom, mat)
}

func NewTransvoxelChunk(lod, x, y, z int, voxels [][][]int8) *TransvoxelChunk {
	voxels = inflate(voxels)
	N, M, L := len(voxels), len(voxels[0]), len(voxels[0][0])
	positions, normals, indices := marchTransvoxels(voxels)
	m := NewFastMesh(positions, normals, indices)
	group := core.NewNode()
	group.Add(m)
	group.SetPosition(-float32(N)/2-1+float32(x), -float32(M)/2-1+float32(y), -float32(L)/2-1+float32(z))
	scale := float32(int(1) << lod)
	group.SetScale(scale, scale, scale)
	return &TransvoxelChunk{group, voxels, x, y, z, lod}
}

func NewTransvoxelTerrainScene() *core.Node {
	scene := core.NewNode()

	N, M, L := 16, 16, 16
	for x := 0; x < 10; x++ {
		for z := 0; z < 10; z++ {
			dx, dz := float32(x-5), float32(z-5)
			dist := math32.Floor(math32.Sqrt(dx*dx + dz*dz))
			lod := int(dist)
			if lod > 2 {
				lod = 2
			}
			voxels := simplexTerrain(lod, x*16, 0, z*16, N, M, L)
			terrain := NewTransvoxelChunk(lod, x*16, 0, z*16, voxels)
			scene.Add(terrain)
		}
	}

	log.Println("Lights")
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	sphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), material.NewStandard(math32.NewColor("White")))
	sphere.SetPosition(30, 60, -30)
	scene.Add(sphere)
	pointLight := light.NewPoint(math32.NewColor("White"), 10000.0)
	pointLight.SetPosition(30, 60, -30)
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

	/*
		caster := collision.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})
		mouseX, mouseY := float32(0), float32(0)
		a.SubscribeID(window.OnCursor, a, func(evname string, ev interface{}) {
			e := ev.(*window.CursorEvent)
			mouseX = e.Xpos
			mouseY = e.Ypos
		})

		mouseDown := false
		shift := false
		a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
			mouseDown = true
			shift = ev.(*window.MouseEvent).Button == 1
		})
		a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
			mouseDown = false
			shift = false
		})
	*/

	log.Println("Action!")
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		/*
			if mouseDown {
				width, height := a.GetSize()
				x := 2*(mouseX/float32(width)) - 1
				y := -2*(mouseY/float32(height)) + 1

				caster.SetFromCamera(cam, x, y)
				intersects := caster.IntersectObject(terrain, true)
				if len(intersects) > 0 {
					firstHit := intersects[0]
					pos := terrain.Position()
					vx := int(math32.Clamp(firstHit.Point.X-pos.X+0.5+0.5*caster.Direction().X, 0, float32(N-1)))
					vy := int(math32.Clamp(firstHit.Point.Y-pos.Y+0.5+0.5*caster.Direction().Y, 0, float32(M-1)))
					vz := int(math32.Clamp(firstHit.Point.Z-pos.Z+0.5+0.5*caster.Direction().Z, 0, float32(L-1)))
					terrain.HandleVoxelClick(vx, vy, vz, shift)
				}
			}
		*/

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})

	return scene
}
