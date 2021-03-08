package main

import (
	"math/rand"
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
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type WorldScene struct {
	*core.Node

	cam *camera.Camera

	pitch, yaw     float32
	mouseX, mouseY float32
	chunks         map[int64]map[int64]*Chunk
}

type Chunk struct {
	*core.Node
	x      int64
	y      int64
	blocks [16][256][16]float64
}

func (c *Chunk) f(x, y, z int) float64 {
	if x >= 0 && x < 16 && y >= 0 && y < 256 && z >= 0 && z < 16 {
		return c.blocks[x][y][z]
	}
	return 0
}

func (c *Chunk) findBestVertex(x, y, z int) *math32.Vector3 {
	var v [2][2][2]float64
	for dx := 0; dx <= 1; dx++ {
		for dy := 0; dy <= 1; dy++ {
			for dz := 0; dz <= 1; dz++ {
				v[dx][dy][dz] = c.f(x+dx, y+dy, z+dz)
			}
		}
	}
	changes := 0
	for dx := 0; dx <= 1; dx++ {
		for dy := 0; dy <= 1; dy++ {
			if (v[dx][dy][0] > 0) != (v[dx][dy][1] > 0) {
				changes++
			}
		}
	}
	for dx := 0; dx <= 1; dx++ {
		for dz := 0; dz <= 1; dz++ {
			if (v[dx][0][dz] > 0) != (v[dx][1][dz] > 0) {
				changes++
			}
		}
	}
	for dy := 0; dy <= 1; dy++ {
		for dz := 0; dz <= 1; dz++ {
			if (v[0][dy][dz] > 0) != (v[1][dy][dz] > 0) {
				changes++
			}
		}
	}
	if changes <= 1 {
		return nil
	}
	return &math32.Vector3{X: float32(x) + 0.5, Y: float32(y) + 0.5, Z: float32(z) + 0.5}
}

func (c *Chunk) Generate(mat material.IMaterial) {
	vertIndices := make(map[int]map[int]map[int]uint32)

	var positions []float32
	var normals []float32
	var indices math32.ArrayU32

	for x := -1; x <= 16; x++ {
		for y := -1; y <= 256; y++ {
			for z := -1; z <= 16; z++ {
				v := c.findBestVertex(x, y, z)
				if v == nil {
					continue
				}
				if vertIndices[x] == nil {
					vertIndices[x] = make(map[int]map[int]uint32)
				}
				if vertIndices[x][y] == nil {
					vertIndices[x][y] = make(map[int]uint32)
				}
				vertIndices[x][y][z] = uint32(len(positions) / 3)
				positions = append(positions, v.X)
				positions = append(positions, v.Y)
				positions = append(positions, v.Z)
				normals = append(normals, 0)
				normals = append(normals, 1)
				normals = append(normals, 0)
			}
		}
	}
	for x := -1; x <= 16; x++ {
		for y := -1; y <= 256; y++ {
			for z := -1; z <= 16; z++ {
				solid := c.f(x, y, z) > 0
				solidX := c.f(x+1, y, z) > 0
				solidY := c.f(x, y+1, z) > 0
				solidZ := c.f(x, y, z+1) > 0
				if solid != solidX {
					if solidX {
						indices = append(indices, vertIndices[x][y-1][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x][y][z-1])
						indices = append(indices, vertIndices[x][y-1][z-1])
						indices = append(indices, vertIndices[x][y-1][z])
						indices = append(indices, vertIndices[x][y][z])
					} else {
						indices = append(indices, vertIndices[x][y-1][z-1])
						indices = append(indices, vertIndices[x][y][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x][y-1][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x][y-1][z])
					}
				}
				if solid != solidY {
					if solidY {
						indices = append(indices, vertIndices[x-1][y][z-1])
						indices = append(indices, vertIndices[x][y][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x-1][y][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x-1][y][z])
					} else {
						indices = append(indices, vertIndices[x-1][y][z-1])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x][y][z-1])
						indices = append(indices, vertIndices[x-1][y][z-1])
						indices = append(indices, vertIndices[x-1][y][z])
						indices = append(indices, vertIndices[x][y][z])
					}
				}
				if solid != solidZ {
					if solidZ {
						indices = append(indices, vertIndices[x-1][y-1][z])
						indices = append(indices, vertIndices[x-1][y][z])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x-1][y-1][z])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x][y-1][z])
					} else {
						indices = append(indices, vertIndices[x-1][y-1][z])
						indices = append(indices, vertIndices[x][y][z])
						indices = append(indices, vertIndices[x-1][y][z])
						indices = append(indices, vertIndices[x-1][y-1][z])
						indices = append(indices, vertIndices[x][y-1][z])
						indices = append(indices, vertIndices[x][y][z])
					}
				}
			}
		}
	}

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mesh := graphic.NewMesh(geom, mat)
	c.Add(mesh)
}

type PointLightMesh struct {
	*graphic.Mesh
	Light *light.Point
}

func NewPointLightMesh(color *math32.Color, x, y, z float32) *PointLightMesh {

	l := new(PointLightMesh)

	geom := geometry.NewSphere(0.05, 16, 8)
	mat := material.NewStandard(color)
	mat.SetUseLights(0)
	mat.SetEmissiveColor(color)
	l.Mesh = graphic.NewMesh(geom, mat)
	l.Mesh.SetVisible(true)
	l.Mesh.SetPosition(x, y, z)

	l.Light = light.NewPoint(color, 1.0)
	l.Light.SetLinearDecay(1)
	l.Light.SetQuadraticDecay(1)
	l.Light.SetVisible(true)

	l.Mesh.Add(l.Light)

	return l
}

func NewWorldScene() *WorldScene {
	scene := core.NewNode()

	cam := camera.New(1)
	cam.SetPosition(32, 32, 32)
	scene.Add(cam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	cam.SetAspect(float32(width) / float32(height))

	axes := helper.NewAxes(16)
	scene.Add(axes)

	/*
		l := light.NewDirectional(&math32.Color{R: 1, G: 0, B: 0}, 0.8)
		l.SetDirection(1, 0, 0)
		l.SetVisible(true)
		scene.Add(l)
		l = light.NewDirectional(&math32.Color{R: 0, G: 1, B: 0}, 0.8)
		l.SetDirection(0, 1, 0)
		l.SetVisible(true)
		scene.Add(l)
		l = light.NewDirectional(&math32.Color{R: 0, G: 0, B: 1}, 0.8)
		l.SetDirection(0, 0, 1)
		l.SetVisible(true)
		scene.Add(l)

		l2 := light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, 1.0)
		l2.SetPosition(0, 20, 0)
		l2.SetLinearDecay(1)
		l2.SetQuadraticDecay(1)
		l2.SetVisible(true)
	*/

	chunks := make(map[int64]map[int64]*Chunk)
	chunks[0] = make(map[int64]*Chunk)
	chunks[0][0] = &Chunk{
		Node: core.NewNode(),
		x:    0,
		y:    0,
	}
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunks[0][0].blocks[x][0][z] = 1
			for y := 0; y <= 10; y++ {
				if rand.Float32() > 0.5 {
					break
				}
				chunks[0][0].blocks[x][y][z] = 1
			}
		}
	}

	mat := material.NewMaterial()
	mat.AddTexture(textures["rock"])
	mat.AddTexture(textures["grass2"])
	mat.AddTexture(textures["grass"])
	mat.SetShader("terrain")

	for _, row := range chunks {
		for _, chunk := range row {
			chunk.Generate(mat)
			chunk.SetPosition(0.5, 0.5, 0.5)
			scene.Add(chunk)
		}
	}

	s := &WorldScene{
		Node:   scene,
		cam:    cam,
		yaw:    3.6,
		pitch:  -0.6,
		mouseX: float32(width) / 2,
		mouseY: float32(height) / 2,
	}
	s.updateCamera()
	a.SubscribeID(window.OnCursor, a, s.OnMouseMove)
	window.Get().(*window.GlfwWindow).SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	return s
}

func (s *WorldScene) OnMouseMove(evname string, ev interface{}) {
	e := ev.(*window.CursorEvent)
	s.yaw += math32.DegToRad(0.1 * (s.mouseX - e.Xpos))
	s.pitch += math32.DegToRad(0.1 * (s.mouseY - e.Ypos))
	s.mouseX = e.Xpos
	s.mouseY = e.Ypos
	s.updateCamera()
}

func (s *WorldScene) updateCamera() {
	forward := &math32.Vector3{
		X: -math32.Sin(s.yaw) * math32.Cos(s.pitch),
		Y: -math32.Sin(s.pitch),
		Z: -math32.Cos(s.yaw) * math32.Cos(s.pitch),
	}
	right := &math32.Vector3{
		X: -math32.Cos(s.yaw),
		Y: 0,
		Z: math32.Sin(s.yaw),
	}
	up := forward.Clone().Cross(right)
	pos := s.cam.Position()
	forward, right, up = forward.Normalize(), right.Normalize(), up.Normalize()
	s.cam.SetMatrix(&math32.Matrix4{
		right.X, right.Y, right.Z, 0,
		up.X, up.Y, up.Z, 0,
		forward.X, forward.Y, forward.Z, 0,
		pos.X, pos.Y, pos.Z, 0,
	})
}

func (s *WorldScene) Update(renderer *renderer.Renderer, deltaTime time.Duration) {
	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	s.cam.SetAspect(float32(width) / float32(height))

	forward := &math32.Vector3{
		X: -math32.Sin(s.yaw) * math32.Cos(s.pitch),
		Y: -math32.Sin(s.pitch),
		Z: -math32.Cos(s.yaw) * math32.Cos(s.pitch),
	}
	right := &math32.Vector3{
		X: -math32.Cos(s.yaw),
		Y: 0,
		Z: math32.Sin(s.yaw),
	}
	pos := s.cam.Position()
	ks := a.KeyState()
	if ks.Pressed(window.KeyUp) {
		s.cam.SetPositionVec((&pos).Add(forward.Negate()))
	}
	if ks.Pressed(window.KeyDown) {
		s.cam.SetPositionVec((&pos).Add(forward))
	}
	if ks.Pressed(window.KeyLeft) {
		s.cam.SetPositionVec((&pos).Add(right.Negate()))
	}
	if ks.Pressed(window.KeyRight) {
		s.cam.SetPositionVec((&pos).Add(right))
	}

	a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	if err := renderer.Render(s, s.cam); err != nil {
		panic(err)
	}
}
