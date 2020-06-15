package main

import (
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

	s := NewTransvoxelTerrainScene()
	a.Run(s.Update)
}
