package main

import (
	"fmt"
	"github.com/wiless/vlib"
	"strconv"

	// "encoding/json"
	// "fmt"
	"github.com/wiless/cellular/antenna"
)

func main() {

	var Ncols int = 4
	var Freq float64 = 1.e9
	vcell1 := make([]antenna.SettingAAS, Ncols)

	var AntennaLocation vlib.Location3D
	Lamda := antenna.GetLamda(Freq)
	AntennaLength := float64(Ncols-1) * Lamda / 2.0
	fmt.Println(AntennaLength)
	AntennaLocation.SetLoc(complex(0, -(AntennaLength/2.0)), 25.0)
	for i := 0; i < Ncols; i++ {

		vcell1[i].SetDefault()
		vcell1[i].N = 4
		vcell1[i].Freq = 1e9
		vcell1[i].BeamTilt = 0
		vcell1[i].DisableBeamTit = true
		vcell1[i].VTiltAngle = 0
		vcell1[i].ESpacingVFactor = .5
		vcell1[i].HTiltAngle = 0
		vcell1[i].Omni = false
		vcell1[i].MfileName = "output" + strconv.Itoa(i) + ".m"
		vcell1[i].HoldOn = false
		vcell1[i].AASArrayType = antenna.LinearPhaseArray
		vcell1[i].CurveWidthInDegree = 30.0
		vcell1[i].CurveRadius = 1.00

		vcell1[i].CreateElements(AntennaLocation)
		AntennaLocation.Y += (Lamda / 2.0)
		fmt.Printf("\nLocations %f", vcell1[i].GetElements())
	}

	RxLocation := vlib.Location3D{10, 10, 5}

	fmt.Printf("\nRxcomponent from all Elements : %f", vcell1[0].GetRxPhase(RxLocation))
	vlib.Sum(v)
	gain, _, _ := vcell1[0].AASGain(RxLocation)
	fmt.Printf("\nRxcomponent from all Elements : %fdB", vlib.Db(gain))

}
