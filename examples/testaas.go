package main

import (
	"fmt"

	// "encoding/json"
	// "fmt"
	"github.com/wiless/cellular/antenna"
)

func main() {
	var setting antenna.SettingAAS

	setting.SetDefault()
	setting.N = 3
	setting.Freq = 2e9
	setting.BeamTilt = 0
	setting.DisableBeamTit = true
	setting.VTiltAngle = 0
	setting.HTiltAngle = 0
	setting.Omni = false
	setting.MfileName = "outputCircular.m"
	setting.HoldOn = false
	setting.AASArrayType = antenna.CircularPhaseArray
	setting.CurveWidthInDegree = 30.0
	setting.CurveRadius = 1.00

	fmt.Print(setting)

	// setting.CreateElements(vlib.Origin3D)
	// fmt.Println(setting.GetElements())
	// fmt.Printf("v=%f", vlib.Location3DtoVecC(setting.GetElements()))

	antenna.RunAAS(setting)

}
