package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

func parseMtl(fileName string) map[string][][]Color {
	mtl, err := os.ReadFile(filepath.Join(objectRoot, fileName))
	if err != nil {
		panic(err)
	}

	// map[materialName][x][y]Color
	material := map[string][][]Color{}
	currentMaterial := ""

	for ln, line := range strings.Split(string(mtl), "\n") {
		cmd := strings.SplitN(line, " ", 2)
		if len(cmd) < 2 {
			continue
		}

		data := cmd[1]

		switch cmd[0] {
		case "newmtl": // Material Name
			{
				fmt.Printf("New L%d: %s\n", ln, line)
				currentMaterial = data
			}
		case "map_Kd": // Material texture file
			{
				fmt.Printf("Texture L%d: %s =>%s\n", ln, line, currentMaterial)

				texture, err := os.Open(filepath.Join(objectRoot, data))
				if err != nil {
					panic(err)
				}
				defer texture.Close()

				img, _, _ := image.Decode(texture)
				bounds := img.Bounds()

				// scan image
				var colorMap [][]Color
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					yColors := []Color{}
					for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
						r, g, b, _ := img.At(x, y).RGBA()
						yColors = append(yColors, Color{
							r: int(r >> 8),
							g: int(g >> 8),
							b: int(b >> 8),
						})
					}
					colorMap = append(colorMap, yColors)
				}

				material[currentMaterial] = colorMap
			}
		default:
			fmt.Printf("Skip L%d: %s\n", ln, line)
		}
	}

	return material
}
