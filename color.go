package main

import (
	"math"
	"slices"
)

func getBlock(target Color) (blockID string) {
	if colorBitDepth < 6 {
		return colorMap[target.r>>uint8(8-colorBitDepth)][target.g>>uint8(8-colorBitDepth)][target.b>>uint8(8-colorBitDepth)]
	}
	return nearestColorBlock(target)
}
func nearestColorBlock(target Color) (blockID string) {
	type Distance struct {
		blockID string
		d       float64
	}
	distance := []Distance{}
	for blockID, color := range blockColor {
		// // RGB
		// distance = append(distance, Distance{
		// 	blockID: blockID,
		// 	d:       RGBDistance(color, target),
		// })

		// // HLS
		// distance = append(distance, Distance{
		// 	blockID: blockID,
		// 	d:       HSLDistance(color, target),
		// })

		// Lab
		distance = append(distance, Distance{
			blockID: blockID,
			d:       LabDistance(color, target),
		})

	}

	slices.SortFunc(distance, func(a, b Distance) int {
		// Sort by distance
		if a.d < b.d {
			return -1
		}
		if a.d > b.d {
			return 1
		}
		// Sort by blockID (dictionary sort)
		if a.blockID < b.blockID {
			return -1
		}
		if a.blockID > b.blockID {
			return 1
		}

		return 0
	})

	return distance[0].blockID
}

func RGBDistance(a, b Color) float64 {
	red := math.Pow(float64(a.r-b.r), 2)
	green := math.Pow(float64(a.g-b.g), 2)
	blue := math.Pow(float64(a.b-b.b), 2)
	return math.Sqrt(red + green + blue)
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
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

func HSLDistance(a, b Color) float64 {
	Ha, Sa, La := rgbToHSL(uint8(a.r), uint8(a.g), uint8(a.b))
	Hb, Sb, Lb := rgbToHSL(uint8(b.r), uint8(b.g), uint8(b.b))

	// 色相 (H) の円環距離を考慮
	dh := math.Abs(Ha - Hb)
	if dh > 180 {
		dh = 360 - dh
	}

	// ユークリッド距離計算
	ds := Sa - Sb
	dl := La - Lb
	return math.Sqrt(dh*dh*0.001 + ds*ds + dl*dl)
}

func rgbToLab(rgb Color) (float64, float64, float64) {
	// RGB => XYZ
	red := float64(rgb.r) / 255.0
	green := float64(rgb.g) / 255.0
	blue := float64(rgb.b) / 255.0

	x := red*0.4124564 + green*0.3575761 + blue*0.1804375
	y := red*0.2126729 + green*0.7151522 + blue*0.0721750
	z := red*0.0193339 + green*0.1191920 + blue*0.9503041

	// XYZ => Lab
	// 規格化された D65 白色点 (基準値)
	refX, refY, refZ := 0.95047, 1.00000, 1.08883
	x /= refX
	y /= refY
	z /= refZ

	// Lab 変換のための関数
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Pow(t, 1.0/3.0)
		}
		return (7.787*t + 16.0/116.0)
	}

	x = f(x)
	y = f(y)
	z = f(z)

	// Lab 計算
	L := (116 * y) - 16
	A := 500 * (x - y)
	B := 200 * (y - z)
	return L, A, B
}

func LabDistance(a, b Color) float64 {
	La, Aa, Ba := rgbToLab(a)
	Lb, Ab, Bb := rgbToLab(b)

	dL := math.Pow(La-Lb, 2)
	dA := math.Pow(Aa-Ab, 2)
	dB := math.Pow(Ba-Bb, 2)
	return math.Sqrt(dL + dA + dB)
}
