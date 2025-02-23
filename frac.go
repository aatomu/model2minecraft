package main

import (
	"fmt"
	"math"
	"math/big"
)

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
