package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/g3n/engine/app"
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
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/window"
)

type CubesChunk struct {
	core.INode

	voxels                 [][][]int8
	indices                map[int]int
	defaultMaterial        *material.Standard
	highlightMaterial      *material.Standard
	shiftHighlightMaterial *material.Standard
}

func (c *CubesChunk) HandleVoxelClick(x, y, z int, shift bool) {
	for _, child := range c.Children() {
		child.(*graphic.Mesh).SetMaterial(c.defaultMaterial)
	}
	m, l := len(c.voxels[0]), len(c.voxels[0][0])
	cubeMesh := c.Children()[c.indices[x+y*m+z*m*l]].(*graphic.Mesh)
	if shift {
		cubeMesh.SetMaterial(c.shiftHighlightMaterial)
	} else {
		cubeMesh.SetMaterial(c.highlightMaterial)
	}
}

func NewCubesChunk(voxels [][][]int8) *CubesChunk {
	mat := material.NewStandard(math32.NewColor("Gray"))
	cubes := core.NewNode()
	indices := make(map[int]int)
	n, m, l := len(voxels), len(voxels[0]), len(voxels[0][0])
	for x := 0; x < n; x++ {
		for y := 0; y < m; y++ {
			for z := 0; z < l; z++ {
				if voxels[x][y][z] < 0 {
					cube := geometry.NewCube(1)
					mesh := graphic.NewMesh(cube, mat)
					mesh.SetPosition(float32(x), float32(y), float32(z))
					indices[x+y*m+z*m*l] = len(cubes.Children())
					cubes.Add(mesh)
				}
			}
		}
	}
	cubes.SetPosition(-float32(len(voxels))/2+0.5, -1, -float32(len(voxels[0][0]))/2+0.5)
	return &CubesChunk{
		cubes, voxels, indices,
		mat, material.NewStandard(math32.NewColor("Red")), material.NewStandard(math32.NewColor("Blue"))}
}

func countVertices(n core.INode) int {
	sum := 0
	if m, ok := n.(*graphic.Mesh); ok {
		sum += len(*m.GetGeometry().VBO(gls.VertexPosition).Buffer())
	}
	for _, child := range n.Children() {
		sum += countVertices(child)
	}
	return sum
}

type VoxelVizualizationChunk struct {
	core.INode
}

func (c *VoxelVizualizationChunk) HandleVoxelClick(x, y, z int, shift bool) {

}

func visualizeVoxels(voxels [][][]int8) *VoxelVizualizationChunk {
	font, err := text.NewFont("fonts/arial.ttf")
	if err != nil {
		panic(err)
	}
	font.SetPointSize(14)
	font.SetDPI(90)
	font.SetFgColor(math32.NewColor4("White"))

	wireframeMat := material.NewStandard(math32.NewColor("White"))
	wireframeMat.SetWireframe(true)
	wireframe := core.NewNode()
	labels := core.NewNode()
	for x := 0; x < len(voxels); x++ {
		for y := 0; y < len(voxels[x]); y++ {
			for z := 0; z < len(voxels[x][y]); z++ {
				wire := graphic.NewMesh(geometry.NewCube(1), wireframeMat)
				wire.SetPosition(float32(x)-0.5, float32(y)-0.5, float32(z)-0.5)
				wireframe.Add(wire)

				textImg := font.DrawText(fmt.Sprintf("%d", voxels[x][y][z]))
				tex := texture.NewTexture2DFromRGBA(textImg)
				textMat := material.NewStandard(math32.NewColor("White"))
				textMat.AddTexture(tex)
				textMat.SetTransparent(true)
				label := graphic.NewSprite(float32(textImg.Bounds().Dx())/100, float32(textImg.Bounds().Dy())/100, textMat)
				label.SetPosition(float32(x), float32(y), float32(z))
				labels.Add(label)
			}
		}
	}
	wireframe.SetPosition(-float32(len(voxels))/2+1, 0, -float32(len(voxels[0][0]))/2+1)
	labels.SetPosition(-float32(len(voxels))/2+1, 0, -float32(len(voxels[0][0]))/2+1)
	group := core.NewNode()
	group.Add(wireframe)
	group.Add(labels)
	return &VoxelVizualizationChunk{group}
}

type VoxelChunk interface {
	HandleVoxelClick(x, y, z int, shift bool)
}

func voxelDemo() {

	a := app.App()
	scene := core.NewNode()
	gui.Manager().Set(scene)

	voxels := simplexTerrain(32, 32, 32)

	index := uint8(0)
	group := core.NewNode()
	group.Add(transvoxelCase(index))
	group.Add(marchingCubesCase(index))
	group.Add(NewCubesChunk(voxels))
	group.Add(NewMarchingCubesChunk(voxels))
	group.Add(NewTransvoxelChunk(voxels))
	for i, c := range group.Children() {
		if i != 0 {
			c.SetVisible(false)
		}
	}
	scene.Add(group)

	fpsLabel := gui.NewLabel("FPS:")
	fpsLabel.SetPosition(10, 10)
	fpsLabel.SetFontSize(14)
	fpsLabel.SetColor(math32.NewColor("White"))
	scene.Add(fpsLabel)

	vertexCountLabel := gui.NewLabel(fmt.Sprintf("Vertices: %d", countVertices(group.Children()[0])))
	vertexCountLabel.SetPosition(10, 30)
	vertexCountLabel.SetFontSize(14)
	vertexCountLabel.SetColor(math32.NewColor("White"))
	scene.Add(vertexCountLabel)

	log.Println("Lights")
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	sphere := graphic.NewMesh(geometry.NewSphere(1, 10, 10), material.NewStandard(math32.NewColor("White")))
	sphere.SetPosition(-40, 40, -40)
	scene.Add(sphere)
	pointLight := light.NewPoint(math32.NewColor("White"), 1000.0)
	pointLight.SetPosition(-40, 40, -40)
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
		if e.Key == window.KeySpace {
			var target int
			for i, c := range group.Children() {
				if strings.HasPrefix(c.Name(), "Transvoxel Case") {
					target = i
					break
				}
			}
			group.RemoveAt(target)
			index++
			group.AddAt(target, transvoxelCase(index))
		}
		if e.Key == window.KeyTab {
			for i, c := range group.Children() {
				if c.Visible() {
					c.SetVisible(false)
					if i+1 >= len(group.Children()) {
						i = -1
					}
					group.Children()[i+1].SetVisible(true)
					vertexCountLabel.SetText(fmt.Sprintf("Vertices: %d", countVertices(group.Children()[i+1])))
					break
				}
			}
		}
		var offset *math32.Vector3
		if e.Key == window.KeyW {
			offset = &math32.Vector3{-5, 0, 0}
		}
		if e.Key == window.KeyA {
			offset = &math32.Vector3{0, 0, 5}
		}
		if e.Key == window.KeyS {
			offset = &math32.Vector3{5, 0, 0}
		}
		if e.Key == window.KeyD {
			offset = &math32.Vector3{0, 0, -5}
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

	caster := collision.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})
	mouseX, mouseY := float32(0), float32(0)
	a.SubscribeID(window.OnCursor, a, func(evname string, ev interface{}) {
		e := ev.(*window.CursorEvent)
		mouseX = e.Xpos
		mouseY = e.Ypos
	})

	mouseDown := false
	shift := false
	a.SubscribeID(window.OnMouseDown, a, func(evname string, ev interface{}) {
		mouseDown = true
		shift = ev.(*window.MouseEvent).Button == 1
	})
	a.SubscribeID(window.OnMouseUp, a, func(evname string, ev interface{}) {
		mouseDown = false
		shift = false
	})

	log.Println("Action!")
	frames := 0
	t := time.Now()
	a.Run(func(renderer *renderer.Renderer, _ time.Duration) {
		frames++
		elapsed := time.Now().Sub(t).Seconds()
		fpsLabel.SetText(fmt.Sprintf("FPS: %.0f", float64(frames)/elapsed))
		if elapsed >= 1 {
			frames = 0
			t = time.Now()
		}

		if mouseDown {
			width, height := a.GetSize()
			x := 2*(mouseX/float32(width)) - 1
			y := -2*(mouseY/float32(height)) + 1

			var curr core.INode
			for _, c := range group.Children() {
				if c.Visible() {
					curr = c
					break
				}
			}
			caster.SetFromCamera(cam, x, y)
			intersects := caster.IntersectObject(curr, true)
			if len(intersects) > 0 {
				firstHit := intersects[0]
				pos := curr.Position()
				voxelX := int(math32.Round(firstHit.Point.X - pos.X))
				voxelY := int(math32.Round(firstHit.Point.Y - pos.Y))
				voxelZ := int(math32.Round(firstHit.Point.Z - pos.Z))
				if c, ok := curr.(VoxelChunk); ok {
					c.HandleVoxelClick(voxelX, voxelY, voxelZ, shift)
				}
			}
		}

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		if err := renderer.Render(scene, cam); err != nil {
			panic(err)
		}
	})
}
