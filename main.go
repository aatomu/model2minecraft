package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	// output config
	scale   = NewFrac(9, 5)
	spacing = NewFrac(1, 1)
	// command            = "particle dust{color:[0f,0f,0f],scale:1f} ~%.2f ~%.2f ~%.2f 0 0 0 0 1 force @a"
	chain int = 700000
	// object config
	objectRoot = "./3d"
	objectFile = "HatsuneMiku.obj"
	// minecraft config
	minecraftRoot = "./minecraft"
	acceptBlockId = []string{""}                                                    //allowed regexp
	ignoreBlockId = []string{"powder", "sand", "gravel", "glass", "spawner", "ice"} //allowed regexp
)

func main() {
	start := time.Now()
	// minecraft block
	block_start := time.Now()
	fmt.Printf("Block parse start...\n")
	blockModelList := scanBlockModel()
	blockColor := blockFilter(blockModelList)
	fmt.Printf("Block parse duration: %s\n", time.Since(block_start))

	fmt.Printf("Object parse start...\n")
	obj_start := time.Now()
	obj, _ := os.ReadFile(filepath.Join(objectRoot, objectFile))
	commands := []string{}

	// map[materialName][x][y]Color
	var material map[string][][]Color

	// object/polygon
	polygonVectors := [][3]Frac{}
	var face int64 = 0
	min := [3]float64{}
	max := [3]float64{}
	// texture/mtl
	textureVectors := [][2]Frac{}
	currentTexture := ""
	useBlocks := map[string]int{}

	for ln, line := range strings.Split(string(obj), "\n") {
		cmd := strings.SplitN(line, " ", 2)
		if len(cmd) < 2 {
			continue
		}

		data := cmd[1]

		switch cmd[0] {
		case "mtllib":
			{
				fmt.Printf("MTL L%d: %s\n", ln, line)
				material = parseMtl(data)
			}
		case "v": // Polygon top
			{
				var x, y, z float64
				fmt.Sscanf(data, "%f %f %f", &x, &y, &z)
				polygonVectors = append(polygonVectors, [3]Frac{Float2Frac(x).Mul(scale), Float2Frac(y).Mul(scale), Float2Frac(z).Mul(scale)})
				fmt.Printf("PolygonVector L%d: %s\n", ln, line)
			}
		case "vt": // Texture top
			{
				var x, y float64
				fmt.Sscanf(data, "%f %f", &x, &y)
				textureVectors = append(textureVectors, [2]Frac{Float2Frac(x), Float2Frac(y)})
				fmt.Printf("TextureVector L%d: %s\n", ln, line)
			}
		case "usemtl": // Set use material
			{
				fmt.Printf("SetTexture L%d: %s\n", ln, line)
				currentTexture = data
			}
		case "f": // Object surface/polygon
			{
				indexes := strings.Split(data, " ")
				if len(indexes) < 3 {
					fmt.Printf("Skip L%d: %s\n", ln, line)
					continue
				}

				// Get surface polygon top
				polygonPaIndex, _ := strconv.Atoi(strings.Split(indexes[0], "/")[0])
				polygonPbIndex, _ := strconv.Atoi(strings.Split(indexes[1], "/")[0])
				polygonPcIndex, _ := strconv.Atoi(strings.Split(indexes[2], "/")[0])
				polygonPa := polygonVectors[polygonPaIndex-1]
				polygonPb := polygonVectors[polygonPbIndex-1]
				polygonPc := polygonVectors[polygonPcIndex-1]
				// Get texture polygon top
				texturePaIndex, _ := strconv.Atoi(strings.Split(indexes[0], "/")[1])
				texturePbIndex, _ := strconv.Atoi(strings.Split(indexes[1], "/")[1])
				texturePcIndex, _ := strconv.Atoi(strings.Split(indexes[2], "/")[1])
				texturePa := textureVectors[texturePaIndex-1]
				texturePb := textureVectors[texturePbIndex-1]
				texturePc := textureVectors[texturePcIndex-1]

				// Get min,max polygon top
				for i := 0; i < 3; i++ {
					// min
					if min[i] > polygonPa[i].Float() {
						min[i] = polygonPa[i].Float()
					}
					if min[i] > polygonPb[i].Float() {
						min[i] = polygonPb[i].Float()
					}
					if min[i] > polygonPc[i].Float() {
						min[i] = polygonPc[i].Float()
					}
					// max
					if max[i] < polygonPa[i].Float() {
						max[i] = polygonPa[i].Float()
					}
					if max[i] < polygonPb[i].Float() {
						max[i] = polygonPb[i].Float()
					}
					if max[i] < polygonPc[i].Float() {
						max[i] = polygonPc[i].Float()
					}
				}

				step := getStep(polygonPa, polygonPb, polygonPc, spacing)
				polygonPoints := getPolygonPoints(step, polygonPa, polygonPb, polygonPc)
				texturePoints := getTexturePoints(step, texturePa, texturePb, texturePc)

				var generateCmds []string
				for i := 0; i < len(polygonPoints); i++ {
					polygonPoint := polygonPoints[i]
					x := math.Round(polygonPoint[0].Div(spacing).Float()) * spacing.Float()
					y := math.Round(polygonPoint[1].Div(spacing).Float()) * spacing.Float()
					z := math.Round(polygonPoint[2].Div(spacing).Float()) * spacing.Float()

					// Image position mapping
					//  Golang:   | Obj:
					//   0 - X+   |  Y+
					//   |        |  |
					//   Y+       |  0 - X+

					texturePoint := texturePoints[i]
					// -1..1 => -width..width
					textureX := texturePoint[0].Mod(NewFrac(1, 1)).Float()
					textureIndexX := int(textureX * float64(len(material[currentTexture])))
					// -1..1 => height..-height
					textureY := 1 - texturePoint[1].Mod(NewFrac(1, 1)).Float()
					textureIndexY := int(textureY * float64(len(material[currentTexture][textureIndexX])))
					textureColor := material[currentTexture][textureIndexX][textureIndexY]
					blockId := nearestColorBlock(textureColor, blockColor)

					generateCmds = append(generateCmds, fmt.Sprintf("setblock ~%.2f ~%.2f ~%.2f %s", x, y, z, blockId))
					useBlocks[blockId] = useBlocks[blockId] + 1
				}

				prefix := fmt.Sprintf("Face L%d: %s", ln, line)
				fmt.Printf("% -60s Step:%f Now:%s\n    ABC:%s,%s,%s\n", prefix, step.Float(), time.Since(obj_start), polygonPa, polygonPb, polygonPc)
				face++
				commands = append(commands, removeDupe(generateCmds)...)
			}
		default:
			fmt.Printf("Skip L%d: %s\n", ln, line)
		}
	}
	fmt.Printf("Object parse duration: %s", time.Since(obj_start))

	fmt.Printf("Create function...\n")
	create_start := time.Now()
	_, n := CommandToMCfunction(commands, "", chain)
	fmt.Printf("create function duration: %s\n", time.Since(create_start))

	fmt.Printf("\n\nDuration: %s, Point:%d Face:%d Cmd:%d\n", time.Since(start), len(polygonVectors), face, n)
	fmt.Printf("Min:[%.2f,%.2f,%.2f] Max:[%.2f,%.2f,%.2f] H:%.2f W:%.2f D:%.2f\n", min[0], min[1], min[2], max[0], max[1], max[2], max[0]-min[0], max[1]-min[1], max[2]-min[2])

	type Count struct {
		BlockID string `json:"blockID"`
		Count   int    `json:"count"`
	}
	var blockCount []Count
	for blockID, count := range useBlocks {
		blockCount = append(blockCount, Count{
			BlockID: blockID,
			Count:   count,
		})
	}

	slices.SortFunc(blockCount, func(a, b Count) int {
		return b.Count - a.Count // 降順にソート
	})
	for i, v := range blockCount {
		fmt.Printf("% 4d: %-5s %d 0x%02X%02X%02X\n", i+1, v.BlockID, v.Count, blockColor[v.BlockID].r, blockColor[v.BlockID].g, blockColor[v.BlockID].b)
	}
}

func removeDupe(in []string) []string {
	results := []string{}
	encountered := map[string]struct{}{}
	for _, v := range in {
		if _, ok := encountered[v]; !ok {
			encountered[v] = struct{}{}
			results = append(results, v)
		}
	}
	return results
}
