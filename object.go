package main

import (
	"math"
	"strconv"
	"strings"
	"sync"
)

func getStep(Pa, Pb, Pc [3]float64, spacing float64) (step float64) {
	Distance := func(Xa, Ya, Za, Xb, Yb, Zb float64) float64 {
		Xab := (Xa - Xb) * (Xa - Xb)
		Yab := (Ya - Yb) * (Ya - Yb)
		Zab := (Za - Zb) * (Za - Zb)
		return math.Sqrt(Xab + Yab + Zab)
	}

	Vab := Distance(Pa[0], Pa[1], Pa[2], Pb[0], Pb[1], Pb[2])
	Vbc := Distance(Pb[0], Pb[1], Pb[2], Pc[0], Pc[1], Pc[2])
	Vca := Distance(Pc[0], Pc[1], Pc[2], Pa[0], Pa[1], Pa[2])
	Vmax := math.Max(math.Max(Vab, Vbc), Vca)
	step = spacing / Vmax
	return
}

func getPolygonPoints(step float64, Pa, Pb, Pc [3]float64) (points [][3]float64) {
	// allocate slice capacity
	points = make([][3]float64, 0, int(1/step)*int(1/step))

	var lambdaA, lambdaB, lambdaC float64

	for mulLambdaA := 0.0; mulLambdaA*step <= 1; mulLambdaA++ {
		for mulLambdaB := 0.0; (mulLambdaA+mulLambdaB)*step <= 1; mulLambdaB++ {
			lambdaA = mulLambdaA * step
			lambdaB = mulLambdaB * step
			lambdaC = 1.0 - lambdaA - lambdaB
			x, y, z := weightedPoint3D(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

			points = append(points, [3]float64{x, y, z})
		}
	}

	return
}

func getTexturePoints(step float64, Pa, Pb, Pc [2]float64) (points [][2]float64) {
	// allocate slice capacity
	points = make([][2]float64, 0, int(1/step)*int(1/step))

	var lambdaA, lambdaB, lambdaC float64

	for mulLambdaA := 0.0; mulLambdaA*step <= 1; mulLambdaA++ {
		for mulLambdaB := 0.0; (mulLambdaA+mulLambdaB)*step <= 1; mulLambdaB++ {
			lambdaA = mulLambdaA * step
			lambdaB = mulLambdaB * step
			lambdaC = 1.0 - lambdaA - lambdaB
			x := Pa[0]*lambdaA + Pb[0]*lambdaB + Pc[0]*lambdaC
			y := Pa[1]*lambdaA + Pb[1]*lambdaB + Pc[1]*lambdaC

			points = append(points, [2]float64{x, y})
		}
	}

	return
}

func weightedPoint3D(Pa, Pb, Pc [3]float64, lambdaA, lambdaB, lambdaC float64) (x, y, z float64) {
	x = Pa[0]*lambdaA + Pb[0]*lambdaB + Pc[0]*lambdaC
	y = Pa[1]*lambdaA + Pb[1]*lambdaB + Pc[1]*lambdaC
	z = Pa[2]*lambdaA + Pb[2]*lambdaB + Pc[2]*lambdaC
	return
}

func calcSurface(indexes []string, polygonVectors [][3]float64, textureVectors [][2]float64, texture [][]Color) (step float64, min [3]float64, max [3]float64, args []CommandArgument, usedBlock map[string]int) {
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
		min[i] = math.Min(min[i], polygonPa[i])
		min[i] = math.Min(min[i], polygonPb[i])
		min[i] = math.Min(min[i], polygonPc[i])
		// max
		max[i] = math.Max(max[i], polygonPa[i])
		max[i] = math.Max(max[i], polygonPb[i])
		max[i] = math.Max(max[i], polygonPc[i])
	}

	step = getStep(polygonPa, polygonPb, polygonPc, objectGridSpacing)
	wg := sync.WaitGroup{}
	wg.Add(2)
	var polygonPoints [][3]float64
	go func() {
		defer wg.Done()
		polygonPoints = getPolygonPoints(step, polygonPa, polygonPb, polygonPc)
	}()
	var texturePoints [][2]float64
	go func() {
		defer wg.Done()
		texturePoints = getTexturePoints(step, texturePa, texturePb, texturePc)
	}()
	wg.Wait()

	usedBlock = map[string]int{}
	const threshold = 1e-10

	args = make([]CommandArgument, 0, len(polygonPoints))
	for i := 0; i < len(polygonPoints); i++ {
		polygonPoint := polygonPoints[i]
		x := math.Round(polygonPoint[0]/objectGridSpacing) * objectGridSpacing
		if math.Abs(x) < threshold {
			x = 0
		}
		y := math.Round(polygonPoint[1]/objectGridSpacing) * objectGridSpacing
		if math.Abs(y) < threshold {
			y = 0
		}
		z := math.Round(polygonPoint[2]/objectGridSpacing) * objectGridSpacing
		if math.Abs(z) < threshold {
			z = 0
		}

		// Image position mapping
		//  Golang:   | Obj:
		//   0 - X+   |  Y+
		//   |        |  |
		//   Y+       |  0 - X+

		texturePoint := texturePoints[i]
		// -1..1 => -width..width
		textureX := math.Mod(texturePoint[0], 1)
		if textureX < 0 {
			textureX = 1 + textureX
		}
		textureIndexX := int(textureX * float64(len(texture)))
		// -1..1 => height..-height
		textureY := math.Mod(texturePoint[1], 1)
		if isObjectUVYAxisUp {
			textureY = 1 - textureY
		}
		if textureY < 0 {
			textureY = 1 + textureY
		}
		textureIndexY := int(textureY * float64(len(texture[textureIndexX])))
		texturePixel := texture[textureIndexX][textureIndexY]
		blockId := getBlock(texturePixel)

		args = append(args, CommandArgument{
			color:   texturePixel,
			blockId: blockId,
			position: Position{
				x: x,
				y: y,
				z: z,
			},
		})
		usedBlock[blockId] = usedBlock[blockId] + 1
	}

	return
}
