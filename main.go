package main

import (
	"flag"
)

var (
	demo = flag.String("demo", "heightmap", "heightmap, marching_cubes, or transvoxel")
)

func main() {
	switch *demo {
	case "heightmap":
		heightmapTerrainDemo()
	case "marching_cubes":
		marchingCubesDemo()
	case "transvoxel":
		marchingCubesDemo()
	}
}
