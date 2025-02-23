package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Configuration Area
var (
	// output config
	generateSource Source  = Video
	chain          int     = 700000
	generator      Command = func(rgb Color, x, y, z float64, blockId string) (cmd string) {
		return fmt.Sprintf("setblock ~%.2f ~%.2f ~%.2f %s", x, y, z, blockId)
	}
	isCountBlock bool = true
	// object config
	objectRoot    = "./3d"
	objectFile    = "HatsuneMiku.obj"
	objectScale   = NewFrac(9, 5)
	objectSpacing = NewFrac(1, 1)
	// example.png
	imageFile = "../develop/assets/cbw32.png"
	// video.mp4 *ffmpeg required
	videoFile      = "./minecraft/example.mp4"
	frameRate  int = 5
	videoScale     = "128:-1" // ffmpeg rescale argument
	// minecraft config
	minecraftRoot = "./minecraft"
	acceptBlockId = []string{""}                                                    //allowed regexp
	ignoreBlockId = []string{"powder", "sand", "gravel", "glass", "spawner", "ice"} //allowed regexp
)

// Supported file format
type Source int

const (
	Object Source = iota // Supported .obj(using .mtl&.png)
	Image                // Supported .png .jpeg
	Video                // Supported .mp4
)

func main() {
	start := time.Now()
	// minecraft block
	block_start := time.Now()
	fmt.Printf("\nBlock parse start...\n")
	blockModelList := scanBlockModel()
	blockColor := blockFilter(blockModelList)
	fmt.Printf("\nBlock parse duration: %s\n", time.Since(block_start))

	// [index][]command:string{}
	generatedCommand := [][]string{}

	usedBlock := map[string]int{}

	switch generateSource {
	case Object:
		obj_start := time.Now()
		fmt.Printf("\nObject parse start...\n")
		obj, _ := os.ReadFile(filepath.Join(objectRoot, objectFile))

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
		// command
		command := []string{}

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
					polygonVectors = append(polygonVectors, [3]Frac{Float2Frac(x).Mul(objectScale), Float2Frac(y).Mul(objectScale), Float2Frac(z).Mul(objectScale)})
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

					step := getStep(polygonPa, polygonPb, polygonPc, objectSpacing)
					polygonPoints := getPolygonPoints(step, polygonPa, polygonPb, polygonPc)
					texturePoints := getTexturePoints(step, texturePa, texturePb, texturePc)

					var generateCmds []string
					for i := 0; i < len(polygonPoints); i++ {
						polygonPoint := polygonPoints[i]
						x := math.Round(polygonPoint[0].Div(objectSpacing).Float()) * objectSpacing.Float()
						y := math.Round(polygonPoint[1].Div(objectSpacing).Float()) * objectSpacing.Float()
						z := math.Round(polygonPoint[2].Div(objectSpacing).Float()) * objectSpacing.Float()

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

						generateCmds = append(generateCmds, generator(textureColor, x, y, z, blockId))
						usedBlock[blockId] = usedBlock[blockId] + 1
					}

					prefix := fmt.Sprintf("Face L%d: %s", ln, line)
					fmt.Printf("% -60s Step:%f Now:%s\n    ABC:%s,%s,%s\n", prefix, step.Float(), time.Since(obj_start), polygonPa, polygonPb, polygonPc)
					face++
					command = append(command, removeDupe(generateCmds)...)
				}
			default:
				fmt.Printf("Skip L%d: %s\n", ln, line)
			}
		}
		fmt.Printf("\nObject parse duration: %s\n", time.Since(obj_start))

		fmt.Printf("\nDuration: %s, Point:%d Face:%d\n", time.Since(start), len(polygonVectors), face)
		fmt.Printf("Min:[%.2f,%.2f,%.2f] Max:[%.2f,%.2f,%.2f] H:%.2f W:%.2f D:%.2f\n", min[0], min[1], min[2], max[0], max[1], max[2], max[0]-min[0], max[1]-min[1], max[2]-min[2])
		generatedCommand = append(generatedCommand, command)

	case Image:
		image_start := time.Now()
		fmt.Printf("\nImage parse start...\n")

		// command
		command := []string{}

		var W, H float64
		for _, p := range parseImage(imageFile) {
			blockId := nearestColorBlock(p.c, blockColor)

			command = append(command, generator(p.c, p.x, p.y, 0.0, blockId))
			W = math.Max(W, p.x)
			H = math.Max(H, p.y)
			usedBlock[blockId] = usedBlock[blockId] + 1
		}
		fmt.Printf("\nImage parse duration: %s\n", time.Since(image_start))

		fmt.Printf("\nDuration: %s W:%d W:%d\n", time.Since(start), int(W), int(H))
		generatedCommand = append(generatedCommand, command)

	case Video:
		video_start := time.Now()
		fmt.Printf("\nVideo parse start...\n")

		// Get video duration
		var duration float64
		fmt.Printf("\nGet video duration start...\n")
		out, _ := exec.Command("ffmpeg", "-i", videoFile).CombinedOutput()
		for _, line := range strings.Split(string(out), "\n") {
			// 動画時間入手
			if strings.Contains(line, "Duration") {
				line = regexp.MustCompile(".*([0-9]{2}):([0-9]{2}):([0-9]{2}).*").ReplaceAllString(line, "$1 $2 $3")
				var hour, min, sec int
				fmt.Sscanf(line, "%d %d %d", &hour, &min, &sec)
				duration = float64(hour*3600 + min*60 + sec - 1)
				break
			}
		}
		fmt.Printf("Video duration: %f", duration)
		fmt.Printf("\nGet video duration: %s\n", time.Since(video_start))

		fmt.Printf("\nFrame to function start...\n")
		var W, H float64
		var frame = 0
		for current := 0.0; current < duration; current += 1.0 / float64(frameRate) {
			frame++
			fmt.Printf("Now: %5.2f/%5.2f, F: %d\n", current, duration, frame)
			exec.Command("ffmpeg", "-y", "-ss", fmt.Sprintf("%.3f", current), "-i", videoFile, "-frames:v", "1", "-vf", "scale="+videoScale, "./temp.png").Run()

			// command
			command := []string{}

			for _, p := range parseImage("./temp.png") {
				blockId := nearestColorBlock(p.c, blockColor)

				command = append(command, generator(p.c, p.x, p.y, 0.0, blockId))
				W = math.Max(W, p.x)
				H = math.Max(H, p.y)
				usedBlock[blockId] = usedBlock[blockId] + 1
			}

			generatedCommand = append(generatedCommand, command)
		}
		fmt.Printf("\nFrame to function duration: %s\n", time.Since(video_start))

		fmt.Printf("\nDuration: %s, Frame: %d, W: %d, H: %d,\n", time.Since(start), len(generatedCommand), int(W), int(H))
	}

	fmt.Printf("\nCreate function...\n")
	create_start := time.Now()
	var totalFunctions, totalCommand int
	for i, command := range generatedCommand {
		functions, commandCount := CommandToMCfunction(command, fmt.Sprintf("f%04d-i", i), chain)
		totalFunctions += len(functions)
		totalCommand += commandCount

		for _, f := range functions {
			fmt.Printf("%s.mcfunction\n", f)
		}
	}
	fmt.Printf("Total generated command/function: %d/%d\n", totalCommand, totalFunctions)
	fmt.Printf("\nCreate function duration: %s\n", time.Since(create_start))

	if isCountBlock {
		fmt.Printf("\nBlock information:\n")
		type Count struct {
			BlockID string `json:"blockID"`
			Count   int    `json:"count"`
		}
		var blockCount []Count
		for blockID, count := range usedBlock {
			blockCount = append(blockCount, Count{
				BlockID: blockID,
				Count:   count,
			})
		}

		slices.SortFunc(blockCount, func(a, b Count) int {
			return b.Count - a.Count // 降順にソート
		})
		for i, v := range blockCount {
			fmt.Printf("% 4d: %-5s %d\n", i+1, v.BlockID, v.Count)
		}
	}

	fmt.Printf("\nFinished program: %s\n", time.Since(start))
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
