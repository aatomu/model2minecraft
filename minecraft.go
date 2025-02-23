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
	textureDir := filepath.Join(minecraftRoot, "textures", "block")

	type Model struct {
		Parent   string            `json:"parent"`
		Textures map[string]string `json:"textures"`
	}

	modelDir := filepath.Join(minecraftRoot, "models", "block")
	for blockID, blockModel := range blockModelList {
		b, err := os.ReadFile(filepath.Join(modelDir, blockModel))
		if err != nil {
			continue
		}

		var model Model
		json.Unmarshal(b, &model)

		// model filter
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
		r := math.Pow(float64(target.r-color.r), 2)
		g := math.Pow(float64(target.g-color.g), 2)
		b := math.Pow(float64(target.b-color.b), 2)
		distance = append(distance, Distance{
			blockID: blockID,
			d:       math.Sqrt(r + g + b),
		})
	}
	sort.Slice(distance, func(i, j int) bool {
		return distance[i].d < distance[j].d
	})
	return distance[0].blockID
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
