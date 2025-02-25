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

func NewFrac(top, bottom int64) Frac {
	f := Frac{
		Top:    big.NewInt(top),
		Bottom: big.NewInt(bottom),
	}
	f.approx()
	return f
}

func Float2Frac(x float64) Frac {
	return Frac{
		Top:    new(big.Int).SetInt64(int64(math.Round(x * 1e10))),
		Bottom: big.NewInt(1e10),
	}.approx()
}

// 約分
func (f Frac) approx() Frac {
	gcd := new(big.Int).GCD(nil, nil, f.Top, f.Bottom)
	if gcd.Cmp(big.NewInt(1)) != 0 { // gcd != 1
		f.Top.Div(f.Top, gcd)
		f.Bottom.Div(f.Bottom, gcd)
	}

	if f.Bottom.Sign() < 0 {
		f.Top.Neg(f.Top)
		f.Bottom.Neg(f.Bottom)
	}

	return f
}

func (f Frac) Add(n Frac) Frac {
	top := new(big.Int).Mul(f.Top, n.Bottom)
	top.Add(top, new(big.Int).Mul(n.Top, f.Bottom))

	bottom := new(big.Int).Mul(f.Bottom, n.Bottom)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) Sub(n Frac) Frac {
	top := new(big.Int).Mul(f.Top, n.Bottom)
	top.Sub(top, new(big.Int).Mul(n.Top, f.Bottom))

	bottom := new(big.Int).Mul(f.Bottom, n.Bottom)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) Mul(n Frac) Frac {
	top := new(big.Int).Mul(f.Top, n.Top)
	bottom := new(big.Int).Mul(f.Bottom, n.Bottom)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) Div(n Frac) Frac {
	top := new(big.Int).Mul(f.Top, n.Bottom)
	bottom := new(big.Int).Mul(f.Bottom, n.Top)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) Mod(n Frac) Frac {
	top := new(big.Int).Mul(f.Top, n.Bottom)
	top.Mod(top, new(big.Int).Mul(f.Bottom, n.Top))

	bottom := new(big.Int).Mul(f.Bottom, n.Bottom)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) Pow2() Frac {
	top := new(big.Int).Mul(f.Top, f.Top)
	bottom := new(big.Int).Mul(f.Bottom, f.Bottom)
	return Frac{Top: top, Bottom: bottom}.approx()
}

func (f Frac) String() string {
	return fmt.Sprintf("%d/%d", f.Top, f.Bottom)
}

func (f Frac) Float() float64 {
	top, _ := f.Top.Float64()
	bottom, _ := f.Bottom.Float64()
	return top / bottom
}
