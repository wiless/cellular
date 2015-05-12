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
	vcell1.FreqHz = 1e9
	vcell1.BeamTilt = 0
	vcell1.DisableBeamTit = true
	vcell1.VTiltAngle = 0
	vcell1.ESpacingVFactor = .5
	vcell1.HTiltAngle = -120
	vcell1.MfileName = "output.m"
	vcell1.Omni = false

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
	vpattern := vlib.NewVectorF(360)
	hpattern := vlib.NewVectorF(360)
	anglesRad := vlib.NewVectorF(360)

	vlib.IterateF(anglesRad)
	k := 0
	for i := 0.0; i < 360.0; i++ {
		hpattern[k] = vcell1.ElementDirectionHGain(i)
		vpattern[k] = vcell1.ElementDirectionHGain(i)
		k++
	}
	var matlab vlib.Matlab
	matlab.SetDefaults()
	matlab.SetFile("hpattern.m")
	matlab.Silent = true
	// matlab.Export("Weights", WeightVector)
	// matlab.Export("AntennaLocations", AntennaElementLocations)
	// matlab.Export("Locations", NodeLocations)
	matlab.Export("hpattern", hpattern)
	matlab.Export("vpattern", vpattern)
	matlab.Export("angleRad", anglesRad)
	// matlab.Export("N", N)
	// matlab.Export("Lamda", params.lamda)

	matlab.Command("\nangles=;")
	matlab.Command("polar(angle(hpattern),abs(hpattern),'k-')")
	matlab.Command("polar(angle(vpattern),abs(vpattern),'k-')")

	matlab.Close()

	// vcell1.ElementEffectiveGain(thetaH, 0)

	// antenna.RunAAS(setting)

}
