package main

import (
	"flag"
	"fmt"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
)

var (
	demo = flag.String("demo", "heightmap", "heightmap, voxels")
	a    *app.Application
	font *text.Font
)

func main() {
	flag.Parse()

	var err error
	font, err = text.NewFont("fonts/arial.ttf")
	if err != nil {
		panic(err)
	}
	font.SetPointSize(14)
	font.SetDPI(90)
	font.SetFgColor(math32.NewColor4("White"))

	a = app.App()

	switch *demo {
	case "heightmap":
		NewHeightmapTerrainDemoScene()
	case "voxels":
		NewVoxelDemoScene()
	case "transvoxel":
		s := NewTransvoxelTerrainScene()
		a.Run(s.Update)
	default:
		panic(fmt.Sprintf("Invalid demo %q", *demo))
	}
}
