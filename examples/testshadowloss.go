package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/wiless/vlib"
)

var matlab *vlib.Matlab

func main() {
	rand.Seed(time.Now().Unix())
	variance := 10.0

	Area := 100.0 // 100x100 meter
	CorrD := 20.0 // every 20m a new shadow loss is generated
	rows := int(math.Ceil(Area / CorrD))
	// ShadowMatrix := vlib.NewMatrixF(rows, rows)
	ShadowMatrix := vlib.RandNMatrix(rows, rows, variance)
	// var pathloss pathloss.PathLossType
	sloss := vlib.RandNF(variance)

	fmt.Println("SLvar = ", sloss)
	fmt.Println("IIDShadow = ", ShadowMatrix)

	/// find corr per unit mtr
	point := vlib.Point{0, 0}
	cnt := 0

	correlation := vlib.NewVectorF(20) /// correlation per unit meter till 20m
	for i := 0.0; i < CorrD; i++ {
		delta := math.Abs(point.X - float64(i))
		correlation[cnt] = variance * math.Exp(-delta/CorrD)
		cnt++
	}
	fmt.Println("SLfromOrigin = ", correlation)

	fmt.Println("done")
}
