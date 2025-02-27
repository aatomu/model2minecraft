package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type Command func(arg CommandArgument) (cmd string)

type CommandArgument struct {
	color   Color
	blockId string
	x, y, z float64
}

type BlockModel struct {
	namespace string
	path      string
}

// 0..255 RGB color
type Color struct {
	r, g, b uint8
}

func scanBlockModel() (blockModelList map[string]BlockModel) {
	blockModelList = map[string]BlockModel{}
	jsonModels := 0
	noStateModels := 0

	type blockstates struct {
		Variants map[string]interface{} `json:"variants"`
	}

	// Scan Model By Dir
	filepath.WalkDir(filepath.Join(minecraftDirectory, "minecraft", "blockstates"), func(path string, d fs.DirEntry, err error) error {
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

		switch modelJson := op.(type) {
		case []interface{}: // Random Rotate Texture(Array)
			modelPath := modelJson[0].(map[string]interface{})["model"].(string)
			blockModelList[removeExt(path)] = parsePath(modelPath)

		case map[string]interface{}: // Not Random Rotate Texture
			modelPath := modelJson["model"].(string)
			blockModelList[removeExt(path)] = parsePath(modelPath)
		}
		return nil
	})

	fmt.Printf("Find blocks: %d\n", jsonModels)
	fmt.Printf("Stateless model: %d\n", noStateModels)
	return
}

func blockFilter(blockModelList map[string]BlockModel) (blockColor map[string]Color) {
	blockColor = map[string]Color{}

	type Model struct {
		Parent   string            `json:"parent"`
		Textures map[string]string `json:"textures"`
	}

	for blockID, blockModel := range blockModelList {
		b, err := os.ReadFile(filepath.Join(minecraftDirectory, blockModel.namespace, "models", blockModel.path+".json"))
		if err != nil {
			continue
		}

		var model Model
		json.Unmarshal(b, &model)

		imagePath, ok := model.Textures["all"]
		if !ok {
			continue
		}

		// name filter
		var isSkip = true
		for _, filterBlockID := range allowedBlockIds {
			if regexp.MustCompile(filterBlockID).MatchString(blockID) {
				isSkip = false
				break
			}
		}

		for _, filterBlockID := range ignoredBlockIds {
			if regexp.MustCompile(filterBlockID).MatchString(blockID) {
				isSkip = true
				break
			}
		}
		if isSkip {
			continue
		}

		// block to color
		texture := parsePath(imagePath)
		blockImagePath := filepath.Join(minecraftDirectory, texture.namespace, "textures", texture.path+".png")

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
			r: uint8(red / pixel),
			g: uint8(green / pixel),
			b: uint8(blue / pixel),
		}
	}

	blockList = slices.Sorted(maps.Keys(blockColor))

	fmt.Printf("All sides are same& name filtered block: %d\n", len(blockColor))
	return
}

func CommandToMCfunction(args []CommandArgument, filePrefix string) (funcs []string, count int) {
	result := removeDupeArgument(args)
	count = len(result)

	funcs = []string{}
	for i := 0; i <= (len(result)/maxCommandChain-1)/maxCommandChain; i++ {
		var builder strings.Builder
		start := i * maxCommandChain
		end := Min((i+1)*maxCommandChain, len(result))
		for _, arg := range result[start:end] {
			builder.WriteString(commandGenerator(arg))
			builder.WriteString("\n")
		}
		name := fmt.Sprintf("%s%04d", filePrefix, i+1)
		funcs = append(funcs, name)
		err := os.WriteFile(filepath.Join("./output", name+".mcfunction"), []byte(builder.String()), 0777)
		if err != nil {
			panic(err)
		}
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

func removeExt(s string) string {
	return filepath.Base(s[:len(s)-len(filepath.Ext(s))])
}

func parsePath(p string) BlockModel {
	modelPath := strings.SplitN(p, ":", 2)
	if len(modelPath) == 1 {
		return BlockModel{
			namespace: "minecraft",
			path:      modelPath[0],
		}
	}
	return BlockModel{
		namespace: modelPath[0],
		path:      modelPath[1],
	}
}
