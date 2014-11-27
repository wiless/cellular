package simulation

import (
	"math"
	"math/cmplx"
)

type GenericStruct map[string]interface{}

func GetEJtheta(degree float64) complex128 {
	return cmplx.Exp(complex(0.0, -degree*math.Pi/180.0))
}

func Radian(degree float64) float64 {
	return degree * math.Pi / 180.0
}
