package main

import (
	"fmt"
	"github.com/wiless/vlib"

	// "encoding/json"
	// "fmt"
	"github.com/wiless/cellular/antenna"
)

func main() {

	var vcell1 antenna.SettingAAS

	vcell1.SetDefault()
	vcell1.N = 4
	vcell1.Freq = 1e9
	vcell1.BeamTilt = 0
	vcell1.DisableBeamTit = true
	vcell1.VTiltAngle = 0
	vcell1.ESpacingVFactor = .5
	vcell1.HTiltAngle = 0
	vcell1.Omni = false
	vcell1.MfileName = "output.m"
	vcell1.HoldOn = false
	vcell1.AASArrayType = antenna.LinearPhaseArray
	vcell1.CurveWidthInDegree = 30.0
	vcell1.CurveRadius = 1.00

	fmt.Print(vcell1)
	var AntennaLocation vlib.Location3D
	AntennaLocation.SetLoc(complex(0, 0), 25.0)
	vcell1.CreateElements(AntennaLocation)

	fmt.Printf("\nLocations %#v", vcell1.GetElements())

	fmt.Printf("\nLamda %#v", vcell1.GetLamda())

	///fmt.Printf("\nAntenna=%f", vlib.Location3DtoVecC(vcell1.GetElements()))

	// vcell1.ElementEffectiveGain(thetaH, thetaV)

	// antenna.RunAAS(setting)

}
