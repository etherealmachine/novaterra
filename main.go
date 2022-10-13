package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"
)

var (
	a        *app.Application
	font     *text.Font
	textures map[string]*texture.Texture2D
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
	a.Renderer().AddProgram("terrain", "terrain.vert", "terrain.frag", "wireframe.geom")
	a.Renderer().AddProgram("wireframe", "terrain.vert", "wireframe.frag", "wireframe.geom")

	textures = make(map[string]*texture.Texture2D)
	for _, f := range []string{"water", "dirt", "grass", "grass2", "rock", "snow"} {
		tex, err := texture.NewTexture2DFromImage(filepath.Join("textures", f) + ".jpg")
		if err != nil {
			panic(err)
		}
		tex.SetWrapS(gls.REPEAT)
		tex.SetWrapT(gls.REPEAT)
		textures[f] = tex
	}

	s := NewScene()
	a.Run(s.Update)
}
