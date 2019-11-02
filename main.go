package main

import (
	"fmt"
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
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type TerrainMaterial struct {
	material.Standard
	terrain *texture.Texture2D
}

func NewTerrainMaterial() *TerrainMaterial {
	m := new(TerrainMaterial)
	blue := math32.ColorName("blue")
	m.Init("terrainShader", &blue)
	noise := opensimplex.NewNormalized32(0)
	var data []float32
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			data = append(data, math32.Max(0, 100*octaveNoise(noise, 16, float32(x), float32(y), .5, 0.007)-20))
			data = append(data, 0)
			data = append(data, 0)
			data = append(data, 0)
		}
	}
	m.terrain = texture.NewTexture2DFromData(128, 128, gls.RGBA, gls.FLOAT, gls.RGBA32F, data)
	m.AddTexture(m.terrain)
	for _, f := range []string{"water", "dirt", "grass", "grass2", "rock", "snow"} {
		tex, err := texture.NewTexture2DFromImage(fmt.Sprintf("textures/%s.jpg", f))
		if err != nil {
			panic(err)
		}
		tex.SetRepeat(64, 64)
		tex.SetWrapS(gls.REPEAT)
		tex.SetWrapT(gls.REPEAT)
		m.AddTexture(tex)
	}
	return m
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

	a.Renderer().AddShader("terrainVertexShader", terrainVertexShader)
	a.Renderer().AddShader("terrainFragmentShader", terrainFragmentShader)
	a.Renderer().AddProgram("terrainShader", "terrainVertexShader", "terrainFragmentShader")

	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 100)
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	geom := geometry.NewSegmentedPlane(128, 128, 128, 128)
	mat := NewTerrainMaterial()
	mesh := graphic.NewMesh(geom, mat)
	mesh.RotateX(-90 * math.Pi / 180)
	scene.Add(mesh)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{R: 1, G: 1, B: 1}, 0.8))
	pointLight := light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, 1000.0)
	pointLight.SetPosition(1, 0, 50)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	// Set background color to gray
	a.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
