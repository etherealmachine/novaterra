// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"log"
	"time"

	"github.com/ojrac/opensimplex-go"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/experimental/collision"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

func marchTransvoxelCube(x, y, z int, voxels [][][]int8) ([]float32, []float32, []uint32) {
	c := [8]int8{
		voxels[x][y][z],
		voxels[x+1][y][z],
		voxels[x][y][z+1],
		voxels[x+1][y][z+1],
		voxels[x][y+1][z],
		voxels[x][y+1][z+1],
		voxels[x+1][y+1][z+1],
	}
	index := ((c[0] >> 7) & 0x01) |
		((c[1] >> 6) & 0x02) |
		((c[2] >> 5) & 0x04) |
		((c[3] >> 4) & 0x08) |
		((c[4] >> 3) & 0x10) |
		((c[5] >> 2) & 0x20) |
		((c[6] >> 1) & 0x40) |
		int8((byte(c[7]) & 0x80))
	if (index ^ int8((byte(c[7])>>7)&0xFF)) != 0 {
		cell := RegularCellData[RegularCellClass[index]]
		log.Println(cell.GetVertexCount(), cell.GetTriangleCount())
		for _, edgeCode := range RegularVertexData[index] {
			if edgeCode == 0 {
				break
			}
			v0 := (edgeCode >> 4) & 0x0F
			v1 := edgeCode & 0x0F
			d0, d1 := c[v0], c[v1]
			t := int16((d1 << 8) / (d1 - d0))
			if (t & 0x00FF) != 0 {
				// Vertex lies in the interior of the edge.
				//u := 0x0100 - t
				//Q := t*P0 + u*P1
			} else if t == 0 {
				// Vertex lies at the higher-numbered endpoint.
				if v1 == 7 {
					// This cell owns the vertex.
				} else {
					// Reuse corner vertex from a preceding cell.
				}
			} else {
				// Vertex lies at the lower-numbered endpoint.
				// Always reuse corner vertex from a preceding cell.
			}
		}
	}
	return nil, nil, nil
}

func marchTransvoxels(voxels [][][]int8, N, M, L int) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	for x := 0; x < N; x++ {
		for z := 0; z < L; z++ {
			for y := 0; y < M; y++ {
				p, n, i := marchTransvoxelCube(x, y, z, voxels)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
	}
	return positions, normals, indices
}

func transvoxelDemo() {

	a := app.App()
	scene := core.NewNode()

	N, M, L := 16, 16, 16

	noise := opensimplex.NewNormalized32(0)
	voxels := make([][][]int8, N+3)
	for x := 0; x < N+3; x++ {
		voxels[x] = make([][]int8, M+3)
		for y := 0; y < M+3; y++ {
			voxels[x][y] = make([]int8, L+3)
		}
	}
	for x := 0; x < N+3; x++ {
		for z := 0; z < L+3; z++ {
			height := math32.Min(float32(M+3)-1, float32(M+3)*octaveNoise(noise, 16, float32(x), float32(z), .5, 0.07)+1)
			for y := 0; height > 0; y++ {
				if height >= 1 {
					voxels[x][y][z] = -127
				} else {
					voxels[x][y][z] = -int8(height) * 127
				}
				height--
			}
		}
	}

	// Create a mesh and add it to the scene
	positions, normals, indices := marchTransvoxels(voxels, N, M, L)
	mesh := NewFastMesh(positions, normals, indices)
	mesh.SetName("Mesh")
	mesh.(*graphic.Mesh).SetPosition(-float32(N)/2, 0, -float32(L)/2)
	scene.Add(mesh)

	log.Println("Lights")
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	mat := material.NewStandard(math32.NewColor("White"))
	sphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), mat)
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
		var offset *math32.Vector3
		if e.Key == window.KeyW {
			offset = &math32.Vector3{-10, 0, 0}
		}
		if e.Key == window.KeyA {
			offset = &math32.Vector3{0, 0, 10}
		}
		if e.Key == window.KeyS {
			offset = &math32.Vector3{10, 0, 0}
		}
		if e.Key == window.KeyD {
			offset = &math32.Vector3{0, 0, -10}
		}
		if offset != nil {
			p := mesh.Position()
			newPos := (&p).Add(offset)
			mesh.(*graphic.Mesh).SetPosition(newPos.X, newPos.Y, newPos.Z)
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

	caster := collision.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})
	mouseDown := false
	a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
		mouseDown = true
	})
	a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
		mouseDown = false
	})

	log.Println("Action!")
	a.Run(func(renderer *renderer.Renderer, _ time.Duration) {

		if mouseDown {
			width, height := a.GetSize()
			x := 2*(mouseX/float32(width)) - 1
			y := -2*(mouseY/float32(height)) + 1

			caster.SetFromCamera(cam, x, y)
			intersects := caster.IntersectObject(mesh, false)
			if len(intersects) > 0 {
				firstHit := intersects[0]
				voxelX := int(math32.Round(firstHit.Point.X - mesh.Position().X))
				voxelY := int(math32.Round(firstHit.Point.Y - mesh.Position().Y))
				voxelZ := int(math32.Round(firstHit.Point.Z - mesh.Position().Z))
				for ox := -1; ox <= 1; ox++ {
					for oy := -1; oy <= 1; oy++ {
						for oz := -1; oz <= 1; oz++ {
							x, y, z := voxelX+ox, voxelY+oy, voxelZ+oz
							if x >= 0 && x < N && y >= 0 && y < M && z >= 0 && z < L {
								if voxels[x][y][z] < 127 {
									voxels[x][y][z]++
								}
							}
						}
					}
				}
				positions, normals, indices = marchTransvoxels(voxels, N, M, L)
				geom := mesh.(*graphic.Mesh).GetGeometry()
				geom.Init()
				geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
				geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
				geom.SetIndices(indices)
			}
		}

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
