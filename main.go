package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/ojrac/opensimplex-go"

	"github.com/g3n/engine/app"
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
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/window"
)

// BrushType can modify the terrain in different ways.
type BrushType int32

// Brush Types
const (
	None    = BrushType(0)
	Raise   = BrushType(1)
	Lower   = BrushType(2)
	Water   = BrushType(3)
	Smooth  = BrushType(4)
	Flatten = BrushType(5)
)

// TerrainMaterial can render and manipulate a heightmap.
type TerrainMaterial struct {
	material.Standard
	uniformCameraPosition gls.Uniform
	uniformBrushPosition  gls.Uniform
	uniformBrushSize      gls.Uniform
	uniformBrushType      gls.Uniform
	uniformMouseButton    gls.Uniform

	CameraPosition math32.Vector3
	Size           int
	BrushPosition  math32.Vector2
	BrushSize      float32
	BrushType      BrushType
	MouseButton    int32
}

// NewTerrainMaterial initializes a terrain with OpenSimplex noise.
func NewTerrainMaterial(size int) *TerrainMaterial {
	m := &TerrainMaterial{
		Size:      size,
		BrushSize: 0.1,
	}
	blue := math32.ColorName("blue")
	m.Init("terrain", &blue)
	m.uniformCameraPosition.Init("CameraPosition")
	m.uniformBrushPosition.Init("BrushPosition")
	m.uniformBrushSize.Init("BrushSize")
	m.uniformBrushType.Init("BrushType")
	m.uniformMouseButton.Init("MouseButton")
	noise := opensimplex.NewNormalized32(0)
	var heightmap []float32
	for y := 0; y < m.Size; y++ {
		for x := 0; x < m.Size; x++ {
			height := math32.Max(0, 100*octaveNoise(noise, 16, float32(x), float32(y), .5, 0.007)-20)
			heightmap = append(heightmap, height)
			heightmap = append(heightmap, 0)
			heightmap = append(heightmap, 0)
			heightmap = append(heightmap, 0)
		}
	}
	heightTexture := texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, heightmap)
	nextHeightTexture := texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, nil)
	flowTexture := texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, nil)
	nextFlowTexture := texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, nil)
	heightTexture.SetMagFilter(gls.NEAREST)
	heightTexture.SetMinFilter(gls.NEAREST)
	nextHeightTexture.SetMagFilter(gls.NEAREST)
	nextHeightTexture.SetMinFilter(gls.NEAREST)
	m.AddTexture(heightTexture)
	m.AddTexture(nextHeightTexture)
	for _, f := range []string{"water", "dirt", "grass", "grass2", "rock", "snow"} {
		tex, err := texture.NewTexture2DFromImage(fmt.Sprintf("textures/%s.jpg", f))
		if err != nil {
			panic(err)
		}
		tex.SetRepeat(float32(m.Size)/2, float32(m.Size)/2)
		tex.SetWrapS(gls.REPEAT)
		tex.SetWrapT(gls.REPEAT)
		m.AddTexture(tex)
	}
	m.AddTexture(flowTexture)
	m.AddTexture(nextFlowTexture)
	return m
}

// RenderSetup is called before rendering a mesh with the material.
// It updates shader uniform variables with their Go values.
func (m *TerrainMaterial) RenderSetup(gl *gls.GLS) {
	m.Standard.RenderSetup(gl)
	gl.Uniform3f(m.uniformCameraPosition.Location(gl), m.CameraPosition.X, m.CameraPosition.Y, m.CameraPosition.Z)
	gl.Uniform2f(m.uniformBrushPosition.Location(gl), m.BrushPosition.X, m.BrushPosition.Y)
	gl.Uniform1f(m.uniformBrushSize.Location(gl), m.BrushSize)
	gl.Uniform1i(m.uniformBrushType.Location(gl), int32(m.BrushType))
	gl.Uniform1i(m.uniformMouseButton.Location(gl), m.MouseButton)
}

func octaveNoise(noise opensimplex.Noise32, iters int, x, y float32, persistence, scale float32) float32 {
	var maxamp float32 = 0
	var amp float32 = 1
	freq := scale
	var value float32 = 0

	for i := 0; i < iters; i++ {
		value += noise.Eval2(x*freq, y*freq) * amp
		maxamp += amp
		amp *= persistence
		freq *= 2
	}

	return value / maxamp
}

func main() {

	// Create application and scene
	a := app.App()

	files, err := ioutil.ReadDir("shaders")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		b, err := ioutil.ReadFile(fmt.Sprintf("shaders/%s", f.Name()))
		if err != nil {
			panic(err)
		}
		a.Renderer().AddShader(f.Name(), string(b))
	}
	a.Renderer().AddProgram("terrain", "terrain.vert", "terrain.frag")
	a.Renderer().AddProgram("color", "terrain.vert", "color.frag")
	a.Renderer().AddProgram("apply_brush", "compute.vert", "apply_brush.frag")
	a.Renderer().AddProgram("compute_flow", "compute.vert", "compute_flow.frag")
	a.Renderer().AddProgram("compute_deltav", "compute.vert", "compute_deltav.frag")
	a.Renderer().AddProgram("update_water", "compute.vert", "update_water.frag")

	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 100)
	scene.Add(cam)

	// Get framebuffer size and update viewport accordingly
	width, height := a.GetFramebufferSize()
	a.Gls().Viewport(0, 0, int32(width), int32(height))
	// Update the camera's aspect ratio
	cam.SetAspect(float32(width) / float32(height))

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	terrain := NewTerrainMaterial(128)
	geom := geometry.NewSegmentedPlane(float32(terrain.Size), float32(terrain.Size), terrain.Size, terrain.Size)
	mesh := graphic.NewMesh(geom, terrain)
	mesh.RotateX(-90 * math.Pi / 180)
	scene.Add(mesh)

	ui := gui.NewPanel(0, 50)
	ui.SetPaddings(4, 4, 4, 4)
	ui.SetLayout(gui.NewDockLayout())
	ui.SetColor4(&math32.Color4{G: 1.0, A: 0.4})
	ui.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockBottom})
	scene.Add(ui)

	dock := gui.NewPanel(0, 50)
	layout := gui.NewHBoxLayout()
	dock.SetLayout(layout)
	layout.SetSpacing(10)
	ui.Add(dock)

	brushes := []struct {
		Type BrushType
		Name string
	}{
		{Raise, "Raise"},
		{Lower, "Lower"},
		{Water, "Water"},
		{Smooth, "Smooth"},
		{Flatten, "Flatten"},
	}

	onClick := func(brushType BrushType) func(string, interface{}) {
		return func(name string, ev interface{}) {
			terrain.BrushType = brushType
			gui.Manager().SetKeyFocus(gui.Manager())
		}
	}

	for _, brush := range brushes {
		btn := gui.NewButton(brush.Name)
		btn.SetSize(40, 40)
		btn.Subscribe(gui.OnClick, onClick(brush.Type))
		dock.Add(btn)
	}

	dir := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.9)
	dir.SetPosition(10, 50, 0)
	dir.SetDirection(0, -1, 0)
	scene.Add(dir)
	scene.Add(light.NewAmbient(&math32.Color{R: 1, G: 1, B: 1}, 0.8))

	// Set background color to gray
	a.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	var mouseX, mouseY float32
	a.SubscribeID(window.OnCursor, a, func(evname string, ev interface{}) {
		e := ev.(*window.CursorEvent)
		scaleX, scaleY := a.GetScale()
		mouseX = e.Xpos * float32(scaleX)
		mouseY = e.Ypos * float32(scaleY)
	})

	a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
		e := ev.(*window.MouseEvent)
		if e.Button == window.MouseButton1 {
			terrain.MouseButton = 1
			gui.Manager().SetCursorFocus(gui.Manager())
		}
	})

	a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
		e := ev.(*window.MouseEvent)
		if e.Button == window.MouseButton1 {
			terrain.MouseButton = 0
			gui.Manager().SetCursorFocus(nil)
		}
	})

	mousePickFramebuffer := a.Gls().GenerateFramebuffer()
	colorBuf := a.Gls().GenerateRenderbuffer()
	depthBuf := a.Gls().GenerateRenderbuffer()
	a.Gls().BindFramebuffer(mousePickFramebuffer)
	a.Gls().BindRenderbuffer(colorBuf)
	a.Gls().RenderbufferStorage(gls.RGBA32F, width, height)
	a.Gls().FramebufferRenderbuffer(gls.COLOR_ATTACHMENT0, colorBuf)
	a.Gls().BindRenderbuffer(depthBuf)
	a.Gls().RenderbufferStorage(gls.DEPTH_COMPONENT16, width, height)
	a.Gls().FramebufferRenderbuffer(gls.DEPTH_ATTACHMENT, depthBuf)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetFramebufferSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))

		// Update UI size
		dock.SetWidth(float32(width))
		ui.SetWidth(float32(width))

		a.Gls().BindFramebuffer(mousePickFramebuffer)
		a.Gls().BindRenderbuffer(colorBuf)
		a.Gls().RenderbufferStorage(gls.RGBA32F, width, height)
		a.Gls().FramebufferRenderbuffer(gls.COLOR_ATTACHMENT0, colorBuf)
		a.Gls().BindRenderbuffer(depthBuf)
		a.Gls().RenderbufferStorage(gls.DEPTH_COMPONENT16, width, height)
		a.Gls().FramebufferRenderbuffer(gls.DEPTH_ATTACHMENT, depthBuf)
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	computeFramebuffer := a.Gls().GenerateFramebuffer()
	computeScene := core.NewNode()
	computeCam := camera.NewOrthographic(1.0, 0, 1.0, 2.0, camera.Vertical)
	computeScene.Add(graphic.NewMesh(geometry.NewPlane(2, 2), terrain))
	computeScene.Add(computeCam)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		width, height := a.GetFramebufferSize()
		// Update camera position
		terrain.CameraPosition = cam.Position()

		// Color-based mouse-picking shader pass
		a.Gls().BindFramebuffer(mousePickFramebuffer)
		terrain.SetShader("color")
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
		a.Gls().ReadBuffer(gls.COLOR_ATTACHMENT0)
		c := a.Gls().ReadPixels(int(mouseX), int(mouseY), 1, 1)[0][0]
		if c.R == 0.5 && c.G == 0.5 && c.B == 0.5 {
			return
		}
		terrain.BrushPosition.X = c.R
		terrain.BrushPosition.Y = c.G

		// Compute shader passes
		a.Gls().Enable(gls.COLOR_LOGIC_OP)
		a.Gls().BindFramebuffer(computeFramebuffer)
		a.Gls().Viewport(0, 0, int32(terrain.Size), int32(terrain.Size))

		a.Gls().FramebufferTexture(gls.COLOR_ATTACHMENT0, terrain.Textures[1].TexName())
		terrain.SetShader("apply_brush")
		a.Gls().Clear(gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(computeScene, computeCam); err != nil {
			panic(err)
		}
		terrain.Textures[0], terrain.Textures[1] = terrain.Textures[1], terrain.Textures[0]

		a.Gls().Viewport(0, 0, int32(width), int32(height))
		a.Gls().Disable(gls.COLOR_LOGIC_OP)

		// Standard render pass
		a.Gls().BindFramebuffer(0)
		terrain.SetShader("terrain")
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
