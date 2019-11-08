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

// TerrainMaterial can render and manipulate a heightmap.
type TerrainMaterial struct {
	material.Standard
	heightTexture         *texture.Texture2D
	nextHeightTexture     *texture.Texture2D
	heightmap             []float32
	uniformCameraPosition gls.Uniform
	uniformBrushPosition  gls.Uniform
	uniformBrushSize      gls.Uniform

	CameraPosition math32.Vector3
	Size           int
	BrushPosition  math32.Vector2
	BrushSize      float32
	BrushType      string
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
	noise := opensimplex.NewNormalized32(0)
	for y := 0; y < m.Size; y++ {
		for x := 0; x < m.Size; x++ {
			m.heightmap = append(m.heightmap, math32.Max(0, 100*octaveNoise(noise, 16, float32(x), float32(y), .5, 0.007)-20))
			m.heightmap = append(m.heightmap, 0)
			m.heightmap = append(m.heightmap, 0)
			m.heightmap = append(m.heightmap, 0)
		}
	}
	m.heightTexture = texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, m.heightmap)
	m.nextHeightTexture = texture.NewTexture2DFromData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, m.heightmap)
	m.AddTexture(m.heightTexture)
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
	return m
}

// Raise the area under the brush.
func (m *TerrainMaterial) Raise() {
	radius := int(m.BrushSize * float32(m.Size))
	bx := int(m.BrushPosition.X * float32(m.Size))
	by := int(m.BrushPosition.Y * float32(m.Size))
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if math32.Sqrt(float32(x*x+y*y)) < float32(radius) {
				i := (by+y)*m.Size*4 + (bx+x)*4
				if i >= 0 && i < len(m.heightmap) {
					m.heightmap[i] += 0.1
				}
			}
		}
	}
	m.heightTexture.SetData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, m.heightmap)
}

// Lower the area under the brush.
func (m *TerrainMaterial) Lower() {
	radius := int(m.BrushSize * float32(m.Size))
	bx := int(m.BrushPosition.X * float32(m.Size))
	by := int(m.BrushPosition.Y * float32(m.Size))
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if math32.Sqrt(float32(x*x+y*y)) < float32(radius) {
				i := (by+y)*m.Size*4 + (bx+x)*4
				if i >= 0 && i < len(m.heightmap) {
					m.heightmap[i] -= 0.1
					m.heightmap[i] = math32.Max(0, m.heightmap[i])
				}
			}
		}
	}
	m.heightTexture.SetData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, m.heightmap)
}

// Water the area under the brush.
func (m *TerrainMaterial) Water() {
	radius := int(m.BrushSize * float32(m.Size))
	bx := int(m.BrushPosition.X * float32(m.Size))
	by := int(m.BrushPosition.Y * float32(m.Size))
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if math32.Sqrt(float32(x*x+y*y)) < float32(radius) {
				i := (by+y)*m.Size*4 + (bx+x)*4 + 1
				if i >= 0 && i < len(m.heightmap) {
					m.heightmap[i] += 0.1
				}
			}
		}
	}
	m.heightTexture.SetData(m.Size, m.Size, gls.RGBA, gls.FLOAT, gls.RGBA32F, m.heightmap)
}

// RenderSetup is called before rendering a mesh with the material.
// It updates shader uniform variables with their Go values.
func (m *TerrainMaterial) RenderSetup(gl *gls.GLS) {
	m.Standard.RenderSetup(gl)
	gl.Uniform3f(m.uniformCameraPosition.Location(gl), m.CameraPosition.X, m.CameraPosition.Y, m.CameraPosition.Z)
	gl.Uniform2f(m.uniformBrushPosition.Location(gl), m.BrushPosition.X, m.BrushPosition.Y)
	gl.Uniform1f(m.uniformBrushSize.Location(gl), m.BrushSize)
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
	a.Renderer().AddProgram("compute", "compute.vert", "compute.frag")

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

	geom := geometry.NewSegmentedPlane(128, 128, 128, 128)
	terrain := NewTerrainMaterial(128)
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

	btn := gui.NewButton("Raise")
	btn.SetSize(40, 40)
	btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		terrain.BrushType = "Raise"
		gui.Manager().SetKeyFocus(gui.Manager())
	})
	dock.Add(btn)
	btn = gui.NewButton("Lower")
	btn.SetSize(40, 40)
	btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		terrain.BrushType = "Lower"
		gui.Manager().SetKeyFocus(gui.Manager())
	})
	dock.Add(btn)
	btn = gui.NewButton("Water")
	btn.SetSize(40, 40)
	btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		terrain.BrushType = "Water"
		gui.Manager().SetKeyFocus(gui.Manager())
	})
	dock.Add(btn)

	yellow := math32.ColorName("yellow")
	orb := graphic.NewMesh(
		geometry.NewSphere(1, 100, 100),
		material.NewStandard(&yellow),
	)
	scene.Add(orb)

	dir := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.9)
	dir.SetPosition(10, 50, 0)
	dir.SetDirection(0, -1, 0)
	orb.SetPosition(10, 50, 0)
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

	var mouseDown bool
	a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
		e := ev.(*window.MouseEvent)
		if e.Button == window.MouseButton1 {
			mouseDown = true
			gui.Manager().SetCursorFocus(gui.Manager())
		}
	})

	a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
		e := ev.(*window.MouseEvent)
		if e.Button == window.MouseButton1 {
			mouseDown = false
			gui.Manager().SetCursorFocus(nil)
		}
	})

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
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		// Update camera position
		terrain.CameraPosition = cam.Position()

		// Compute shader pass
		m.nextHeightTexture.Render(a.Gls())
		a.Gls().WithTexture(m.nextHeightTexture, func() {
			a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
			if err := renderer.Render(scene, cam); err != nil {
				panic(err)
			}
		})

		// Color-based mouse-picking shader pass
		a.Gls().WithFramebuffer(func() {
			orb.SetVisible(false)
			terrain.SetShader("color")
			a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
			if err := renderer.Render(scene, cam); err != nil {
				panic(err)
			}
			c := a.Gls().ReadPixels(int(mouseX), int(mouseY), 1, 1)[0][0]
			if c.R == 0.5 && c.G == 0.5 && c.B == 0.5 {
				return
			}
			terrain.BrushPosition.X = c.R
			terrain.BrushPosition.Y = c.G
		})

		// Standard render pass
		orb.SetVisible(true)
		terrain.SetShader("terrain")
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}

		// User interaction
		if mouseDown {
			switch terrain.BrushType {
			case "Raise":
				terrain.Raise()
			case "Lower":
				terrain.Lower()
			case "Water":
				terrain.Water()
			}
		}
	})
}
