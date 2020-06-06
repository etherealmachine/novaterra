package main

import (
	"flag"
	"fmt"
)

var (
	demo = flag.String("demo", "heightmap", "heightmap, voxels")
)

func main() {
	flag.Parse()
	switch *demo {
	case "heightmap":
		heightmapTerrainDemo()
	case "voxels":
		voxelDemo()
	default:
		panic(fmt.Sprintf("Invalid demo %q", *demo))
	}
}
