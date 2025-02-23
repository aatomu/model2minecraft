package main

import (
	"math"
	"strconv"
	"strings"
	"sync"
)

func getStep(Pa, Pb, Pc [3]Frac, spacing Frac) (step Frac) {
	var lambdaA, lambdaB, lambdaC Frac
	step = NewFrac(1, 1)

	// lambdaA=0(C=>B)
	for {
		lambdaA = NewFrac(0, 1)
		lambdaB = NewFrac(0, 1)
		lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
		Xa, Ya, Za := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		lambdaB = step
		lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
		Xb, Yb, Zb := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		V := calcDistance(Xa, Ya, Za, Xb, Yb, Zb)
		if V.Sub(spacing).Float() > 0 {
			step = step.Mul(NewFrac(2, 3))
			continue
		}
		break
	}

	// lambdaB=0(A=>C)
	for {
		lambdaB = NewFrac(0, 1)
		lambdaC = NewFrac(0, 1)
		lambdaA = NewFrac(1, 1).Sub(lambdaB).Sub(lambdaC)
		Xa, Ya, Za := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		lambdaC = step
		lambdaA = NewFrac(1, 1).Sub(lambdaB).Sub(lambdaC)
		Xb, Yb, Zb := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		V := calcDistance(Xa, Ya, Za, Xb, Yb, Zb)
		if V.Sub(spacing).Float() > 0 {
			step = step.Mul(NewFrac(2, 3))
			continue
		}
		break
	}

	// lambdaC=0(B=>A)
	for {
		lambdaC = NewFrac(0, 1)
		lambdaA = NewFrac(0, 1)
		lambdaB = NewFrac(1, 1).Sub(lambdaC).Sub(lambdaA)
		Xa, Ya, Za := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		lambdaA = step
		lambdaB = NewFrac(1, 1).Sub(lambdaC).Sub(lambdaA)
		Xb, Yb, Zb := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

		V := calcDistance(Xa, Ya, Za, Xb, Yb, Zb)
		if V.Sub(spacing).Float() > 0 {
			step = step.Mul(NewFrac(2, 3))
			continue
		}
		break
	}

	return
}

func getPolygonPoints(step Frac, Pa, Pb, Pc [3]Frac) (points [][3]Frac) {
	// allocate slice capacity
	lambdaRange := int(1 / step.Float())
	points = make([][3]Frac, 0, lambdaRange*lambdaRange)

	var lambdaA, lambdaB, lambdaC Frac

	lambdaA = NewFrac(0, 1)
	for lambdaA.Float() <= 1 {
		lambdaB = NewFrac(0, 1)
		for lambdaA.Add(lambdaB).Float() <= 1 {
			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
			x, y, z := getPoint(Pa, Pb, Pc, lambdaA, lambdaB, lambdaC)

			points = append(points, [3]Frac{x, y, z})

			lambdaB = lambdaB.Add(step)
		}
		lambdaA = lambdaA.Add(step)
	}

	return
}

func getTexturePoints(step Frac, Pa, Pb, Pc [2]Frac) (points [][2]Frac) {
	// allocate slice capacity
	lambdaRange := int(1 / step.Float())
	points = make([][2]Frac, 0, lambdaRange*lambdaRange)

	var lambdaA, lambdaB, lambdaC Frac

	lambdaA = NewFrac(0, 1)
	for lambdaA.Float() <= 1 {
		lambdaB = NewFrac(0, 1)
		for lambdaA.Add(lambdaB).Float() <= 1 {
			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
			x := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
			y := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))

			points = append(points, [2]Frac{x, y})

			lambdaB = lambdaB.Add(step)
		}
		lambdaA = lambdaA.Add(step)
	}

	return
}

func getPoint(Pa, Pb, Pc [3]Frac, lambdaA, lambdaB, lambdaC Frac) (x, y, z Frac) {
	x = Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
	y = Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
	z = Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))
	return
}

func calcDistance(Xa, Ya, Za, Xb, Yb, Zb Frac) Frac {
	Xab := Xa.Sub(Xb).Pow(2)
	Yab := Ya.Sub(Yb).Pow(2)
	Zab := Za.Sub(Zb).Pow(2)
	return Xab.Add(Yab).Add(Zab).Sqrt()
}

func calcSurface(indexes []string, polygonVectors [][3]Frac, textureVectors [][2]Frac, texture [][]Color, blockColor map[string]Color) (step Frac, min [3]float64, max [3]float64, commands []string, usedBlock map[string]int) {
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

	step = getStep(polygonPa, polygonPb, polygonPc, objectSpacing)
	wg := sync.WaitGroup{}
	wg.Add(2)
	var polygonPoints [][3]Frac
	go func() {
		defer wg.Done()
		polygonPoints = getPolygonPoints(step, polygonPa, polygonPb, polygonPc)
	}()
	var texturePoints [][2]Frac
	go func() {
		defer wg.Done()
		texturePoints = getTexturePoints(step, texturePa, texturePb, texturePc)
	}()
	wg.Wait()

	usedBlock = map[string]int{}
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
		textureIndexX := int(textureX * float64(len(texture)))
		// -1..1 => height..-height
		textureY := 1 - texturePoint[1].Mod(NewFrac(1, 1)).Float()
		textureIndexY := int(textureY * float64(len(texture[textureIndexX])))
		textureColor := texture[textureIndexX][textureIndexY]
		blockId := nearestColorBlock(textureColor, blockColor)

		commands = append(commands, generator(textureColor, x, y, z, blockId))
		usedBlock[blockId] = usedBlock[blockId] + 1
	}

	commands = removeDupe(commands)

	return
}
