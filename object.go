package main

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
