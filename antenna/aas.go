// Implements a simple phase delay from different n-Antenna elements
package antenna

import (
	"encoding/json"
	"io"
	"log"

	// "flag"
	"fmt"
	"github.com/wiless/vlib"
	"math"
	"math/cmplx"
)

var Nodes = 360
var Radius float64 = 1

var imagval complex128 = 0 + 1i

var freq float64 = 2.0e9
var cspeed float64 = 3.0e8
var mfileName string

// var omni bool

var Bobwriter io.Writer

// var VBeamWidth, HBeamWidth float64 = 65, 65

type SettingAAS struct {
	elementLocations                 []vlib.Location3D
	lamda                            float64
	Freq                             float64
	hColumns                         float64 // To be exported later
	N                                int
	Nodes                            int
	Omni                             bool
	MfileName                        string
	VTiltAngle                       float64
	HTiltAngle                       float64
	BeamTilt                         float64
	DisableBeamTit                   bool
	HoldOn                           bool
	VBeamWidth, HBeamWidth           float64
	SLAV                             float64
	ESpacingVFactor, ESpacingHFactor float64
	Centre                           vlib.Location3D `json:-`
	weightVector                     vlib.VectorC
}

func (s *SettingAAS) SetDefault() {
	s.Freq = 2.0e9
	s.N = 1
	s.Nodes = 360
	s.Omni = false
	s.MfileName = "output.m"
	s.VTiltAngle = 14
	s.HTiltAngle = 0
	s.HBeamWidth = 65
	s.VBeamWidth = 65
	s.SLAV = 30.0
	s.lamda = cspeed / freq
	s.ESpacingHFactor = 0 /// Factor mulplied by params.lamda
	s.ESpacingVFactor = .5
}

func NewAAS() *SettingAAS {
	result := new(SettingAAS)
	result.SetDefault()
	return result
}
func (s *SettingAAS) Set(str string) {
	err := json.Unmarshal([]byte(str), s)
	if err != nil {
		log.Print("Error ", err)
	}
}
func init() {

	// flag.Float64Var(&freq, "fc", 2.0e9, "Carrier Frequency in Hz (not GHz)")
	// flag.IntVar(&N, "N", 1, "No. of Element array")
	// flag.IntVar(&Nodes, "Nodes", 360, "No. of Samples in 2*pi ")
	// flag.BoolVar(&omni, "omni", false, "Default uses 3GPP Element pattern TR37.840, Set true for ideal omnidirectional")
	// flag.StringVar(&mfileName, "mfile", "output.m", "Name of .m file to be created")
}

// func (s SettingAAS) SayHello() {
// 	fmt.Printf("\n ========== Simulator SAYS HELLO %v\n   ==========", s)
// }

func RunJSON(jstring string) {
	var s SettingAAS
	err := json.Unmarshal([]byte(jstring), &s)
	if err != nil {
		log.Print("Error ", err)
		return

	}
	fmt.Printf("Wowf %v is this %v", jstring, s)
	// RunAAS(s)
}

func (params *SettingAAS) CreateElements(centre vlib.Location3D) {
	if params.N == 0 {
		return
	}
	params.Centre = centre
	params.lamda = cspeed / freq
	dv := params.ESpacingVFactor * params.lamda
	// dh := params.ESpacingHFactor * params.lamda
	params.elementLocations = make([]vlib.Location3D, params.N)
	// = dropLinearNodes(params.N, dv, 0)
	rotateTilt := GetEJtheta(params.VTiltAngle) // cmplx.Exp(complex(0, -(params.VTiltAngle)*math.Pi/180.0))
	for i := 0; i < params.N; i++ {
		params.elementLocations[i].X = centre.X
		params.elementLocations[i].Y = centre.Y
		params.elementLocations[i].Z = centre.Z + dv*float64(i) - float64(params.N-1)*dv/2.0

		rotatedpos := params.elementLocations[i].Cmplx() * rotateTilt
		params.elementLocations[i].FromCmplx(rotatedpos)
	}
	// fmt.Printf("\n AAS Elem %d locations :  %v", params.N, params.elementLocations)
	params.weightVector = params.FindWeights(params.BeamTilt)
}

func (params *SettingAAS) AASGain(dest vlib.Location3D) (gain float64, effectiveThetaH, effectiveThetaV float64) {
	// src := params.MyLocation()

	params.lamda = cspeed / params.Freq
	AntennaElementLocations := vlib.ToVectorC(params.elementLocations)

	w := vlib.NewOnesC(params.N)
	w = w.Scale(1.0 / float64(params.N))
	// fmt.Print(AntennaElementLocations)
	phaseDelay := vlib.NewVectorF(AntennaElementLocations.Size())
	var Rxcomponent complex128
	Rxcomponent = 0.0

	// fmt.Printf("\n Weights : %v", w)
	for i := 0; i < params.N; i++ {
		dist, thetaH, thetaV := vlib.RelativeGeo(params.elementLocations[i], dest)
		// dist= cmplx.Abs(params.elementLocations[i].Cmplx()-dest.Cmplx())
		aGain := complex((params.ElementEffectiveGain(thetaH, thetaV)), 0)
		_, phaseDelay[i] = math.Modf(vlib.ToDegree(dist / params.lamda))
		Rxcomponent += GetEJtheta(phaseDelay[i]) * w[i] * aGain
	}
	gain = math.Pow(cmplx.Abs(Rxcomponent), 2)
	_, effectiveThetaH, effectiveThetaV = vlib.RelativeGeo(params.Centre, dest)
	return gain, effectiveThetaH, effectiveThetaV

}

func RunAAS(params SettingAAS) {
	// fmt.Printf("\n AAS Parameters : \n %#v \n====", params)
	freq = params.Freq
	N := params.N
	Nodes = params.Nodes
	mfileName = params.MfileName
	// omni := params.Omni
	// flag.Parse()
	// TiltAngle = params.VTiltAngle

	params.lamda = cspeed / freq
	AntennaElementLocations := dropLinearNodes(N, params.lamda/2.0, 0)
	// WeightVector := vlib.NewVectorF(N)
	// for i := 0; i < N; i++ {
	// 	WeightVector[i] = rand.Float64() * 2 * math.Pi

	// 	// WeightVector[i] = 1.0 / math.Sqrt(float64(N))

	// }
	AntennaElementLocations = AntennaElementLocations.AddC(-vlib.MeanC(AntennaElementLocations))
	rotateTilt := cmplx.Exp(complex(0, -(params.VTiltAngle)*math.Pi/180.0))
	AntennaElementLocations = AntennaElementLocations.ScaleC(rotateTilt)

	WeightVector := params.FindWeights(params.BeamTilt)
	if params.DisableBeamTit {
		WeightVector = vlib.NewOnesC(AntennaElementLocations.Size())
	}
	WeightVector = WeightVector.Scale(math.Sqrt(1.0 / float64(N)))
	fmt.Printf("\nWeights  = %f", WeightVector)

	NodeLocations := dropCircularNodes(Nodes, Radius)
	meanvalue := vlib.MeanC(AntennaElementLocations)
	NodeLocations = NodeLocations.AddC(meanvalue)
	fmt.Println("Mid = ", meanvalue)
	// fmt.Printf("\nNodeLocations = %f ", NodeLocations)

	/// Evaluate

	Gains := vlib.NewVectorF(NodeLocations.Size())
	for nindx, pos := range NodeLocations {

		var gain complex128 //:= vlib.NewVectorF(AntennaElementLocations.Size())
		// gain = 1.0
		phaseDelay := vlib.NewVectorF(AntennaElementLocations.Size())
		for eindx, epos := range AntennaElementLocations {
			dist := cmplx.Abs(epos - pos)
			_, phaseDelay[eindx] = math.Modf(dist / params.lamda)
			phaseDelay[eindx] *= (2.0 * math.Pi) //+ WeightVector[eindx]
			jtheta := complex(0.0, phaseDelay[eindx])
			directionGain := math.Sqrt(params.ElementDirectionGain(cmplx.Phase(pos - epos)))
			gain += complex(directionGain, 0) * cmplx.Exp(-jtheta) * WeightVector[eindx]
		}
		// fmt.Printf("\n Phase[%d]=%v", nindx, phaseDelay)
		Gains[nindx] = cmplx.Abs(gain) * cmplx.Abs(gain) / Radius

		// fmt.Println("%Result : ", nindx, phaseDelay)

	}
	// locs := NodeLocations.Scale(Radius)

	// matlab := vlib.NewMatlab(mfileName)
	var matlab vlib.Matlab
	matlab.SetDefaults()

	if Bobwriter != nil {
		matlab.SetWriter(Bobwriter)
	} else {
		matlab.SetFile(params.MfileName)
	}
	matlab.Silent = true
	matlab.Export("Weights", WeightVector)
	matlab.Export("AntennaLocations", AntennaElementLocations)
	matlab.Export("Locations", NodeLocations)
	matlab.Export("Gain", Gains)
	matlab.Export("N", N)
	matlab.Command("\npattern=(Locations.*sqrt(Gain ));")
	if !params.HoldOn {
		matlab.Command("figure;")
		// matlab.Command("plot(real(pattern),imag(pattern ),'k-')")
		// matlab.Command("hold on")
		// matlab.Command("axis([-15 +15 -15 +15]);")

	}
	matlab.Command("plot(real(AntennaLocations ),imag(AntennaLocations ),'r*');")
	matlab.Command("grid on;")
	if !params.HoldOn {

		matlab.Command("figure;")
	}
	matlab.Command("polar(angle(pattern),abs(pattern ),'k-')")

	matlab.Close()

}

func GetEJtheta(degree float64) complex128 {
	return cmplx.Exp(complex(0.0, -degree*math.Pi/180.0))
}

func Radian(degree float64) float64 {
	return degree * math.Pi / 180.0
}

func (params *SettingAAS) oldFindWeights(theta float64) vlib.VectorC {
	WeightVectors := vlib.NewVectorC(params.N)
	// var gain complex128
	AE := vlib.ToVectorC(params.elementLocations)
	meanpos := vlib.MeanC(AE)
	pos := GetEJtheta(theta) + meanpos
	// gain := complex(1.0/math.Sqrt(float64(N)), 0)
	phaseDelay := vlib.NewVectorF(AE.Size())
	for eindx, epos := range AE {
		dist := cmplx.Abs(epos - pos)
		_, phaseDelay[eindx] = math.Modf(dist / params.lamda)
		phaseDelay[eindx] *= (2.0 * math.Pi)
		WeightVectors[eindx] = cmplx.Exp(complex(0.0, -phaseDelay[eindx]))

	}
	return WeightVectors
}
func (params *SettingAAS) FindWeights(theta float64) vlib.VectorC {
	// for nindx, pos := range NodeLocations {

	WeightVectors := vlib.NewVectorC(params.N)

	for i := 0; i < params.N; i++ {
		m := float64(i)
		arg := complex(0, 2*math.Pi*(m-1)*params.lamda/2.0*math.Cos(Radian(theta+90))/params.lamda)
		WeightVectors[i] = cmplx.Exp(arg)
	}
	return WeightVectors

}

func (s SettingAAS) ElementDirectionHGain(degree float64) float64 {
	if s.Omni {
		return 1.0
	}

	// fmt.Println("Origina ", degree)
	if degree > 180 {
		rem := math.Mod(degree, 180.0)
		degree = -180 + rem

	} else if degree < -180 {
		rem := math.Mod(degree, 180.0)
		//	fmt.Println("Remainder for ", degree, rem)
		degree = 180 + rem
	}
	theta := -(degree)
	theta3Db := (s.HBeamWidth)
	SLAV := s.SLAV
	tilt := -(s.HTiltAngle)
	val := math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
	return val
}

func (s SettingAAS) ElementDirectionVGain(degree float64) float64 {
	if s.Omni {
		return 1.0
	}

	if degree > 180 {
		rem := math.Mod(degree, 180.0)
		degree = -180 + rem
	} else if degree < -180 {
		rem := math.Mod(degree, 180.0)
		degree = 180 + rem
	}

	theta := (degree)
	theta3Db := s.VBeamWidth
	SLAV := s.SLAV
	tilt := -s.VTiltAngle
	val := math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
	return val

}

func (s SettingAAS) ElementEffectiveGain(thetaH, thetaV float64) float64 {
	hgain, vgain := s.ElementDirectionHGain(thetaH), s.ElementDirectionVGain(thetaV)

	sumgain := hgain * vgain
	// sumgain = 1.0 / sumgain
	// sumgain = 1 / (sumgain ^ 2)
	// ZZ(x, y) = 1 / min(sumgain, 1000)
	return math.Max(sumgain, vlib.InvDb(-s.SLAV))
}

func (s SettingAAS) ElementDirectionGain(theta float64) float64 {
	if s.Omni {
		return 1.0
	}

	theta3Db := 65.0 * math.Pi / 180.0
	SLAV := 20.0
	tilt := -s.VTiltAngle * math.Pi / 180.0
	// fmt.Printf("\n %% %f dump %v", theta, math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV))
	return math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)

	// if theta >= -math.Pi/2 && theta <= math.Pi/2 {
	// 	return 1
	// 	// math.Abs(math.Cos(theta))
	// } else {
	// 	return 0.2 * math.Abs(math.Cos(theta))
	// }

	// return math.Abs(math.Cos(theta))
	// return math.Abs(real(cmplx.Exp(-complex(0.0, theta)))
}

/// Draws nNodes in a circular fashion centered around 0,0
func dropCircularNodes(N int, radius float64) vlib.VectorC {
	result := vlib.NewVectorC(N)
	delTheta := 2 * math.Pi / float64(N)
	angle := 0.0
	for i := 0; i < N; i++ {
		angle += delTheta
		jtheta := complex(0.0, angle)
		result[i] = complex(radius, 0) * cmplx.Exp(-jtheta)
	}
	return result
}

/// Drops Linear Vertical Nodes spaced with dh,dv linearly
func dropLinearNodes(N int, dv, dh float64) vlib.VectorC {
	result := vlib.NewVectorC(N)
	var xloc, yloc float64
	// dv := 10.0
	// dh:=0.0
	for i := 0; i < N; i++ {
		result[i] = complex(xloc, yloc)
		yloc += dv
		xloc += dh

	}

	return result
}
