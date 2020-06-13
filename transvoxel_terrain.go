package main

import (
	"log"
	"time"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/experimental/physics"
	"github.com/g3n/engine/experimental/physics/object"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
	"github.com/ojrac/opensimplex-go"
)

type TransvoxelChunk struct {
	*core.Node

	voxels  [][][]int8
	x, y, z int
	lod     int

	mesh *graphic.Mesh
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

func NewTransvoxelChunk(lod, x, y, z int) *TransvoxelChunk {
	noise := opensimplex.NewNormalized32(0)
	voxels := make([][][]int8, 16)
	for i := 0; i < 16; i++ {
		voxels[i] = make([][]int8, 16)
		for j := 0; j < 16; j++ {
			voxels[i][j] = make([]int8, 16)
		}
	}
	for ox := 0; ox < 16; ox++ {
		for oz := 0; oz < 16; oz++ {
			height := 15*octaveNoise(noise, 16, float32(x*16+ox), 0, float32(z*16+oz), .5, 0.09) + 1
			density := float32(-127)
			deltaDensity := 256 / height
			for oy := 0; oy < int(height); oy++ {
				voxels[ox][oy][oz] = int8(density)
				density += deltaDensity
			}
		}
	}

	voxels = inflate(voxels)
	positions, normals, indices := marchTransvoxels(voxels)
	m := NewFastMesh(positions, normals, indices)
	group := core.NewNode()
	group.Add(m)
	group.SetPosition(float32(x)*16, float32(y)*16, float32(z)*16)
	return &TransvoxelChunk{group, voxels, x, y, z, lod, m}
}

func NewTransvoxelTerrainScene() *core.Node {
	scene := core.NewNode()

	terrain := NewTransvoxelChunk(0, 0, 0, 0)
	scene.Add(terrain)

	log.Println("Lights")
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	lightSphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), material.NewStandard(math32.NewColor("White")))
	lightSphere.SetPosition(30, 60, -30)
	scene.Add(lightSphere)
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

	log.Println("Action!")

	sim := physics.NewSimulation(scene)
	terrainBody := object.NewBody(terrain.mesh)
	terrainBody.SetBodyType(object.Static)
	sim.AddBody(terrainBody, "Terrain")

	sphereGeom := geometry.NewSphere(0.5, 16, 8)
	mat := material.NewStandard(&math32.Color{1, 1, 1})
	sphere := graphic.NewMesh(sphereGeom, mat)
	sphere.SetPosition(8, 8, 8)
	scene.Add(sphere)
	sphereBody := object.NewBody(sphere)
	sim.AddBody(sphereBody, "Sphere")

	gravity := physics.NewConstantForceField(&math32.Vector3{0, -9.8, 0})
	sim.AddForceField(gravity)

	a.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		switch kev.Key {
		case window.KeyW:
			sphereBody.ApplyVelocityDeltas(math32.NewVector3(-1, 0, 0), math32.NewVector3(0, 0, 1))
		case window.KeyA:
			sphereBody.ApplyVelocityDeltas(math32.NewVector3(0, 0, -1), math32.NewVector3(0, 0, -1))
		case window.KeyS:
			sphereBody.ApplyVelocityDeltas(math32.NewVector3(1, 0, 0), math32.NewVector3(0, 0, 1))
		case window.KeyD:
			sphereBody.ApplyVelocityDeltas(math32.NewVector3(0, 0, 1), math32.NewVector3(0, 0, -1))
		}
	})

	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		sim.Step(float32(deltaTime.Seconds()))

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})

	return scene
}
