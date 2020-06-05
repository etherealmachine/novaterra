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

func generateTriangles(index int, n [8]*math32.Vector3, v [8]float32, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var l [12]*math32.Vector3
	for i := range MarchingCubesEdgeConnections {
		if MarchingCubesEdges[index]&(1<<i) != 0 {
			i0, i1 := MarchingCubesEdgeConnections[i][0], MarchingCubesEdgeConnections[i][1]
			v0, v1 := isosurface(n[i0]), isosurface(n[i1])
			l[i] = n[i0].Clone().Lerp(n[i1], (isolevel-v0)/(v1-v0))
		}
	}
	var vertices []*math32.Vector3
	triangles := MarchingCubesTriangles[index]
	for i := 0; triangles[i] != -1; i += 3 {
		a, b, c := l[triangles[i]], l[triangles[i+1]], l[triangles[i+2]]
		vertices = append(vertices, a)
		vertices = append(vertices, b)
		vertices = append(vertices, c)
	}
	return vertices
}

func marchCube(origin *math32.Vector3, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var n [8]*math32.Vector3
	var v [8]float32
	for i, offset := range MarchingCubesNeighborOffsets {
		n[i] = origin.Clone().Add(offset)
		v[i] = isosurface(n[i])
	}
	index := 0
	for i, value := range v {
		if value >= isolevel {
			index |= (1 << i)
		}
	}
	if MarchingCubesEdges[index] == 0 {
		return nil
	}
	return generateTriangles(index, n, v, isosurface, isolevel)
}

func marchCubes(size int, isosurface func(v *math32.Vector3) float32, isolevel float32) []*math32.Vector3 {
	var vertices []*math32.Vector3
	for i := -1; i <= size; i++ {
		for j := -1; j <= size; j++ {
			for k := -1; k <= size; k++ {
				v := &math32.Vector3{float32(i), float32(j), float32(k)}
				vertices = append(vertices, marchCube(v, isosurface, isolevel)...)
			}
		}
	}
	return vertices
}

func marchVoxels(voxels [][][]float32) []*math32.Vector3 {
	return marchCubes(len(voxels), func(v *math32.Vector3) float32 {
		x, y, z := int(v.X), int(v.Y), int(v.Z)
		if x >= 0 && x < len(voxels) && y >= 0 && y < len(voxels[x]) && z >= 0 && z < len(voxels[x][y]) {
			return voxels[x][y][z]
		}
		return 0
	}, 0.5)
}

func MarchingCubesDemo() {

	a := app.App()
	scene := core.NewNode()

	N, M, L := 16, 16, 16

	noise := opensimplex.NewNormalized32(0)
	voxels := make([][][]float32, N)
	for x := 0; x < N; x++ {
		voxels[x] = make([][]float32, M)
		for y := 0; y < M; y++ {
			voxels[x][y] = make([]float32, L)
		}
	}
	for x := 0; x < N; x++ {
		for z := 0; z < L; z++ {
			height := math32.Min(float32(M)-1, float32(M)*octaveNoise(noise, 16, float32(x), float32(z), .5, 0.07)+1)
			for y := 0; height > 0; y++ {
				voxels[x][y][z] = math32.Min(1, height)
				height--
			}
		}
	}

	// Create a mesh and add it to the scene
	vertices := marchVoxels(voxels)
	mesh := NewMesh(vertices)
	mesh.SetName("Mesh")
	wireframe := NewWireframeMesh(vertices)
	wireframe.SetVisible(false)
	wireframe.SetName("Wireframe")
	points := NewPointsMesh(vertices)
	points.SetVisible(false)
	points.SetName("Points")
	normals := NewNormalsMesh(vertices)
	normals.SetVisible(false)
	normals.SetName("Normals")
	group := core.NewNode()
	group.Add(mesh)
	group.Add(wireframe)
	group.Add(points)
	group.Add(normals)
	group.SetPosition(-float32(N)/2, 0, -float32(L)/2)
	scene.Add(group)

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
			wireframe.SetVisible(!wireframe.Visible())
			points.SetVisible(!points.Visible())
			normals.SetVisible(!normals.Visible())
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
				voxelX := int(math32.Round(firstHit.Point.X - group.Position().X))
				voxelY := int(math32.Round(firstHit.Point.Y - group.Position().Y))
				voxelZ := int(math32.Round(firstHit.Point.Z - group.Position().Z))
				for ox := -1; ox <= 1; ox++ {
					for oy := -1; oy <= 1; oy++ {
						for oz := -1; oz <= 1; oz++ {
							x, y, z := voxelX+ox, voxelY+oy, voxelZ+oz
							if x >= 0 && x < N && y >= 0 && y < M && z >= 0 && z < L {
								voxels[x][y][z] += 0.1
								voxels[x][y][z] = math32.Min(voxels[x][y][z], 1)
							}
						}
					}
				}
				vertices = marchVoxels(voxels)
				geom := mesh.(*graphic.Mesh).GetGeometry()
				geom.Init()
				indices := indices(len(vertices))
				positions := flatten(vertices)
				normals := computeNormals(vertices)
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
