package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Color struct {
	r, g, b int
}

func scanBlockModel() (blockModelList map[string]string) {
	blockModelList = map[string]string{}
	jsonModels := 0
	noStateModels := 0

	type blockstates struct {
		Variants map[string]interface{} `json:"variants"`
	}

	// Scan Model By Dir
	filepath.WalkDir(filepath.Join(minecraftRoot, "blockstates"), func(path string, d fs.DirEntry, err error) error {
		// check json file
		if filepath.Ext(path) != ".json" {
			return nil
		}
		jsonModels++

		b, _ := os.ReadFile(path)
		var states blockstates
		json.Unmarshal(b, &states)

		op, ok := states.Variants[""]
		if !ok {
			return nil
		}
		noStateModels++

		switch op.(type) {
		case []interface{}: // Random Rotate Texture(Array)
			modelPath := op.([]interface{})[0].(map[string]interface{})["model"].(string)
			modelPath = strings.ReplaceAll(modelPath, ":", "")
			modelPath = filepath.Base(modelPath)
			blockModelList[getBlockID(path)] = modelPath + ".json"
		case map[string]interface{}: // Not Random Rotate Texture
			modelPath := op.(map[string]interface{})["model"].(string)
			modelPath = strings.ReplaceAll(modelPath, ":", "")
			modelPath = filepath.Base(modelPath)
			blockModelList[getBlockID(path)] = modelPath + ".json"
		}
		return nil
	})

	fmt.Printf("Find blocks: %d\n", jsonModels)
	fmt.Printf("Stateless model: %d\n", noStateModels)
	return
}

func getBlockID(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func blockFilter(blockModelList map[string]string) (blockColor map[string]Color) {
	blockColor = map[string]Color{}

	modelDir := filepath.Join(minecraftRoot, "models", "block")
	textureDir := filepath.Join(minecraftRoot, "textures", "block")

	type Model struct {
		Parent   string            `json:"parent"`
		Textures map[string]string `json:"textures"`
	}

	for blockID, blockModel := range blockModelList {
		b, err := os.ReadFile(filepath.Join(modelDir, blockModel))
		if err != nil {
			continue
		}

		var model Model
		json.Unmarshal(b, &model)

		// model filter
		if model.Parent != "minecraft:block/cube_all" {
			continue
		}

		imageName, ok := model.Textures["all"]
		if ok {
			// name filter
			var isSkip = true
			for _, filterBlockID := range acceptBlockId {
				if regexp.MustCompile(filterBlockID).MatchString(blockID) {
					isSkip = false
					break
				}
			}

			for _, filterBlockID := range ignoreBlockId {
				if regexp.MustCompile(filterBlockID).MatchString(blockID) {
					isSkip = true
					break
				}
			}
			if isSkip {
				continue
			}

			// block to color
			blockImagePath := filepath.Join(textureDir, filepath.Base(imageName)+".png")

			f, err := os.Open(blockImagePath)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			img, _, _ := image.Decode(f)
			bounds := img.Bounds()
			var red, green, blue int
			var pixel int
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					r, g, b, _ := img.At(x, y).RGBA()
					red += int(r >> 8)
					green += int(g >> 8)
					blue += int(b >> 8)
					pixel++
				}
			}

			blockColor[blockID] = Color{
				r: red / pixel,
				g: green / pixel,
				b: blue / pixel,
			}
		}
	}

	fmt.Printf("All sides are same& name filtered block: %d\n", len(blockColor))
	return
}

func nearestColorBlock(target Color, blocks map[string]Color) (blockID string) {
	type Distance struct {
		blockID string
		d       float64
	}
	distance := []Distance{}
	for blockID, color := range blocks {
		Ha, Sa, La := RGBToHSL(uint8(color.r), uint8(color.g), uint8(color.b))
		Hc, Sb, Lb := RGBToHSL(uint8(target.r), uint8(target.g), uint8(target.b))

		// r := math.Pow(float64(target.r-color.r), 2)
		// g := math.Pow(float64(target.g-color.g), 2)
		// b := math.Pow(float64(target.b-color.b), 2)
		distance = append(distance, Distance{
			blockID: blockID,
			d:       HSLDistance(Ha, Sa, La, Hc, Sb, Lb),
		})
	}
	sort.Slice(distance, func(i, j int) bool {
		return distance[i].d < distance[j].d
	})
	return distance[0].blockID
}

func RGBToHSL(r, g, b uint8) (h, s, l float64) {
	// 0-255 の RGB 値を 0-1 の範囲に正規化
	fr := float64(r) / 255.0
	fg := float64(g) / 255.0
	fb := float64(b) / 255.0

	// 最大値・最小値を求める
	max := math.Max(math.Max(fr, fg), fb)
	min := math.Min(math.Min(fr, fg), fb)

	// 輝度 (Lightness)
	l = (max + min) / 2.0

	// 彩度 (Saturation)
	if max == min {
		s = 0 // グレースケール
		h = 0 // H は任意の値（通常 0）
	} else {
		delta := max - min
		if l > 0.5 {
			s = delta / (2.0 - max - min)
		} else {
			s = delta / (max + min)
		}

		// 色相 (Hue) の計算
		switch max {
		case fr:
			h = (fg - fb) / delta
			if fg < fb {
				h += 6
			}
		case fg:
			h = (fb-fr)/delta + 2
		case fb:
			h = (fr-fg)/delta + 4
		}

		h *= 60 // 角度に変換
	}

	return h, s, l
}

func HSLDistance(h1, s1, l1, h2, s2, l2 float64) float64 {
	// 色相 (H) の円環距離を考慮
	dh := math.Abs(h1 - h2)
	if dh > 180 {
		dh = 360 - dh
	}

	// ユークリッド距離計算
	ds := s1 - s2
	dl := l1 - l2
	return math.Sqrt(dh*dh*0.001 + ds*ds + dl*dl)
}

func CommandToMCfunction(commands []string, filePrefix string, maxChain int) (funcs []string, count int) {
	results := removeDupe(commands)
	count = len(results)

	funcs = []string{}
	for i := 1; i <= (len(results)/maxChain)+1; i++ {
		cmd := strings.Join(results[(i-1)*maxChain:Min(i*maxChain, len(results))], "\n")
		name := fmt.Sprintf("%s%04d", filePrefix, i)
		funcs = append(funcs, name)
		os.WriteFile(filepath.Join("./output", name+".mcfunction"), []byte(cmd), 0777)
	}

	return
}

func Min(x, y int) (m int) {
	if x < y {
		return x
	} else {
		return y
	}
}
