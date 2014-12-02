package main

import (
	"fmt"
	"math/rand"
	"time"
	"wiless/cellular/pathloss"
	"wiless/vlib"
)

var matlab *vlib.Matlab

func main() {

	rand.Seed(time.Now().Unix())

	var model pathloss.PathLossModel
	model.ModelSetting.SetDefault()
	model.ModelSetting.FreqHz = 2.0e9
	model.ModelSetting.CutOffDistance = 10
	MaxDistance := 500.0
	matlab = vlib.NewMatlab("pathloss.m")
	matlab.Silent = true
	indx := 0

	// LEN := int(math.Floor(MaxDistance / 10.0))
	var result vlib.VectorF //vlib.NewVectorF(LEN)
	for i := 0.0; i < MaxDistance; i += 5.0 {
		loss := model.LossInDb(i)
		// result[indx] = loss
		result.AppendAtEnd(loss)
		fmt.Printf("\n Distance %f : Loss %f dB", i, loss)
		indx++
	}
	fmt.Printf("Result %v", result)
	matlab.Export("Loss", result)
	matlab.Close()
	fmt.Println("\n")
}
