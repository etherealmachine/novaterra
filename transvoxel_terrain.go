package main

import (
	"fmt"
	"log"
	"time"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/experimental/collision"
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

type TransvoxelNode struct {
	level    int
	x, y, z  int
	voxels   *[19][19][19]int8
	mesh     *graphic.Mesh
	group    *core.Node
	children *[8]*TransvoxelNode
}

func (n *TransvoxelNode) HandleVoxelClick(x, y, z int, shift bool) {
	for ox := -1; ox <= 1; ox++ {
		for oy := -1; oy <= 1; oy++ {
			for oz := -1; oz <= 1; oz++ {
				cx, cy, cz := x+ox, y+oy, z+oz
				if cx >= 0 && cx < 16 && cy >= 0 && cy < 16 && cz >= 0 && cz < 16 {
					if shift && n.voxels[cx][cy][cz] < 0 {
						n.voxels[cx][cy][cz]++
					} else if !shift && n.voxels[cx][cy][cz] > -127 {
						n.voxels[cx][cy][cz]--
					}
				}
			}
		}
	}
	positions, normals, indices := marchTransvoxels(n.voxels)
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	n.mesh.Init(geom, n.mesh.GetMaterial(0))
}

func (n *TransvoxelNode) Expand() {
	lodMult := 1 << n.level
	size := lodMult * 16
	var children [8]*TransvoxelNode
	for i := 0; i < 8; i++ {
		o := NeighborOffsets[i]
		ox, oy, oz := int(o[0]), int(o[1]), int(o[2])
		children[i] = &TransvoxelNode{
			level: n.level - 1,
			x:     n.x + ox*size/2,
			y:     n.y + oy*size/2,
			z:     n.z + oz*size/2,
		}
		if children[i].level > 0 {
			children[i].Expand()
		}
	}
	n.children = &children
}

func (n *TransvoxelNode) Render(scene *core.Node) {
	if n.children == nil {
		n.render(scene)
		return
	}
	for i, c := range n.children {
		if i == 0 {
			c.Render(scene)
		} else {
			c.render(scene)
		}
	}
}

func (n *TransvoxelNode) Cast(caster *collision.Raycaster) (*TransvoxelNode, []collision.Intersect) {
	size := float32(int(1) << n.level * 16)
	minX, minY, minZ := float32(n.x), float32(n.y), float32(n.z)
	if !caster.IsIntersectionBox(&math32.Box3{Min: math32.Vector3{minX, minY, minZ}, Max: math32.Vector3{minX + size, minY + size, minZ + size}}) {
		return nil, nil
	}
	if n.mesh != nil {
		return n, caster.IntersectObject(n.mesh, false)
	}
	if n.children != nil {
		for _, c := range n.children {
			if hit, intersects := c.Cast(caster); hit != nil {
				return hit, intersects
			}
		}
	}
	return n, nil
}

func (n *TransvoxelNode) render(scene *core.Node) {
	lodMult := 1 << n.level
	noise := opensimplex.NewNormalized32(0)
	var voxels [19][19][19]int8
	if n.y < 16 {
		for ox := 0; ox < 19; ox++ {
			for oz := 0; oz < 19; oz++ {
				height := 10*octaveNoise(noise, 16, float32(n.x-1+ox*lodMult), 0, float32(n.z-1+oz*lodMult), .5, 0.09) + 1
				height /= float32(lodMult)
				density := float32(-127)
				deltaDensity := 256 / height
				for oy := 1; oy < int(height); oy++ {
					voxels[ox][oy][oz] = int8(density)
					density += deltaDensity
				}
			}
		}
	}
	n.voxels = &voxels
	positions, normals, indices := marchTransvoxels(n.voxels)
	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)
	mat := material.NewMaterial()
	mat.AddTexture(textures["rock"])
	mat.AddTexture(textures["grass2"])
	mat.AddTexture(textures["grass"])
	mat.SetShader("terrain")
	n.mesh = graphic.NewMesh(geom, mat)
	n.mesh.SetPosition(-1, -1, -1)
	scene.Remove(n.group)
	n.group = core.NewNode()
	n.group.Add(n.mesh)

	bb := graphic.NewMesh(geometry.NewBox(16, 16, 16), WireframeMaterial)
	bb.SetPosition(8, 8, 8)
	n.group.Add(bb)

	n.group.SetScale(float32(lodMult), float32(lodMult), float32(lodMult))
	n.group.SetPosition(float32(n.x), float32(n.y), float32(n.z))
	scene.Add(n.group)
}

type TransvoxelTerrainScene struct {
	*core.Node

	root            *TransvoxelNode
	cam             *camera.Camera
	orbitCam, fpCam *camera.Camera
	orbitControl    *camera.OrbitControl
	pitch, yaw      float32
	mouseX, mouseY  float32
	mouseDown       bool
	fpsLabel        *gui.Label

	velocity *math32.Vector3
}

func NewTransvoxelTerrainScene() *TransvoxelTerrainScene {
	scene := core.NewNode()

	root := &TransvoxelNode{level: 3, x: 0, y: 0, z: 0}
	root.Expand()
	root.Render(scene)

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
	fpCam.SetPosition(0, 0, 40)
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
		root:         root,
		cam:          orbitCam,
		orbitCam:     orbitCam,
		orbitControl: orbitControl,
		fpCam:        fpCam,
		fpsLabel:     fpsLabel,
		mouseX:       float32(width) / 2,
		mouseY:       float32(height) / 2,
		velocity:     &math32.Vector3{},
	}

	a.SubscribeID(window.OnKeyDown, a, s.OnKeyPress)
	a.SubscribeID(window.OnKeyRepeat, a, s.OnKeyPress)
	a.SubscribeID(window.OnCursor, a, s.OnMouseMove)
	a.SubscribeID(window.OnMouseDown, a, s.OnMouseDown)
	a.SubscribeID(window.OnMouseUp, a, s.OnMouseUp)
	return s
}

func (s *TransvoxelTerrainScene) OnKeyPress(evname string, ev interface{}) {
	e := ev.(*window.KeyEvent)
	if s.cam == s.orbitCam {
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

func (s *TransvoxelTerrainScene) OnMouseDown(evname string, ev interface{}) {
	s.mouseDown = true
}

func (s *TransvoxelTerrainScene) OnMouseUp(evname string, ev interface{}) {
	s.mouseDown = false
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
	}

	if s.mouseDown {
		width, height := a.GetSize()
		x := 2*(s.mouseX/float32(width)) - 1
		y := -2*(s.mouseY/float32(height)) + 1
		caster := collision.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})
		caster.SetFromCamera(s.cam, x, y)
		_, intersects := s.root.Cast(caster)
		log.Println(intersects)
	}

	a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	if err := renderer.Render(s, s.cam); err != nil {
		panic(err)
	}
}
