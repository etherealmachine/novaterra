package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
)

var (
	a    *app.Application
	font *text.Font
)

func main() {
	var err error
	font, err = text.NewFont("fonts/arial.ttf")
	if err != nil {
		panic(err)
	}
	font.SetPointSize(14)
	font.SetDPI(90)
	font.SetFgColor(math32.NewColor4("White"))

	a = app.App()

	files, err := ioutil.ReadDir("shaders")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		b, err := ioutil.ReadFile(filepath.Join("shaders", f.Name()))
		if err != nil {
			panic(err)
		}
		a.Renderer().AddShader(f.Name(), string(b))
	}
	a.Renderer().AddProgram("terrain", "terrain.vert", "terrain.frag")

	s := NewTransvoxelTerrainScene()
	a.Run(s.Update)
}
