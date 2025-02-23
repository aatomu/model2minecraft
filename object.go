package main

func getStep(Pa, Pb, Pc [3]Frac, spacing Frac) (step Frac) {
	var lambdaA, lambdaB, lambdaC Frac
	step = NewFrac(1, 1)

	// lambdaA=0(C=>B)
	for {
		lambdaA = NewFrac(0, 1)
		lambdaB = NewFrac(0, 1)
		lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
		Xa := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Ya := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Za := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		lambdaA = NewFrac(0, 1)
		lambdaB = step
		lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
		Xb := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Yb := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Zb := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		Xab := Xa.Sub(Xb).Pow(2)
		Yab := Ya.Sub(Yb).Pow(2)
		Zab := Za.Sub(Zb).Pow(2)
		V := Xab.Add(Yab).Add(Zab).Sqrt()

		if V.Sub(spacing).Float() > 0 {
			step = step.Div(NewFrac(2, 1))
			continue
		}
		break
	}
	// lambdaB=0(A=>C)
	for {
		lambdaB = NewFrac(0, 1)
		lambdaC = NewFrac(0, 1)
		lambdaA = NewFrac(1, 1).Sub(lambdaB).Sub(lambdaC)
		Xa := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Ya := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Za := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		lambdaB = NewFrac(0, 1)
		lambdaC = step
		lambdaA = NewFrac(1, 1).Sub(lambdaB).Sub(lambdaC)
		Xb := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Yb := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Zb := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		Xab := Xa.Sub(Xb).Pow(2)
		Yab := Ya.Sub(Yb).Pow(2)
		Zab := Za.Sub(Zb).Pow(2)
		V := Xab.Add(Yab).Add(Zab).Sqrt()
		if V.Sub(spacing).Float() > 0 {
			step = step.Div(NewFrac(2, 1))
			continue
		}
		break
	}
	// lambdaC=0(B=>A)
	for {
		lambdaC = NewFrac(0, 1)
		lambdaA = NewFrac(0, 1)
		lambdaB = NewFrac(1, 1).Sub(lambdaC).Sub(lambdaA)
		Xa := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Ya := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Za := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		lambdaC = NewFrac(0, 1)
		lambdaA = step
		lambdaB = NewFrac(1, 1).Sub(lambdaC).Sub(lambdaA)
		Xb := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
		Yb := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
		Zb := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

		Xab := Xa.Sub(Xb).Pow(2)
		Yab := Ya.Sub(Yb).Pow(2)
		Zab := Za.Sub(Zb).Pow(2)
		V := Xab.Add(Yab).Add(Zab).Sqrt()
		if V.Sub(spacing).Float() > 0 {
			step = step.Div(NewFrac(2, 1))
			continue
		}
		break
	}
	return
}

func getPolygonPoints(step Frac, Pa, Pb, Pc [3]Frac) (points [][3]Frac) {
	var lambdaA, lambdaB, lambdaC Frac

	lambdaA = NewFrac(0, 1)
	for lambdaA.Float() <= 1 {
		lambdaB = NewFrac(0, 1)
		for lambdaA.Add(lambdaB).Float() <= 1 {
			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
			x := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
			y := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
			z := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

			lambdaB = lambdaB.Add(step)

			points = append(points, [3]Frac{x, y, z})
		}
		lambdaA = lambdaA.Add(step)
	}

	return
}
func getTexturePoints(step Frac, Pa, Pb, Pc [2]Frac) (points [][2]Frac) {
	var lambdaA, lambdaB, lambdaC Frac

	lambdaA = NewFrac(0, 1)
	for lambdaA.Float() <= 1 {
		lambdaB = NewFrac(0, 1)
		for lambdaA.Add(lambdaB).Float() <= 1 {
			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
			x := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
			y := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))

			lambdaB = lambdaB.Add(step)

			points = append(points, [2]Frac{x, y})
		}
		lambdaA = lambdaA.Add(step)
	}

	return
}

// func getPoints(step Frac, Pa, Pb, Pc [3]Frac, spacing Frac) (cmd []string) {
// 	var lambdaA, lambdaB, lambdaC Frac
// 	generateCmds := []string{}

// 	lambdaA = NewFrac(0, 1)
// 	for lambdaA.Float() <= 1 {
// 		lambdaB = NewFrac(0, 1)
// 		for lambdaA.Add(lambdaB).Float() <= 1 {
// 			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
// 			x := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
// 			y := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
// 			z := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

// 			lambdaB = lambdaB.Add(step)

// 			generateCmds = append(generateCmds, fmt.Sprintf(command, math.Round(x.Div(spacing).Float())*spacing.Float(), math.Round(y.Div(spacing).Float())*spacing.Float(), math.Round(z.Div(spacing).Float())*spacing.Float()))
// 		}
// 		lambdaA = lambdaA.Add(step)
// 	}

// 	return generateCmds
// }
