package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

var (
	scale   = NewFrac(9, 5)
	spacing = NewFrac(1, 1)
	// command         = "particle dust{color:[0f,0f,0f],scale:1f} ~%.2f ~%.2f ~%.2f 0 0 0 0 1 force @a"
	face_command       = "setblock ~%.2f ~%.2f ~%.2f minecraft:gray_wool"
	vector_command     = "setblock ~%.2f ~%.2f ~%.2f minecraft:light_gray_wool"
	top_command        = "setblock ~%.2f ~%.2f ~%.2f minecraft:white_wool"
	chain          int = 700000
	rootDir            = "./3d"
	fileName           = "HatsuneMiku.obj"
)

func main() {
	parse_start := time.Now()

	obj, _ := os.ReadFile(filepath.Join(rootDir, fileName))
	commands := []string{}

	polygonVectors := [][3]Frac{}
	points := big.NewInt(0)
	face := big.NewInt(0)
	min := [3]float64{}
	max := [3]float64{}

	for ln, line := range strings.Split(string(obj), "\n") {
		cmd := strings.SplitN(line, " ", 2)
		if len(cmd) < 2 {
			continue
		}

		data := cmd[1]

		switch cmd[0] {
		case "v": //点座標
			{
				var x, y, z float64
				fmt.Sscanf(data, "%f %f %f", &x, &y, &z)
				polygonVectors = append(polygonVectors, [3]Frac{Float2Frac(x).Mul(scale), Float2Frac(y).Mul(scale), Float2Frac(z).Mul(scale)})
				fmt.Printf("PolygonVector L%d: %s\n", ln, line)
				points = new(big.Int).Add(points, big.NewInt(1))
			}
		case "f": //面情報
			{
				indexes := strings.Split(data, " ")
				if len(indexes) < 3 {
					fmt.Printf("Skip L%d: %s\n", ln, line)
					continue
				}

				polygonPa, _ := strconv.Atoi(strings.Split(indexes[0], "/")[0])
				polygonPb, _ := strconv.Atoi(strings.Split(indexes[1], "/")[0])
				polygonPc, _ := strconv.Atoi(strings.Split(indexes[2], "/")[0])
				Pa := polygonVectors[polygonPa-1]
				Pb := polygonVectors[polygonPb-1]
				Pc := polygonVectors[polygonPc-1]

				var lambdaA, lambdaB, lambdaC Frac
				var step Frac = NewFrac(1, 1)
				var exponential = 1

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
						exponential++
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
						exponential++
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
						exponential++
						continue
					}
					break
				}

				for i := 0; i < 3; i++ {
					if min[i] > Pa[i].Float() {
						min[i] = Pa[i].Float()
					}
					if min[i] > Pb[i].Float() {
						min[i] = Pb[i].Float()
					}
					if min[i] > Pc[i].Float() {
						min[i] = Pc[i].Float()
					}

					if max[i] < Pa[i].Float() {
						max[i] = Pa[i].Float()
					}
					if max[i] < Pb[i].Float() {
						max[i] = Pb[i].Float()
					}
					if max[i] < Pc[i].Float() {
						max[i] = Pc[i].Float()
					}
				}
				prefix := fmt.Sprintf("Face L%d: %s", ln, line)
				var generateCmds []string = generate(step, Pa, Pb, Pc, spacing)

				fmt.Printf("% -60s Step:%f Now:%s\n    ABC:%s,%s,%s\n", prefix, step.Float(), time.Since(parse_start), Pa, Pb, Pc)
				face = new(big.Int).Add(face, big.NewInt(1))
				commands = append(commands, removeDupe(generateCmds)...)
			}
		default:
			fmt.Printf("Skip L%d: %s\n", ln, line)
		}
	}
	fmt.Println("Parse/Calc duration:", time.Since(parse_start))
	fmt.Println("Create function...")
	create_start := time.Now()
	_, n := CommandToMCfunction(commands, "", chain)
	fmt.Println("create function duration:", time.Since(create_start))
	fmt.Printf("\n\nDuration: %s, Point:%s Face:%s Cmd:%d\n", time.Since(parse_start), points.String(), face.String(), n)
	fmt.Printf("Min:[%.2f,%.2f,%.2f] Max:[%.2f,%.2f,%.2f] H:%.2f W:%.2f D:%.2f\n", min[0], min[1], min[2], max[0], max[1], max[2], max[0]-min[0], max[1]-min[1], max[2]-min[2])
}

func CommandToMCfunction(commands []string, filePrefix string, maxChain int) (funcs []string, commandCnt int) {
	results := removeDupe(commands)
	commandCnt = len(results)
	slices.SortFunc(results, func(a string, b string) int {
		if strings.Contains(a, "white") {
			return 1
		}
		if strings.Contains(b, "white") {
			return -1
		}
		if strings.Contains(a, "light_gray") {
			return 1
		}
		if strings.Contains(b, "light_gray") {
			return -1
		}
		return 0
	})

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
	}
	return y
}

func Abs(x float64) float64 {
	if x < 0 {
		return x * -1
	}
	return x
}

func removeDupe(in []string) []string {
	results := []string{}
	encountered := map[string]struct{}{}
	for _, v := range in {
		if _, ok := encountered[v]; !ok {
			encountered[v] = struct{}{}
			results = append(results, v)
		}
	}
	return results
}

// func calcDistance(ax, ay, az, bx, by, bz float64) float64 {
// 	return math.Sqrt(math.Pow(ax-bx, 2) + math.Pow(ay-by, 2) + math.Pow(az-bz, 2))
// }

func generate(step Frac, Pa, Pb, Pc [3]Frac, spacing Frac) (cmd []string) {
	var lambdaA, lambdaB, lambdaC Frac
	generateCmds := []string{}

	lambdaA = NewFrac(0, 1)
	for lambdaA.Float() <= 1 {
		lambdaB = NewFrac(0, 1)
		for lambdaA.Add(lambdaB).Float() <= 1 {
			lambdaC = NewFrac(1, 1).Sub(lambdaA).Sub(lambdaB)
			x := Pa[0].Mul(lambdaA).Add(Pb[0].Mul(lambdaB)).Add(Pc[0].Mul(lambdaC))
			y := Pa[1].Mul(lambdaA).Add(Pb[1].Mul(lambdaB)).Add(Pc[1].Mul(lambdaC))
			z := Pa[2].Mul(lambdaA).Add(Pb[2].Mul(lambdaB)).Add(Pc[2].Mul(lambdaC))

			lambdaB = lambdaB.Add(step)

			if lambdaA.Float() >= 1-1e-3 || lambdaB.Float() >= 1-1e-3 || lambdaC.Float() >= 1-1e-3 {
				generateCmds = append(generateCmds, fmt.Sprintf(top_command, math.Round(x.Div(spacing).Float())*spacing.Float(), math.Round(y.Div(spacing).Float())*spacing.Float(), math.Round(z.Div(spacing).Float())*spacing.Float()))
				continue
			}
			if lambdaA.Float() == 0.1 || lambdaB.Float() == 0.1 || lambdaC.Float() == 0.1 {
				generateCmds = append(generateCmds, fmt.Sprintf(vector_command, math.Round(x.Div(spacing).Float())*spacing.Float(), math.Round(y.Div(spacing).Float())*spacing.Float(), math.Round(z.Div(spacing).Float())*spacing.Float()))
				continue
			}
			generateCmds = append(generateCmds, fmt.Sprintf(face_command, math.Round(x.Div(spacing).Float())*spacing.Float(), math.Round(y.Div(spacing).Float())*spacing.Float(), math.Round(z.Div(spacing).Float())*spacing.Float()))
		}
		lambdaA = lambdaA.Add(step)
	}

	return generateCmds
}

type Frac struct {
	Top    *big.Int
	Bottom *big.Int
}

func NewFrac(top, bottom int64) (f Frac) {
	f = Frac{
		Top:    big.NewInt(top),
		Bottom: big.NewInt(bottom),
	}

	f.approx()
	return
}

func Float2Frac(x float64) (f Frac) {
	bottom := 1e10
	top := int64(math.Round(x * bottom))

	f = Frac{
		Top:    big.NewInt(top),
		Bottom: big.NewInt(int64(bottom)),
	}

	f.approx()
	return
}

// 約分
func (f *Frac) approx() {
	gcd := new(big.Int).GCD(nil, nil, f.Top, f.Bottom)
	f.Top = new(big.Int).Div(f.Top, gcd)
	f.Bottom = new(big.Int).Div(f.Bottom, gcd)

	if f.Bottom.Cmp(big.NewInt(0)) == -1 {
		f.Top = new(big.Int).Mul(f.Top, big.NewInt(-1))
		f.Bottom = new(big.Int).Mul(f.Bottom, big.NewInt(-1))
	}
}

func (f Frac) Add(n Frac) (ans Frac) {
	ans = Frac{
		Top:    new(big.Int).Add(new(big.Int).Mul(f.Top, n.Bottom), new(big.Int).Mul(n.Top, f.Bottom)),
		Bottom: new(big.Int).Mul(f.Bottom, n.Bottom),
	}

	ans.approx()
	return
}

func (f Frac) Sub(n Frac) (ans Frac) {
	ans = Frac{
		Top:    new(big.Int).Sub(new(big.Int).Mul(f.Top, n.Bottom), new(big.Int).Mul(n.Top, f.Bottom)),
		Bottom: new(big.Int).Mul(f.Bottom, n.Bottom),
	}

	ans.approx()
	return
}

func (f Frac) Mul(n Frac) (ans Frac) {

	ans = Frac{
		Top:    new(big.Int).Mul(f.Top, n.Top),
		Bottom: new(big.Int).Mul(f.Bottom, n.Bottom),
	}

	ans.approx()
	return
}

func (f Frac) Div(n Frac) (ans Frac) {
	ans = Frac{
		Top:    new(big.Int).Mul(f.Top, n.Bottom),
		Bottom: new(big.Int).Mul(f.Bottom, n.Top),
	}

	ans.approx()
	return
}

func (f Frac) Pow(n int64) (ans Frac) {
	if n < 0 {
		return Frac{Top: f.Bottom, Bottom: f.Top}.Pow(-n)
	}
	ans = Frac{
		Top:    new(big.Int).Exp(f.Top, big.NewInt(n), nil),
		Bottom: new(big.Int).Exp(f.Bottom, big.NewInt(n), nil),
	}

	ans.approx()
	return
}

func (f Frac) Sqrt() (ans Frac) {
	top := new(big.Int).Sqrt(f.Top)
	bottom := new(big.Int).Sqrt(f.Bottom)

	// 完全平方
	if new(big.Int).Mul(top, top).Cmp(f.Top) == 0 && new(big.Int).Mul(bottom, bottom).Cmp(f.Bottom) == 0 {
		ans = Frac{Top: top, Bottom: bottom}
	} else {
		top = new(big.Int).Sqrt(new(big.Int).Mul(f.Top, new(big.Int).Exp(big.NewInt(10), big.NewInt(100), nil)))
		bottom = new(big.Int).Sqrt(new(big.Int).Mul(f.Bottom, new(big.Int).Exp(big.NewInt(10), big.NewInt(100), nil)))
		ans = Frac{Top: top, Bottom: bottom}
	}

	ans.approx()
	return
}

func (f Frac) String() string {
	return fmt.Sprintf("%d/%d", f.Top, f.Bottom)
}

func (f Frac) Float() float64 {
	top, _ := f.Top.Float64()
	bottom, _ := f.Bottom.Float64()
	return top / bottom
}
