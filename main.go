package main

import (
	"math"
	"math/rand"
	"time"

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
	material.Material
	terrain *texture.Texture2D
}

func NewTerrainMaterial() *TerrainMaterial {
	m := new(TerrainMaterial)
	m.Init()
	m.SetShader("terrainShader")
	var data []float32
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			data = append(data, rand.Float32())
			data = append(data, rand.Float32())
			data = append(data, rand.Float32())
			data = append(data, rand.Float32())
		}
	}
	m.terrain = texture.NewTexture2DFromData(128, 128, gls.RGBA, gls.FLOAT, gls.RGBA32F, data)
	m.AddTexture(m.terrain)
	return m
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

	// Create a blue torus and add it to the scene
	geom := geometry.NewPlane(128, 128)
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
