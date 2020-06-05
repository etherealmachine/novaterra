// Implementation of Eric Lengyel's Transvoxel Algorithm - http://transvoxel.org/
package main

import (
	"log"
	"time"

	"github.com/ojrac/opensimplex-go"

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
	"github.com/g3n/engine/window"
)

var cornerVertices = [8][3]float32{
	{0, 0, 0},
	{1, 0, 0},
	{0, 0, 1},
	{1, 0, 1},
	{0, 1, 0},
	{1, 1, 0},
	{0, 1, 1},
	{1, 1, 1},
}

func generateTransvoxelMesh(c [8]int8, voxels [][][]int8) ([]float32, []float32, []uint32) {
	index := uint8(((c[0] >> 7) & 0x01) |
		((c[1] >> 6) & 0x02) |
		((c[2] >> 5) & 0x04) |
		((c[3] >> 4) & 0x08) |
		((c[4] >> 3) & 0x10) |
		((c[5] >> 2) & 0x20) |
		((c[6] >> 1) & 0x40) |
		int8((byte(c[7]) & 0x80)))
	var positions []float32
	var normals []float32
	var indices []uint32
	if (index ^ uint8((byte(c[7])>>7)&0xFF)) != 0 {

		/*
			var cornerNormals [8][3]float32
			for i := 0; i < 8; i++ {
				p := cornerVertices[i]
				x, y, z := int(p[0]), int(p[1]), int(p[2])
				nx := float32(voxels[x+1][y][z]-voxels[x-1][y][z]) * 0.5
				ny := float32(voxels[x][y+1][z]-voxels[x][y-1][z]) * 0.5
				nz := float32(voxels[x][y][z+1]-voxels[x][y][z-1]) * 0.5
				cornerNormals[i][0] = nx
				cornerNormals[i][1] = ny
				cornerNormals[i][2] = nz
			}
		*/

		cell := RegularCellData[RegularCellClass[index]]
		vertexLocations := RegularVertexData[index]
		vertexCount := cell.GetVertexCount()
		triangleCount := cell.GetTriangleCount()
		log.Printf("%d vertices and %d triangles", vertexCount, triangleCount)
		for i := 0; i < vertexCount; i++ {
			edge := vertexLocations[i] >> 8
			reuseIndex := edge & 0xF               // Vertex id which should be created or reused 1, 2 or 3
			rDir := edge >> 4                      // Direction to go to reach a previous cell for reusing
			v0 := (vertexLocations[i] >> 4) & 0x0F // First Corner Index
			v1 := (vertexLocations[i]) & 0x0F      // Second Corner Index
			d0 := c[v0]
			d1 := c[v1]
			t := float32(d1) / float32(d1-d0)
			log.Printf("vertex %d, reuse %d, direction %d", i, reuseIndex, rDir)
			if byte(t)&0x00FF != 0 {
				log.Println(" Vertex lies in the interior of the edge")
			} else if t == 0 {
				log.Println(" Vertex lies at the higher numbered endpoint")
				if v1 == 7 {
					log.Println("  This cell owns the vertex")
				} else {
					log.Println("  Try to re-use corner vertex from a preceding cell")
				}
			} else {
				log.Println(" Vertex lies at the lower-numbered endpoint. Try to reuse corner vertex from a preceding cell")
			}
			// Vertices at the two corners
			p0 := cornerVertices[v0]
			p1 := cornerVertices[v1]
			// Normals at the two corners
			//n0 := cornerNormals[v0]
			//n1 := cornerNormals[v1]
			// Linearly interpolate along the 2 vertices to get the new vertex
			qX, qY, qZ := p0[0]*t+(1-t)*p1[0], p0[1]*t+(1-t)*p1[1], p0[2]*t+(1-t)*p1[2]
			//nX, nY, nZ := n0[0]*t+(1-t)*n1[0], n0[1]*t+(1-t)*n1[1], n0[2]*t+(1-t)*n1[2]
			positions = append(positions, qX)
			//normals = append(normals, nX)
			indices = append(indices, uint32(len(positions)-1))
			positions = append(positions, qY)
			//normals = append(normals, nY)
			indices = append(indices, uint32(len(positions)-1))
			positions = append(positions, qZ)
			//normals = append(normals, nZ)
			indices = append(indices, uint32(len(positions)-1))
		}
	}
	return positions, normals, indices
}

func marchTransvoxels(voxels [][][]int8, N, M, L int) ([]float32, []float32, []uint32) {
	var positions []float32
	var normals []float32
	var indices []uint32
	for x := 0; x < N; x++ {
		for z := 0; z < L; z++ {
			for y := 0; y < M; y++ {
				corners := [8]int8{
					voxels[x][y][z],
					voxels[x+1][y][z],
					voxels[x][y][z+1],
					voxels[x+1][y][z+1],
					voxels[x][y+1][z],
					voxels[x][y+1][z+1],
					voxels[x+1][y+1][z+1],
				}
				p, n, i := generateTransvoxelMesh(corners, voxels)
				positions = append(positions, p...)
				normals = append(normals, n...)
				indices = append(indices, i...)
			}
		}
	}
	return positions, normals, indices
}

func transvoxelCase(i int) *graphic.Mesh {
	corners := [8]int8{
		1,
		1,
		1,
		1,
		1,
		1,
		1,
		-1,
	}
	voxels := [][][]int8{
		{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
		{{1, 1, 1}, {1, -1, 1}, {1, 1, 1}},
		{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}},
	}
	positions, normals, indices := generateTransvoxelMesh(corners, voxels)
	return NewFastMesh(positions, normals, indices).(*graphic.Mesh)
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

	/*
		// Create a mesh and add it to the scene
		positions, normals, indices := marchTransvoxels(voxels, N, M, L)
		mesh := NewFastMesh(positions, normals, indices)
		mesh.SetName("Mesh")
		mesh.(*graphic.Mesh).SetPosition(-float32(N)/2, 0, -float32(L)/2)
		scene.Add(mesh)
	*/

	index := 0
	mesh := transvoxelCase(index)
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
		if e.Key == window.KeyTab {
			scene.Remove(mesh)
			index++
			if index >= 256 {
				index = 0
			}
			mesh = transvoxelCase(index)
			scene.Add(mesh)
		}
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
			mesh.SetPosition(newPos.X, newPos.Y, newPos.Z)
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
