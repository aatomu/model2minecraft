package main

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// Configuration Area
var (
	// Output Configuration
	sourceType       Source  = Object
	maxCommandChain  int     = 700000
	colorDepthBit    int     = 4 // 1-8
	commandGenerator Command = func(arg CommandArgument) (cmd string) {
		return fmt.Sprintf("setblock ~%.2f ~%.2f ~%.2f %s", arg.x, arg.y, arg.z, arg.blockId)
		// return fmt.Sprintf("particle dust{color:[%ff,%ff,%ff],scale:0.2f} ~%.2f ~%.2f ~%.2f 0 0 0 0 1 force @a", float64(rgb.r)/255, float64(rgb.g)/255, float64(rgb.b)/255, x, y, z)
	}
	enableBlockCount bool = false

	// Object Configuration
	objectDirectory   string = "./3d"
	objectFilename    string = "HatsuneMiku.obj"
	objectScale       Frac   = NewFrac(9, 5)
	objectGridSpacing Frac   = NewFrac(1, 1)
	isObjectUVYAxisUp bool   = true
	parallelLimit     int    = 500

	// Image Configuration
	imageFilename string = "../develop/assets/cbw32.png"

	// Video Configuration (*requires ffmpeg)
	videoFilename  string = "./minecraft/example.mp4"
	videoFrameRate int    = 20
	videoScaleSize string = "200:-1" // ffmpeg rescale argument

	// Minecraft Configuration
	minecraftDirectory string   = "./assets"
	allowedBlockIds    []string = []string{""}                                                       // Allowed regex patterns
	ignoredBlockIds    []string = []string{"powder$", "sand$", "gravel$", "glass", "spawner", "ice"} // Ignored regex patterns
)

// Supported file format
type Source int

const (
	Object Source = iota // Supported .obj(using .mtl&.png)
	Image                // Supported .png .jpeg
	Video                // Supported .mp4
)

// compute variables
var (
	// Parallel
	wg             sync.WaitGroup
	wgCurrentCount int
	wgTotalRoutine int
	wgSession      chan struct{} = make(chan struct{}, parallelLimit)
	mu             sync.Mutex
	// Color
	blockColor map[string]Color
	colorMap   [][][]string // Color map use: colorBitDepth < 6
	// Minecraft
	argumentList   [][]CommandArgument // [index][]command{}
	totalUsedBlock map[string]int      = make(map[string]int)
)

func main() {
	start := time.Now()
	// minecraft block
	block_start := time.Now()
	fmt.Printf("\nBlock parse start...\n")
	blockModelList := scanBlockModel()
	blockColor = blockFilter(blockModelList)
	fmt.Printf("\nBlock parse duration: %s\n", time.Since(block_start))

	// block color to color mapping
	if colorDepthBit < 6 {
		color_start := time.Now()
		fmt.Printf("\n%dBit color mapping start...\n", colorDepthBit)
		colorMaxValue := 0xff >> (8 - colorDepthBit)
		colorMap = make([][][]string, 0, colorMaxValue*colorMaxValue*colorMaxValue)
		for r := 0; r <= colorMaxValue; r++ {
			fmt.Printf("Calc R: %d/%d, G: 0..%d, B:0..%d, Now:%s\n", r, colorMaxValue, colorMaxValue, colorMaxValue, time.Since(color_start))
			// GBColorMap := make([][]string, 0, m*m)
			GBColorMap := map[int][]string{}
			for g := 0; g <= colorMaxValue; g++ {
				wg.Add(1)
				go func(fR, fG int) {
					BColorMap := make([]string, 0, colorMaxValue)
					for b := 0; b <= colorMaxValue; b++ {
						BColorMap = append(BColorMap, nearestColorBlock(Color{uint8(fR << (8 - colorDepthBit)), uint8(fG << (8 - colorDepthBit)), uint8(b << (8 - colorDepthBit))}))
					}

					mu.Lock()
					GBColorMap[fG] = BColorMap
					mu.Unlock()

					wg.Done()
				}(r, g)
			}
			wg.Wait()
			GB := [][]string{}
			for i := 0; i < len(GBColorMap); i++ {
				GB = append(GB, GBColorMap[i])
			}
			colorMap = append(colorMap, GB)
		}
		fmt.Printf("\ncolor mapping duration: %s\n", time.Since(color_start))
	} else {
		fmt.Printf("\n\nDon't use color map.\n\n")
	}

	switch sourceType {
	case Object:
		obj_start := time.Now()
		fmt.Printf("\nObject parse start...\n")
		obj, _ := os.ReadFile(filepath.Join(objectDirectory, objectFilename))

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
		args := []CommandArgument{}

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
						return
					}

					wgSession <- struct{}{}
					wg.Add(1)
					wgTotalRoutine++
					wgCurrentCount++
					go func(fLn int, fData string, fIndexes []string, fPolygonVectors [][3]Frac, fTextureVectors [][2]Frac, fTexture [][]Color, fBlockColor map[string]Color, fObj_start time.Time) {

						defer func() {
							<-wgSession
							wg.Done()
							wgCurrentCount--
						}()

						step, surfaceMin, surfaceMax, generatedArgs, usedBlocks := calcSurface(fIndexes, fPolygonVectors, fTextureVectors, fTexture)

						prefix := fmt.Sprintf("Face L%d: f %s", fLn, fData)
						fmt.Printf("% -60s Step:%f Now:%s Parallel(running/total):%d/%d\n", prefix, step.Float(), time.Since(fObj_start), wgCurrentCount, wgTotalRoutine)

						mu.Lock()
						min[0] = math.Min(min[0], surfaceMin[0])
						min[1] = math.Min(min[1], surfaceMin[1])
						min[2] = math.Min(min[2], surfaceMin[2])
						max[0] = math.Max(max[0], surfaceMax[0])
						max[1] = math.Max(max[1], surfaceMax[1])
						max[2] = math.Max(max[2], surfaceMax[2])
						args = append(args, generatedArgs...)
						for id, count := range usedBlocks {
							totalUsedBlock[id] = totalUsedBlock[id] + count
						}
						face++
						mu.Unlock()
					}(ln, data, indexes, polygonVectors, textureVectors, material[currentTexture], blockColor, obj_start)
				}
			default:
				fmt.Printf("Skip L%d: %s\n", ln, line)
			}
		}
		wg.Wait()
		fmt.Printf("\nObject parse duration: %s\n", time.Since(obj_start))

		fmt.Printf("\nDuration: %s, Point:%d Face:%d\n", time.Since(start), len(polygonVectors), face)
		fmt.Printf("Min:[%.2f,%.2f,%.2f] Max:[%.2f,%.2f,%.2f] H:%.2f W:%.2f D:%.2f\n", min[0], min[1], min[2], max[0], max[1], max[2], max[0]-min[0], max[1]-min[1], max[2]-min[2])
		argumentList = append(argumentList, args)

	case Image:
		image_start := time.Now()
		fmt.Printf("\nImage parse start...\n")

		// command
		args := []CommandArgument{}

		var W, H float64

		f, err := os.Open(imageFilename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		for _, pixel := range parseImage(f) {
			blockId := getBlock(pixel.color)

			args = append(args, CommandArgument{
				color:   pixel.color,
				blockId: blockId,
				x:       pixel.x,
				y:       pixel.y,
				z:       0.0,
			})

			W = math.Max(W, pixel.x)
			H = math.Max(H, pixel.y)
			totalUsedBlock[blockId] = totalUsedBlock[blockId] + 1
		}
		fmt.Printf("\nImage parse duration: %s\n", time.Since(image_start))

		fmt.Printf("\nDuration: %s W:%d W:%d\n", time.Since(start), int(W), int(H))
		argumentList = append(argumentList, args)

	case Video:
		video_start := time.Now()
		fmt.Printf("\nVideo parse start...\n")

		// Get video duration
		var duration float64
		fmt.Printf("\nGet video duration start...\n")
		out, _ := exec.Command("ffmpeg", "-i", videoFilename).CombinedOutput()
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
		var frameData = map[int][]CommandArgument{}
		for current := 0.0; current < duration; current += 1.0 / float64(videoFrameRate) {
			frame++

			wgSession <- struct{}{}
			wg.Add(1)
			wgTotalRoutine++
			wgCurrentCount++

			go func(fCurrent, fDuration float64, fFrame int) {
				defer func() {
					<-wgSession
					wg.Done()
					wgCurrentCount--
				}()

				fmt.Printf("Start: frame: %d(%5.2f/%5.2f)\n", fFrame, fCurrent, fDuration)
				execute := exec.Command("ffmpeg", "-i", videoFilename, "-ss", fmt.Sprintf("%.3f", fCurrent), "-frames:v", "1", "-vf", "scale="+videoScaleSize, "-f", "image2pipe", "-vcodec", "png", "pipe:1")

				var buf bytes.Buffer
				execute.Stdout = &buf
				execute.Run()

				// command
				args := []CommandArgument{}
				usedBlocks := map[string]int{}

				for _, pixel := range parseImage(&buf) {
					blockId := getBlock(pixel.color)

					args = append(args, CommandArgument{
						color:   pixel.color,
						blockId: blockId,
						x:       pixel.x,
						y:       pixel.y,
						z:       0.0,
					})
					W = math.Max(W, pixel.x)
					H = math.Max(H, pixel.y)
					usedBlocks[blockId] = usedBlocks[blockId] + 1
				}

				mu.Lock()
				frameData[fFrame] = args
				for id, count := range usedBlocks {
					totalUsedBlock[id] = totalUsedBlock[id] + count
				}
				mu.Unlock()

				fmt.Printf("Finish: frame: %d(%5.2f/%5.2f) Parallel(running/total):%d/%d\n", fFrame, fCurrent, fDuration, wgCurrentCount, wgTotalRoutine)
			}(current, duration, frame)
		}
		wg.Wait()
		for i := 1; i < len(frameData); i++ {
			argumentList = append(argumentList, frameData[i])
		}

		fmt.Printf("\nFrame to function duration: %s\n", time.Since(video_start))

		fmt.Printf("\nDuration: %s, Frame: %d, W: %d, H: %d,\n", time.Since(start), len(argumentList), int(W), int(H))
	}

	fmt.Printf("\nCreate function...\n")
	create_start := time.Now()
	var totalFunctions, totalCommand int
	for i, args := range argumentList {
		functions, commandCount := CommandToMCfunction(args, fmt.Sprintf("f%04d-i", i+1))
		totalFunctions += len(functions)
		totalCommand += commandCount

		for _, f := range functions {
			fmt.Printf("%s.mcfunction\n", f)
		}
	}
	fmt.Printf("Total generated command/function: %d/%d\n", totalCommand, totalFunctions)
	fmt.Printf("\nCreate function duration: %s\n", time.Since(create_start))

	if enableBlockCount {
		fmt.Printf("\nBlock information:\n")
		type Count struct {
			BlockID string `json:"blockID"`
			Count   int    `json:"count"`
		}
		var blockCount []Count
		for blockID, count := range totalUsedBlock {
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

func removeDupeArgument(in []CommandArgument) []CommandArgument {
	tmp := make([]CommandArgument, len(in))
	copy(tmp, in)

	slices.SortStableFunc(tmp, func(a, b CommandArgument) int {
		// sort by position
		if cmp := floatCompare(a.x, b.x); cmp != 0 {
			return cmp
		}
		if cmp := floatCompare(a.y, b.y); cmp != 0 {
			return cmp
		}
		if cmp := floatCompare(a.z, b.z); cmp != 0 {
			return cmp
		}
		return 0
	})

	return slices.CompactFunc(tmp, func(a, b CommandArgument) bool {
		return floatEqual(a.x, b.x) && floatEqual(a.y, b.y) && floatEqual(a.z, b.z)
	})
}

func floatCompare(a, b float64) int {
	if floatEqual(a, b) {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

func floatEqual(a, b float64) bool {
	const threshold = 1e-6

	return math.Abs(a-b) < threshold
}
