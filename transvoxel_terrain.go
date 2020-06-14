package main

import (
	"fmt"
	"time"

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
	"github.com/go-gl/glfw/v3.2/glfw"
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

type TransvoxelTerrainScene struct {
	*core.Node

	chunks          []*TransvoxelChunk
	cam             *camera.Camera
	orbitCam, fpCam *camera.Camera
	orbitControl    *camera.OrbitControl
	pitch, yaw      float32
	mouseX, mouseY  float32
	fpsLabel        *gui.Label

	velocity *math32.Vector3
}

func NewTransvoxelTerrainScene() *TransvoxelTerrainScene {
	scene := core.NewNode()

	var chunks []*TransvoxelChunk
	for x := 0; x < 10; x++ {
		for z := 0; z < 10; z++ {
			chunk := NewTransvoxelChunk(0, x, 0, z)
			scene.Add(chunk)
			chunks = append(chunks, chunk)
		}
	}

	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	lightSphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), material.NewStandard(math32.NewColor("White")))
	lightSphere.SetPosition(30, 60, -30)
	scene.Add(lightSphere)
	pointLight := light.NewPoint(math32.NewColor("White"), 10000.0)
	pointLight.SetPosition(30, 60, -30)
	scene.Add(pointLight)

	orbitCam := camera.New(1)
	orbitCam.SetPosition(0, 0, 40)
	orbitControl := camera.NewOrbitControl(orbitCam)
	orbitControl.SetEnabled(camera.OrbitZoom | camera.OrbitKeys)
	orbitControl.Rotate(math32.DegToRad(45), math32.DegToRad(-25))
	scene.Add(orbitCam)

	fpCam := camera.New(1)
	fpCam.SetPosition(100, 100, 40)
	scene.Add(fpCam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	orbitCam.SetAspect(float32(width) / float32(height))
	fpCam.SetAspect(float32(width) / float32(height))

	fpsLabel := gui.NewLabel("FPS:")
	fpsLabel.SetPosition(10, 10)
	fpsLabel.SetFontSize(14)
	fpsLabel.SetColor(math32.NewColor("White"))
	scene.Add(fpsLabel)

	s := &TransvoxelTerrainScene{
		Node:         scene,
		cam:          orbitCam,
		orbitCam:     orbitCam,
		orbitControl: orbitControl,
		fpCam:        fpCam,
		chunks:       chunks,
		fpsLabel:     fpsLabel,
		mouseX:       float32(width) / 2,
		mouseY:       float32(height) / 2,
		velocity:     &math32.Vector3{},
	}

	a.SubscribeID(window.OnKeyDown, a, s.OnKeyPress)
	a.SubscribeID(window.OnKeyRepeat, a, s.OnKeyPress)
	a.SubscribeID(window.OnCursor, a, s.OnMouseMove)
	return s
}

func (s *TransvoxelTerrainScene) OnKeyPress(evname string, ev interface{}) {
	e := ev.(*window.KeyEvent)
	if s.cam == s.fpCam {
		forward := &math32.Vector3{
			-math32.Sin(s.yaw) * math32.Cos(s.pitch),
			-math32.Sin(s.pitch),
			-math32.Cos(s.yaw) * math32.Cos(s.pitch),
		}
		right := &math32.Vector3{
			-math32.Cos(s.yaw),
			0,
			math32.Sin(s.yaw),
		}
		pos := s.fpCam.Position()
		switch e.Key {
		case window.KeyUp:
			s.cam.SetPositionVec((&pos).Add(forward.Negate()))
		case window.KeyDown:
			s.cam.SetPositionVec((&pos).Add(forward))
		case window.KeyLeft:
			s.cam.SetPositionVec((&pos).Add(right.Negate()))
		case window.KeyRight:
			s.cam.SetPositionVec((&pos).Add(right))
		case window.KeySpace:
			s.velocity.Add(&math32.Vector3{0, 10, 0})
		}
	} else {
		switch e.Key {
		case window.KeyUp:
			s.orbitControl.Rotate(0, -s.orbitControl.KeyRotSpeed)
		case window.KeyDown:
			s.orbitControl.Rotate(0, s.orbitControl.KeyRotSpeed)
		case window.KeyLeft:
			s.orbitControl.Rotate(-s.orbitControl.KeyRotSpeed, 0)
		case window.KeyRight:
			s.orbitControl.Rotate(s.orbitControl.KeyRotSpeed, 0)
		}
	}

	switch e.Key {
	case window.KeyTab:
		if s.cam == s.orbitCam {
			s.cam = s.fpCam
			window.Get().(*window.GlfwWindow).SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		} else {
			s.cam = s.orbitCam
			window.Get().(*window.GlfwWindow).SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		}
	}
}

func (s *TransvoxelTerrainScene) OnMouseMove(evname string, ev interface{}) {
	e := ev.(*window.CursorEvent)
	if s.cam == s.fpCam {
		s.yaw += math32.DegToRad(0.1 * (s.mouseX - e.Xpos))
		s.pitch += math32.DegToRad(0.1 * (s.mouseY - e.Ypos))
		forward := &math32.Vector3{
			-math32.Sin(s.yaw) * math32.Cos(s.pitch),
			-math32.Sin(s.pitch),
			-math32.Cos(s.yaw) * math32.Cos(s.pitch),
		}
		right := &math32.Vector3{
			-math32.Cos(s.yaw),
			0,
			math32.Sin(s.yaw),
		}
		up := forward.Clone().Cross(right)
		pos := s.cam.Position()
		forward, right, up = forward.Normalize(), right.Normalize(), up.Normalize()
		s.cam.SetMatrix(&math32.Matrix4{
			right.X, right.Y, right.Z, 0,
			up.X, up.Y, up.Z, 0,
			forward.X, forward.Y, forward.Z, 0,
			pos.X, pos.Y, pos.Z, 1,
		})
	}
	s.mouseX = e.Xpos
	s.mouseY = e.Ypos
}

func (s *TransvoxelTerrainScene) Update(renderer *renderer.Renderer, deltaTime time.Duration) {
	s.fpsLabel.SetText(fmt.Sprintf("FPS: %.0f", 1/float64(deltaTime.Seconds())))

	pos := s.fpCam.Position()
	(&pos).Add(s.velocity.Clone().MultiplyScalar(float32(deltaTime.Seconds())))

	s.velocity.Add((&math32.Vector3{0, -9.8, 0}).MultiplyScalar(float32(deltaTime.Seconds())))

	chunkX, chunkZ := int(math32.Floor(pos.X/16)), int(math32.Floor(pos.Z/16))
	if chunkX >= 0 && chunkX < 10 && chunkZ >= 0 && chunkZ < 10 {
		chunk := s.chunks[chunkX+10*chunkZ]

		voxelX, voxelZ := int(math32.Mod(pos.X, 16)), int(math32.Mod(pos.Z, 16))
		var height float32
		for y := 15; y >= 0; y-- {
			if chunk.voxels[voxelX][y][voxelZ] < 0 {
				height = float32(y)
				break
			}
		}
		if pos.Y-10 <= height {
			s.velocity = &math32.Vector3{}
			pos.Y = height + 10
		}
	}
	s.fpCam.SetPositionVec(&pos)

	a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	if err := renderer.Render(s, s.cam); err != nil {
		panic(err)
	}
}
