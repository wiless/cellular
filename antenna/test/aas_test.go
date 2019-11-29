// Implements a simple phase delay from different n-Antenna elements
package antenna_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/wiless/cellular/antenna"
	"github.com/wiless/vlib"
)

/// Drops Linear Vertical Nodes spaced with dh,dv linearly
// func TestNewAAS(t *testing.T) {
// 	output := antenna.NewAAS()
// 	t.Log("Found antenna", output)
// }

func TestNewAAS(t *testing.T) {
	testant := antenna.NewAAS()

	ofile, ferr := os.Create("elementPattern.m")

	hgain := vlib.NewVectorF(360)
	vgain := vlib.NewVectorF(360)
	phaseangle := vlib.NewVectorF(360)
	cnt := 0
	// 	cmd := `delta=pi/180;
	// phaseangle=0:delta:2*pi-delta;`
	// matlab.Command(cmd)
	for d := 0; d < 360; d++ {
		hgain[cnt] = testant.ElementDirectionHGain(float64(d))
		vgain[cnt] = testant.ElementDirectionVGain(float64(d))
		phaseangle[cnt] = float64(d)
		cnt++
	}

	if ferr != nil {
		fmt.Printf("\nphase=%v;", phaseangle)
		fmt.Printf("\nel_hpattern=%v;", hgain)
		fmt.Printf("\nel_hpattern=%v;", vgain)

	} else {
		fmt.Fprintf(ofile, "\nphase=%v;", phaseangle)
		fmt.Fprintf(ofile, "\nel_hpattern=%v;", hgain)
		fmt.Fprintf(ofile, "\nel_vpattern=%v;", vgain)
		fmt.Fprint(ofile, "\n% phaseangle=degtorad(phase);")
		fmt.Fprint(ofile, "\n% polar(phaseangle, el_hpattern);")

		t.Log("See file : ", ofile.Name())
		ofile.Close()

		// t.Error("See file : ", ofile.Name())
	}

}

//AASGain2(thetaH, thetaV float64) (gaindB float64) {
func TestAASGain2(t *testing.T) {
	testant := antenna.NewAAS()
	testant.SetDefault()
	testant.N = 8
	testant.Omni = false
	testant.HTiltAngle = 0
	testant.VTiltAngle = 45
	testant.BeamTilt = 0
	testant.DisableBeamTit = false
	testant.GainDb = 3

	testant.FreqHz = 1.8e9
	testant.CreateLinearElements(vlib.Location3D{0, 0, 30})

	ofile, ferr := os.Create("elementPattern.m")
	defer ofile.Close()
	hgain := vlib.NewVectorF(360)
	vgain := vlib.NewVectorF(360)
	phaseangle := vlib.NewVectorF(360)
	cnt := 0
	// 	cmd := `delta=pi/180;
	// phaseangle=0:delta:2*pi-delta;`
	// matlab.Command(cmd)
	for d := 0; d < 360; d++ {
		hgain[cnt] = testant.ElementDirectionHGain(float64(d))
		vgain[cnt] = testant.ElementDirectionVGain(float64(d))
		phaseangle[cnt] = float64(d)
		cnt++
	}

	if ferr != nil {
		fmt.Printf("\nphase=%v;", phaseangle)
		fmt.Printf("\nel_hpattern=%v;", hgain)
		fmt.Printf("\nel_hpattern=%v;", vgain)

	} else {
		fmt.Fprintf(ofile, "\nphase=%v;", phaseangle)
		fmt.Fprintf(ofile, "\nel_hpattern=%v;", hgain)
		fmt.Fprintf(ofile, "\nel_vpattern=%v;", vgain)
		fmt.Fprint(ofile, "\n% phaseangle=degtorad(phase);")
		fmt.Fprint(ofile, "\n% polar(phaseangle, el_hpattern);")

		t.Log("See file : ", ofile.Name())
	}

	/// Do 3D beam
	fmt.Fprintf(ofile, "\nGain=[")
	for i := 0; i < 360; i++ {
		thetaV := float64(i)
		for j := 0; j < 360; j++ {
			thetaH := float64(j)
			gain3dDb := testant.AASGain2(thetaH, thetaV)
			fmt.Fprintf(ofile, "%3.2f ", gain3dDb)
		}
		fmt.Fprintf(ofile, ";\n")
	}
	fmt.Fprintf(ofile, "];\n")
}
