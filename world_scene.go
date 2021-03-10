package main

import (
	"log"
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
	"github.com/gonum/blas"
	"github.com/gonum/lapack/native"
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
	blocks [16][256][16]float32
}

func (c *Chunk) f(x, y, z float32) float32 {
	if x < 0 || x >= 16 || y < 0 || y >= 256 || z < 0 || z >= 16 {
		return 0
	}
	return c.blocks[int(x)][int(y)][int(z)]
	/*
		x -= 10
		y -= 10
		z -= 10
		return 5 - float32(math.Sqrt(float64(x*x+y*y+z*z)))
	*/
}

func (c *Chunk) fNormal(x, y, z float32) *math32.Vector3 {
	d := float32(0.01)
	n := &math32.Vector3{
		X: -(c.f(x+d, y, z) - c.f(x-d, y, z)) / 2 / d,
		Y: -(c.f(x, y+d, z) - c.f(x, y-d, z)) / 2 / d,
		Z: -(c.f(x, y, z+d) - c.f(x, y, z-d)) / 2 / d,
	}
	return n.Normalize()
}

func adapt(v0 float32, v1 float32) float32 {
	// v0 and v1 are numbers of opposite sign. This returns how far you need to interpolate from v0 to v1 to get to 0.
	if (v0 > 0) == (v1 > 0) {
		log.Fatalf("v0 and v1 do not have opposite sign %.2f %.2f", v0, v1)
	}
	return (0 - v0) / (v1 - v0)
}

func solveQuadraticErrorFunction(x, y, z float32, positions []*math32.Vector3, normals []*math32.Vector3) *math32.Vector3 {
	a := make([]float64, len(positions)*3)
	b := make([]float64, len(positions))
	for i := 0; i < len(positions); i++ {
		v := positions[i]
		n := normals[i]
		a[i*3] = float64(n.X)
		a[i*3+1] = float64(n.Y)
		a[i*3+2] = float64(n.Z)
		b[i] = float64(v.X*n.X + v.Y*n.Y + v.Z*n.Z)
	}
	var impl native.Implementation
	m := len(a) / 3
	n := 3
	nrhs := 1
	lda := n
	ldb := 1
	work := []float64{0}
	success := impl.Dgels(blas.NoTrans, m, n, nrhs, a, lda, b, ldb, work, -1)
	if !success {
		log.Fatal("failed")
	}
	work = make([]float64, int(work[0]))
	success = impl.Dgels(blas.NoTrans, m, n, nrhs, a, lda, b, ldb, work, len(work))
	if !success {
		return &math32.Vector3{X: x + 0.5, Y: y + 0.5, Z: z + 0.5}
	}
	return &math32.Vector3{
		X: math32.Clamp(float32(b[0]), x, x+1),
		Y: math32.Clamp(float32(b[1]), y, y+1),
		Z: math32.Clamp(float32(b[2]), z, z+1),
	}
}

func (c *Chunk) findBestVertex(x, y, z float32) *math32.Vector3 {
	var v [2][2][2]float32
	for dx := 0; dx <= 1; dx++ {
		for dy := 0; dy <= 1; dy++ {
			for dz := 0; dz <= 1; dz++ {
				v[dx][dy][dz] = c.f(x+float32(dx), y+float32(dy), z+float32(dz))
			}
		}
	}
	var changes []*math32.Vector3
	for dx := 0; dx <= 1; dx++ {
		for dy := 0; dy <= 1; dy++ {
			if (v[dx][dy][0] > 0) != (v[dx][dy][1] > 0) {
				changes = append(changes, &math32.Vector3{X: x + float32(dx), Y: y + float32(dy), Z: z + adapt(v[dx][dy][0], v[dx][dy][1])})
			}
		}
	}
	for dx := 0; dx <= 1; dx++ {
		for dz := 0; dz <= 1; dz++ {
			if (v[dx][0][dz] > 0) != (v[dx][1][dz] > 0) {
				changes = append(changes, &math32.Vector3{X: x + float32(dx), Y: y + adapt(v[dx][0][dz], v[dx][1][dz]), Z: z + float32(dz)})
			}
		}
	}
	for dy := 0; dy <= 1; dy++ {
		for dz := 0; dz <= 1; dz++ {
			if (v[0][dy][dz] > 0) != (v[1][dy][dz] > 0) {
				changes = append(changes, &math32.Vector3{X: x + adapt(v[0][dy][dz], v[1][dy][dz]), Y: y + float32(dy), Z: z + float32(dz)})
			}
		}
	}
	if len(changes) <= 1 {
		return nil
	}

	var normals []*math32.Vector3
	for _, v := range changes {
		normals = append(normals, c.fNormal(v.X, v.Y, v.Z))
	}

	return solveQuadraticErrorFunction(x, y, z, changes, normals)
}

func (c *Chunk) Generate(mat material.IMaterial) {
	vertIndices := make(map[int]map[int]map[int]uint32)

	var positions []float32
	var normals []float32
	var indices math32.ArrayU32

	for x := -1; x <= 16; x++ {
		for y := -1; y <= 256; y++ {
			for z := -1; z <= 16; z++ {
				v := c.findBestVertex(float32(x), float32(y), float32(z))
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
				normal := c.fNormal(float32(x), float32(y), float32(z))
				normals = append(normals, normal.X)
				normals = append(normals, normal.Y)
				normals = append(normals, normal.Z)
			}
		}
	}
	for x := -1; x <= 16; x++ {
		for y := -1; y <= 256; y++ {
			for z := -1; z <= 16; z++ {
				solid := c.f(float32(x), float32(y), float32(z)) > 0
				solidX := c.f(float32(x+1), float32(y), float32(z)) > 0
				solidY := c.f(float32(x), float32(y+1), float32(z)) > 0
				solidZ := c.f(float32(x), float32(y), float32(z+1)) > 0
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

	{
		g := geometry.NewGeometry()
		var vertices []float32
		for i := 0; i < len(positions)/3; i++ {
			vertices = append(vertices, positions[i*3])
			vertices = append(vertices, positions[i*3+1])
			vertices = append(vertices, positions[i*3+2])
			vertices = append(vertices, positions[i*3]+normals[i*3])
			vertices = append(vertices, positions[i*3+1]+normals[i*3+1])
			vertices = append(vertices, positions[i*3+2]+normals[i*3+2])
		}
		g.AddVBO(gls.NewVBO(vertices).AddAttrib(gls.VertexPosition))
		mat := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
		mat.SetLineWidth(1)
		c.Add(graphic.NewLines(g, mat))
	}

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mesh := graphic.NewMesh(geom, mat)
	c.Add(mesh)
}

func NewWorldScene() *WorldScene {
	scene := core.NewNode()

	l := light.NewAmbient(math32.NewColor("white"), 0.1)
	scene.Add(l)

	l2 := light.NewDirectional(math32.NewColor("white"), 0.8)
	l2.SetPosition(1, 1, 1)
	scene.Add(l2)

	cam := camera.New(1)
	cam.SetPosition(32, 32, 32)
	scene.Add(cam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	cam.SetAspect(float32(width) / float32(height))

	axes := helper.NewAxes(16)
	scene.Add(axes)

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
