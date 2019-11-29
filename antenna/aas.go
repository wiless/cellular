// Implements a simple phase delay from different n-Antenna elements
package antenna

import (
	"github.com/jvlmdr/lin-go/zvec"
	"encoding/json"
	"log"
	"math/cmplx"
	"os"

	"gonum.org/v1/gonum/mat"

	"gonum.org/v1/gonum/blas/cblas128"

	// "flag"
	"fmt"
	"math"

	"github.com/wiless/vlib"
)

type cvec cblas128.Vector

type Dim struct {
	R, C int
}

type Beam struct {
	HEtilt float64 // in radians
	VEtilt float64 // in radians
	Index  int
	W      mat.CDense
	V      mat.CDense
}

// Element represents an Antenna element, with horiontal and vertical normalized to 0,0. (horizon=0, vertical=0) and unit gain
type Element struct {
	VBeamWidth, HBeamWidth float64 // H & V beamwidth of antenna element
	HDirection             float64
	VDirection             float64
	GainDb                 float64 // ElementGain in dBi
	SLAV                   float64
	Omni                   bool
}

// AAS is the Active Antenna system supporting multiple TxRU and weights based on 3GPP & M.2412
type AAS struct {
	elementLocations   []vlib.Location3D
	lamda              float64         // wavelength
	FreqGHz            float64         // Frequency of operation in GHz
	Size               Dim             // MxN Number of  M=vertical/rows & N=horizontal/columns elements
	NumTxRU            Dim             // MxN Number of  M=vertical/rows & N=horizontal/columns elements
	Polarization       int             // True if P=2, that dual Polarized
	vTiltAngle         float64         // Orientation : Mechanical-Vertical Tilt w.r.t Horizon
	hTiltAngle         float64         //  Orientation :  Mechanical-Horizontal Direction (usually same as sector direction, exceptional case indoor, rooftop etc)
	HSpacing, VSpacing float64         // Horizontal & Vertical Spacing of antenna elements as fraction of Lamda
	Centre             vlib.Location3D // Center position of the AAS Panel
	w                  vlib.VectorC
	v                  vlib.VectorC
	SizeTxRU           Dim    // Nof elements RxC in a TxRU, when all TXRU are of same dimentions
	AASArrayType       string // Only supports Type="URA" Uniform Rectangular Array
	Beams              []Beam
	elem               Element
}

// NewAASsimple initializes an AAS with MxN Rectangular Antenna Array, with HTx
// with f=4GHz , single element
func NewAASsimple(gdB float64) *AAS {
	a := new(AAS)
	a.Init(4.0, 1, 1, 1, 1, 1, gdB) // Assuming default 4GHz

	return a
}

// Init initializes an AAS with MxN Rectangular Antenna Array, with HTx, element Gain
func (a *AAS) Init(fGHz float64, M, N, P, HTxRU, VTxRU int, gdB float64) {
	a.FreqGHz = fGHz
	a.Size = Dim{M, N}
	a.NumTxRU = Dim{HTxRU, VTxRU}
	a.lamda = SPEEDOFLIGHT / freq
	a.HSpacing = 0 /// Factor mulplied by parama.lamda
	a.VSpacing = .5
	a.elem.GainDb = gdB // 0dBi Element
}

// Set sets
func (a *AAS) Set(str string) {
	err := json.Unmarshal([]byte(str), a)
	if err != nil {
		log.Print("Error ", err)
	}
}

// Get gets
func (a *AAS) Get() string {

	bytes, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(bytes)
	}
}

// Check checks with the AAS is initialized
func (a AAS) Check() bool {
	if a.lamda == 0 {
		return false
	} else {

		if len(a.w) == 0 || len(a.v) == 0 {
			return false
		} else {
			return true
		}
	}
}

// BeamGain returns the gain of all the TxRU in the AAS w.r.t dest location
// NOTE : Caution using this function, the antenna orientation and position should be set same as Sector Direction, Sector-Node location
func (a *AAS) BeamGain(dest vlib.Location3D) (Port0Gain float64, effectiveThetaH, effectiveThetaV float64, TxRuGains [][]float64, err error) {
	if !a.Check() {
		err := fmt.Errorf("BeamGain:AAS not initialized")
		return Port0Gain, effectiveThetaH, effectiveThetaV, TxRuGains, err
	}

	// AntennaElementLocations := vlib.Location3DtoVecC(params.elementLocations)

	// // w =

	// w:=a.w //params.FindWeights(params.BeamTilt)
	// if a.DisableBeamTit {
	// 	w = vlib.NewOnesC(AntennaElementLocations.Size())
	// }
	// w = w.Scale(math.Sqrt(1.0 / float64(params.N)))

	// // fmt.Print(AntennaElementLocations)
	// phaseDelay := vlib.NewVectorF(AntennaElementLocations.Size())
	// var Rxcomponent complex128
	// Rxcomponent = 0.0
	// var dist, thetaH, thetaV float64
	// var aGain complex128

	// for i := 0; i < a.N; i++ {
	// 	// dist, thetaH, thetaV := vlib.RelativeGeo(params.elementLocations[i], dest)
	// 	dist, thetaH, thetaV = vlib.RelativeGeo(params.Centre, dest)

	// 	// dist= cmplx.Abs(params.elementLocations[i].Cmplx()-dest.Cmplx())
	// 	aGain = complex((params.ElementEffectiveGain(thetaH, thetaV)), 0)

	// 	_, phaseDelay[i] = math.Modf(2 * math.Pi * dist / params.lamda)

	// 	Rxcomponent += GetEJtheta(vlib.ToDegree(phaseDelay[i])) * w[i] * aGain
	// }

	// gain = math.Pow(cmplx.Abs(Rxcomponent), 1)
	// gain = cmplx.Abs(Rxcomponent) // validate @ssk - May 28th 2017
	// dist, thetaH, thetaV = vlib.RelativeGeo(params.Centre, dest)

	// if gain > vlib.InvDb(params.GainDb) {
	// 	log.Printf("\n AAS  : Rx complex = %v Rx |aGain  %v | %v  | %v limit at dist=%v", Rxcomponent, gain, vlib.InvDb(params.GainDb), dist)
	// }
	// return gain, thetaH, thetaV
	return 0, 0, 0, nil, nil
}

// func (a *AAS) oldFindWeights(theta float64) vlib.VectorC {
// 	WeightVectors := vlib.NewVectorC(params.N)
// 	// var gain complex128
// 	AE := vlib.Location3DtoVecC(params.elementLocations)
// 	meanpos := vlib.MeanC(AE)
// 	pos := GetEJtheta(theta) + meanpos
// 	// gain := complex(1.0/math.Sqrt(float64(N)), 0)
// 	phaseDelay := vlib.NewVectorF(AE.Size())
// 	for eindx, epos := range AE {
// 		dist := cmplx.Abs(epos - pos)
// 		_, phaseDelay[eindx] = math.Modf(dist / params.lamda)
// 		phaseDelay[eindx] *= (2.0 * math.Pi)
// 		WeightVectors[eindx] = cmplx.Exp(complex(0.0, -phaseDelay[eindx]))

// 	}
// 	return WeightVectors
// }

// func (a *AAS) FindWeights(theta float64) vlib.VectorC {
// 	// for nindx, pos := range NodeLocations {

// 	WeightVectors := vlib.NewVectorC(params.N)

// 	for i := 0; i < params.N; i++ {
// 		m := float64(i)
// 		arg := complex(0, 2*math.Pi*(m-1)*params.lamda/2.0*math.Cos(Radian(theta+90))/params.lamda)
// 		WeightVectors[i] = cmplx.Exp(arg)
// 	}
// 	return WeightVectors

// }

// Angle expected between -180 to 180 / in Linear Scale
// returns the gain in dB
func (e Element) HGain(degree float64) float64 {
	if e.Omni {
		return e.GainDb
	}
	degree = Wrap180To180(degree)
	theta := -(degree)

	theta3Db := (e.HBeamWidth)
	SLAV := e.SLAV
	tilt := -(e.HDirection)
	//  Reference TS25.996 - Section 4.5 - BS Antenna Pattern
	val := math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
	return val
}

// VGain Angle expected between 0 to 180 / in Linear Scale
func (e Element) VGain(degree float64) float64 {
	if e.Omni {
		return e.GainDb
	}
	//degree = Wrap0To180(degree)
	if degree > 180 {
		rem := math.Mod(degree, 180.0)
		degree = -180 + rem
	} else if degree < -180 {
		rem := math.Mod(degree, 180.0)
		degree = 180 + rem
	}

	theta := (degree)
	theta3Db := e.VBeamWidth
	SLAV := e.SLAV
	tilt := -e.VDirection
	val := math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
	val = 1
	return val

}

func (e Element) ElementDirectionGain(theta float64) float64 {

	theta3Db := 65.0 * math.Pi / 180.0
	SLAV := 20.0
	tilt := -e.VDirection * math.Pi / 180.0

	return math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
}

// ElementGain generates the antenna gain for given theta,phi in degree
// based on Table 8-6 in Report ITU-R M.2412
// returns effective Antenna Gain Ag, Horizontal gain az, Elevation Gain el
func (e Element) ElementGain(thetaH, thetaV float64) (az, el, Ag float64) {
	phi := Wrap0To180(thetaV)
	theta := Wrap180To180(thetaH)
	theta3dB := 65.0 // degree
	SLAmax := 30.0
	Am := SLAmax
	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)

	MechTiltGCS := e.VDirection // Pointing to Horizon..axis..
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-math.Floor(Av+Ah), Am)
	//result = Ah
	az = Ah
	el = Av
	Ag = result + e.GainDb
	return az, el, Ag
}

func (a AAS) ElementGain(thetaH, thetaV float64) (az, el, Ag float64) {
	az, el, Ag = a.elem.ElementGain(thetaH, thetaV)
	return az, el, Ag
}

func (a AAS) GeneratePattern() {
	fid, _ := os.Create("./results/Antenna_Gain.dat")
	fmt.Fprintf(fid, "%%Az\t\t\t\tEl\t\t\t\tAh\t\t\t\tAv\t\t\t\tAg\t\t\t\tAa")

	Nh, Nv := a.SizeTxRU.R, a.SizeTxRU.C

	// MaxGaindBi := 8.0
	// theta3dB := 65.0 // degree
	// SLAmax := 30.0
	// Am := SLAmax
	// MechTiltGCS := 90.0 // Pointing to Horizon..axis..
	hspace := 0.5
	vspace := 0.8
	dtilt := 157.449 // degree      //can be 5pi/8 ,7pi/8 for urban
	descan := 56.207 //degree       //can be -5pi/16, -3pi/16, -pi/16, 5pi/16, 3pi/16, pi/16 for urban
	temp := -180.0
	temp1 := 0.0

	var sum = complex(0.0, 0.0)
	var theta [182]float64
	var phi [182]float64
	// var Ah [182]float64
	// var Av [182]float64
	// var Ag [182]float64
	// var Aa [182]float64
	var result [182]float64
	for i := 1; i <= 181; i++ {
		theta[i] = temp
		temp = temp + 2

	}

	for i := 1; i <= 181; i++ {
		phi[i] = temp1
		temp1 = temp1 + 1
	}

	// Create weight vectors for a TXRU
	{
		i := 0
		// var w vlib.VectorC
		w := mat.NewCDense(Nh, Nv, nil)
		v := mat.NewCDense(Nh, Nv, nil)
		 vlib.
		
		for m := 1; m < Nh; m++ {
			for n := 1; n < Nv; n++ {
				tw := complex(1/math.Pow(float64(Nh*Nv), 1/2), 0) * cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Sin(dtilt*math.Pi/180)-float64(m-1)*hspace*math.Cos(dtilt*math.Pi/180)*math.Sin(descan*math.Pi/180))))
				tv := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Cos(phi[i]*math.Pi/180)+float64(m-1)*hspace*math.Sin(phi[i]*math.Pi/180)*math.Sin(theta[i]*math.Pi/180))))

				w.Set(m, n, tw)
				v.Set(m, n, tv)
				sum = sum + tw*tv
			}
		}
		fmt.Println(w)
		fmt.Println(v)

	}
	return

	// for i := 1; i <= 181; i++ {
	// 	Ah[i] = -math.Min(12.0*math.Pow(theta[i]/theta3dB, 2.0), Am)
	// 	Av[i] = -math.Min(12.0*math.Pow((phi[i]-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	// 	result[i] = -math.Min(-(Av[i] + Ah[i]), Am)
	// 	Ag[i] = result[i] + MaxGaindBi

	// 	var sum = complex(0.0, 0.0)

	// 	// Generate the weight w and v
	// 	for m := 1; m < Nh; m++ {
	// 		for n := 1; n < Nv; n++ {
	// 			w := complex(1/math.Pow(float64(Nh*Nv), 1/2), 0) * cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Sin(dtilt*math.Pi/180)-float64(m-1)*hspace*math.Cos(dtilt*math.Pi/180)*math.Sin(descan*math.Pi/180))))
	// 			v := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Cos(phi[i]*math.Pi/180)+float64(m-1)*hspace*math.Sin(phi[i]*math.Pi/180)*math.Sin(theta[i]*math.Pi/180))))
	// 			sum = sum + w*v

	// 		}

	// 	}

	// 	Aa[i] = Ag[i] + 10*math.Log10(math.Pow(cmplx.Abs(sum), 2))
	// 	fmt.Fprintf(fid, "\n  %f \t %f \t %f \t %f \t %f\t %f ", theta[i], phi[i], Ah[i], Av[i], Ag[i], Aa[i])
	// }
	_ = result
	fid.Close()
}

//
func GenerateWeightsW(dv float64,elecTilt float64,M int ) vlib.VectorC {
 w:= vlib.NewVectorC(M)
 wm:=zvec.MakeSlice(M)
 elecTiltRad:=vlib.ToRadian(elecTilt)
 KK:=1.0/math.Sqrt(M)
 for m := 0; m < M; m++ {
	 vv:=complex(-2*math.PI)
		cmplx.Exp(2.0*(math.Phi/lamda)*(m-1)*dv*math.Cos(elecTiltRad) )
	 wm[m]=math.Exp(vv)
 }


}