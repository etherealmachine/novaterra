package main

import (
	"os"
	"time"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ojrac/opensimplex-go"
)

type Scene struct {
	*core.Node

	cam *camera.Camera

	pitch, yaw     float32
	mouseX, mouseY float32

	mat *Material
}

func NewScene() *Scene {
	scene := core.NewNode()

	l := light.NewAmbient(math32.NewColor("white"), 0.1)
	scene.Add(l)

	l2 := light.NewDirectional(math32.NewColor("white"), 0.8)
	l2.SetPosition(100, 200, 400)
	scene.Add(l2)

	cam := camera.New(1)
	cam.SetPosition(60, 65, 25)
	scene.Add(cam)

	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	cam.SetAspect(float32(width) / float32(height))

	axes := helper.NewAxes(32)
	scene.Add(axes)

	//chunk := NewChunk()
	//scene.Add(chunk.Mesh())

	mat := NewMaterial()

	tree := NewTree(nil, math32.Vector3{X: 0, Y: 0, Z: 0}, 32)
	noise := opensimplex.New32(0)
	for x := 0; x < int(tree.Size); x++ {
		for z := 0; z < int(tree.Size); z++ {
			maxHeight := int(tree.Size / 2)
			height := int(((noise.Eval3(float32(x), 0, float32(z)) + 1) / 2) * float32(maxHeight))
			for y := 0; y < height; y++ {
				node := tree.At(float32(x)-tree.Size/2, float32(y)-tree.Size/2, float32(z)-tree.Size/2)
				node.Material = 1
			}
		}
	}
	treeNode := tree.Node(mat)
	treeNode.GetNode().SetPosition(tree.Size/2, tree.Size/2, tree.Size/2)
	scene.Add(treeNode)

	s := &Scene{
		Node:   scene,
		cam:    cam,
		yaw:    3.49,
		pitch:  -0.81,
		mouseX: -1,
		mouseY: -1,
		mat:    mat,
	}
	a.SubscribeID(window.OnCursor, a, s.OnMouseMove)
	a.SubscribeID(window.OnKeyDown, a, s.OnKeyDown)
	window.Get().(*window.GlfwWindow).SetInputMode(glfw.InputMode(glfw.CursorMode), glfw.CursorDisabled)
	window.Get().(*window.GlfwWindow).SetFullscreen(true)

	return s
}

func (s *Scene) OnMouseMove(evname string, ev interface{}) {
	e := ev.(*window.CursorEvent)
	if s.mouseX >= 0 || s.mouseY >= 0 {
		s.yaw -= math32.DegToRad(0.1 * (s.mouseX - e.Xpos))
		s.pitch += math32.DegToRad(0.1 * (s.mouseY - e.Ypos))
		s.pitch = math32.Clamp(s.pitch, math32.DegToRad(-89), math32.DegToRad(89))
	}
	s.mouseX = e.Xpos
	s.mouseY = e.Ypos
}

func (s *Scene) OnKeyDown(evname string, ev interface{}) {
	e := ev.(*window.KeyEvent)
	if e.Key == window.KeyEscape {
		os.Exit(0)
	} else if e.Key == window.KeyW {
		s.mat.Mode++
		if s.mat.Mode >= 3 {
			s.mat.Mode = 0
		}
	}
}

func (s *Scene) Update(renderer *renderer.Renderer, deltaTime time.Duration) {
	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	s.cam.SetAspect(float32(width) / float32(height))

	forward := &math32.Vector3{
		X: math32.Cos(s.yaw) * math32.Cos(s.pitch),
		Y: math32.Sin(s.pitch),
		Z: math32.Sin(s.yaw) * math32.Cos(s.pitch),
	}
	forward.Normalize()
	up := &math32.Vector3{X: 0, Y: 1, Z: 0}
	right := forward.Clone().Cross(up).Normalize()
	pos := s.cam.Position()
	ks := a.KeyState()
	if ks.Pressed(window.KeyUp) {
		s.cam.SetPositionVec((&pos).Add(forward))
	}
	if ks.Pressed(window.KeyDown) {
		s.cam.SetPositionVec((&pos).Add(forward.Clone().Negate()))
	}
	if ks.Pressed(window.KeyLeft) {
		s.cam.SetPositionVec((&pos).Add(right.Clone().Negate()))
	}
	if ks.Pressed(window.KeyRight) {
		s.cam.SetPositionVec((&pos).Add(right))
	}
	pos = s.cam.Position()
	s.cam.LookAt(
		(&pos).Clone().Add(forward),
		up,
	)

	a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	if err := renderer.Render(s, s.cam); err != nil {
		panic(err)
	}
}
