package main

import (
	"math"
)

func getBlock(target Color) (blockID string) {
	if colorDepthBit < 6 {
		return colorMap[target.r>>uint8(8-colorDepthBit)][target.g>>uint8(8-colorDepthBit)][target.b>>uint8(8-colorDepthBit)]
	}
	return nearestColorBlock(target)
}

func nearestColorBlock(target Color) (blockID string) {
	if id, ok := colorCache.Load(target); ok {
		return id.(string)
	}

	var distance float64 = math.MaxFloat64
	for _, block := range blockList {
		// RGB
		// d := RGBDistance(block.color, target)
		// // HLS
		// d := HSLDistance(blockColor[id], target)
		// // Lab
		d := LabDistance(block.color, target)

		if d < distance {
			blockID = block.id
			distance = d
		}
	}

	colorCache.Store(target, blockID)
	return
}

func RGBDistance(a, b Color) float64 {
	tmp := float64(a.r) - float64(b.r)
	red := tmp * tmp
	tmp = float64(a.g) - float64(b.g)
	green := tmp * tmp
	tmp = float64(a.b) - float64(b.b)
	blue := tmp * tmp
	return red + green + blue
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	// RGB(ff,ff,ff) => RGB(0..1,0..1,0..1)
	fr := float64(r) / 255.0
	fg := float64(g) / 255.0
	fb := float64(b) / 255.0

	max := math.Max(math.Max(fr, fg), fb)
	min := math.Min(math.Min(fr, fg), fb)

	// Lightness
	l = (max + min) / 2.0

	// Saturation
	if max == min {
		s = 0
		h = 0
	} else {
		delta := max - min
		if l > 0.5 {
			s = delta / (2.0 - max - min)
		} else {
			s = delta / (max + min)
		}

		// Calc Hue
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

		h *= 60 // To angle
	}

	return h, s, l
}

func HSLDistance(a, b Color) float64 {
	Ha, Sa, La := rgbToHSL(uint8(a.r), uint8(a.g), uint8(a.b))
	Hb, Sb, Lb := rgbToHSL(uint8(b.r), uint8(b.g), uint8(b.b))

	// H distance
	dh := math.Abs(Ha - Hb)
	if dh > 180 {
		dh = 360 - dh
	}

	// Calc euclidean distance
	ds := Sa - Sb
	dl := La - Lb
	return dh*dh + ds*ds + dl*dl
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
	// D65 White point (origin)
	refX, refY, refZ := 0.95047, 1.00000, 1.08883
	x /= refX
	y /= refY
	z /= refZ

	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Pow(t, 1.0/3.0)
		}
		return (7.787*t + 16.0/116.0)
	}

	x = f(x)
	y = f(y)
	z = f(z)

	L := (116 * y) - 16
	A := 500 * (x - y)
	B := 200 * (y - z)
	return L, A, B
}

func LabDistance(a, b Color) float64 {
	La, Aa, Ba := rgbToLab(a)
	Lb, Ab, Bb := rgbToLab(b)

	dL := (La - Lb) * (La - Lb)
	dA := (Aa - Ab) * (Aa - Ab)
	dB := (Ba - Bb) * (Ba - Bb)
	return dL + dA + dB
}
